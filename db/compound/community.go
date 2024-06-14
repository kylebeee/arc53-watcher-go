package compound

import (
	"fmt"
	"strings"

	"github.com/kylebeee/arc53-watcher-go/db"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type Community struct {
	*db.Community
	Settings    *db.CommunitySettings   `json:"settings,omitempty"`
	Tokens      []db.CommunityToken     `json:"tokens,omitempty"`
	Associates  []db.CommunityAssociate `json:"associates,omitempty"`
	Collections []Collection            `json:"collections,omitempty"`
	Faq         []db.CommunityFaq       `json:"faq,omitempty"`
	Extras      []db.CommunityExtras    `json:"extras,omitempty"`
}

type CommunityGetExclude string

const (
	CommunityGetExcludeSettings    CommunityGetExclude = "settings"
	CommunityGetExcludeTokens      CommunityGetExclude = "tokens"
	CommunityGetExcludeAssociates  CommunityGetExclude = "associates"
	CommunityGetExcludeCollections CommunityGetExclude = "collections"
	CommunityGetExcludeFaq         CommunityGetExclude = "faq"
	CommunityGetExcludeExtras      CommunityGetExclude = "extras"
)

func GetCommunity[H db.Handle](h H, providerID uint64, exclude ...CommunityGetExclude) (*Community, error) {
	const op errors.Op = "GetCommunity"
	var community Community
	buffer := (7 - len(exclude))
	rChan := make(chan interface{}, buffer)
	defer close(rChan)

	go func() {
		comm, err := db.GetCommunity(h, providerID)
		if err != nil {
			rChan <- err
			return
		}
		rChan <- comm
	}()

	if !misc.InSlice(CommunityGetExcludeSettings, exclude) {
		go func() {
			settings, err := db.GetCommunitySettings(h, providerID)
			if err != nil && !db.ErrNoRows(err) {
				rChan <- &db.CommunitySettings{
					ID:         providerID,
					DefaultTab: db.DefaultCommunityTab,
				}
				return
			}
			rChan <- settings
		}()
	}

	if !misc.InSlice(CommunityGetExcludeTokens, exclude) {
		go func() {
			tokens, err := db.GetCommunityTokens(h, providerID)
			if err != nil && !db.ErrNoRows(err) {
				rChan <- err
				return
			}
			rChan <- tokens
		}()
	}

	if !misc.InSlice(CommunityGetExcludeAssociates, exclude) {
		go func() {
			associates, err := db.GetCommunityAssociates(h, providerID)
			if err != nil && !db.ErrNoRows(err) {
				rChan <- err
				return
			}
			rChan <- associates
		}()
	}

	if !misc.InSlice(CommunityGetExcludeCollections, exclude) {
		go func() {
			collections, err := GetCollectionsByProviderID(h, providerID)
			if err != nil && !db.ErrNoRows(err) {
				rChan <- err
				return
			}
			rChan <- collections
		}()
	}

	if !misc.InSlice(CommunityGetExcludeFaq, exclude) {
		go func() {
			faq, err := db.GetCommunityFaq(h, providerID, 0, 10)
			if err != nil {
				rChan <- err
				return
			}
			rChan <- faq
		}()
	}

	if !misc.InSlice(CommunityGetExcludeExtras, exclude) {
		go func() {
			extras, err := db.GetCommunityExtras(h, providerID)
			if err != nil && !db.ErrNoRows(err) {
				rChan <- err
				return
			}
			rChan <- extras
		}()
	}

	var errs []error
	for i := 0; i < buffer; i++ {
		data := <-rChan
		switch result := data.(type) {
		case *db.Community:
			community.Community = result
		case *db.CommunitySettings:
			community.Settings = result
		case *[]db.CommunityToken:
			community.Tokens = *result
		case *[]db.CommunityAssociate:
			community.Associates = *result
		case *[]Collection:
			community.Collections = *result
		case *[]db.CommunityFaq:
			community.Faq = *result
		case *[]db.CommunityExtras:
			community.Extras = *result
		case error:
			errs = append(errs, result)
		default:
			fmt.Println(op, "defaulted")
			fmt.Println(result)
		}
	}

	if len(errs) > 0 {
		msg := ""
		for _, err := range errs {
			if db.ErrNoRows(err) && strings.Contains(err.Error(), "operation: GetCommunity") {
				return nil, err
			}

			msg += err.Error() + "\n"
		}

		return nil, errors.E(op, fmt.Errorf(msg))
	}

	return &community, nil
}

func DeleteCommunity[H db.Handle](h H, providerID uint64) error {
	const op errors.Op = "DeleteCommunity"
	var err error

	err = db.DeleteCommunity(h, providerID)
	if err != nil {
		return errors.E(op, err)
	}

	// community_json
	err = db.DeleteCommunityJson(h, providerID)
	if err != nil {
		return errors.E(op, err)
	}

	// community_tokens
	err = db.DeleteCommunityTokens(h, providerID)
	if err != nil {
		return errors.E(op, err)
	}

	// community_faq
	err = db.DeleteCommunityFaq(h, providerID)
	if err != nil {
		return errors.E(op, err)
	}

	// community_extras
	err = db.DeleteCommunityExtras(h, providerID)
	if err != nil {
		return errors.E(op, err)
	}

	// collection
	collections, err := db.GetCollectionsByProviderID(h, providerID)
	if err != nil {
		return errors.E(op, err)
	}

	for _, collection := range *collections {

		err = db.DeleteCollectionPrefixes(h, collection.ID)
		if err != nil {
			return errors.E(op, err)
		}

		err = db.DeleteCollectionArtists(h, collection.ID)
		if err != nil {
			return errors.E(op, err)
		}

		err = db.DeleteCollectionAssets(h, collection.ID)
		if err != nil {
			return errors.E(op, err)
		}

		err = db.DeleteCollectionExcludedAssets(h, collection.ID)
		if err != nil {
			return errors.E(op, err)
		}

		err = db.DeleteCollectionExtras(h, collection.ID)
		if err != nil {
			return errors.E(op, err)
		}

		properties, err := db.GetProperties(h, collection.ID)
		if err != nil {
			return errors.E(op, err)
		}

		for i := range *properties {
			property := (*properties)[i]

			err = db.DeletePropertyValues(h, property.ID)
			if err != nil {
				return errors.E(op, err)
			}

			err = db.DeletePropertyValueExtras(h, property.ID)
			if err != nil {
				return errors.E(op, err)
			}
		}

		err = db.DeleteCollectionProperties(h, collection.ID)
		if err != nil {
			return errors.E(op, err)
		}
	}

	err = db.DeleteCollectionsByProviderID(h, providerID)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
