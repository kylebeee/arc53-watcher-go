package compound

import (
	"fmt"

	"github.com/kylebeee/arc53-watcher-go/db"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type Collection struct {
	*db.Collection
	BannerURL      *string           `json:"banner_url,omitempty"`
	BannerMime     *string           `json:"banner_mime,omitempty"`
	AvatarURL      *string           `json:"avatar_url,omitempty"`
	AvatarMime     *string           `json:"avatar_mime,omitempty"`
	Prefixes       []string          `json:"prefixes,omitempty"`
	Addresses      []string          `json:"addresses,omitempty"`
	Assets         []uint64          `json:"assets,omitempty"`
	ExcludedAssets []uint64          `json:"excluded_assets,omitempty"`
	Artists        []string          `json:"artists,omitempty"`
	Properties     []Property        `json:"properties,omitempty"`
	Extras         map[string]string `json:"extras,omitempty"`
}

type CollectionGetExclude string

const (
	CollectionGetExcludePrefixes       CollectionGetExclude = "prefixes"
	CollectionGetExcludeAddresses      CollectionGetExclude = "addresses"
	CollectionGetExcludeAssets         CollectionGetExclude = "assets"
	CollectionGetExcludeExcludedAssets CollectionGetExclude = "excluded_assets"
	CollectionGetExcludeArtists        CollectionGetExclude = "artists"
	CollectionGetExcludeProperties     CollectionGetExclude = "properties"
	CollectionGetExcludeExtras         CollectionGetExclude = "extras"
)

var allCollectionGetExcludes = []CollectionGetExclude{
	CollectionGetExcludePrefixes,
	CollectionGetExcludeAddresses,
	CollectionGetExcludeAssets,
	CollectionGetExcludeExcludedAssets,
	CollectionGetExcludeArtists,
	CollectionGetExcludeProperties,
	CollectionGetExcludeExtras,
}

func GetCollectionsByProviderID[H db.Handle](h H, providerID uint64, exclude ...CollectionGetExclude) (*[]Collection, error) {
	const op errors.Op = "GetCollectionsByProviderID"

	var collections []Collection
	cols, err := db.GetCollectionsByProviderID(h, providerID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	for i := range *cols {
		col := (*cols)[i]
		compCollection := Collection{
			Collection: &col,
		}
		buffer := (len(allCollectionGetExcludes) - len(exclude))
		if buffer == 0 {
			continue
		}

		rChan := make(chan interface{}, buffer)
		defer close(rChan)

		if !misc.InSlice(CollectionGetExcludePrefixes, exclude) {
			go func() {
				prefixes, err := db.GetCollectionPrefixes(h, col.ID)
				if err != nil && !db.ErrNoRows(err) {
					rChan <- err
					return
				}
				rChan <- prefixes
			}()
		}

		if !misc.InSlice(CollectionGetExcludeAddresses, exclude) {
			go func() {
				addresses, err := db.GetCollectionAddresses(h, col.ID)
				if err != nil && !db.ErrNoRows(err) {
					rChan <- err
					return
				}
				rChan <- addresses
			}()
		}

		if !misc.InSlice(CollectionGetExcludeAssets, exclude) {
			go func() {
				assets, err := db.GetCollectionAssets(h, col.ID)
				if err != nil && !db.ErrNoRows(err) {
					rChan <- err
					return
				}
				rChan <- assets
			}()
		}

		if !misc.InSlice(CollectionGetExcludeExcludedAssets, exclude) {
			go func() {
				excludedAssets, err := db.GetCollectionExcludedAssets(h, col.ID)
				if err != nil && !db.ErrNoRows(err) {
					rChan <- err
					return
				}
				rChan <- excludedAssets
			}()
		}

		if !misc.InSlice(CollectionGetExcludeArtists, exclude) {
			go func() {
				artists, err := db.GetCollectionArtistByCollection(h, col.ID)
				if err != nil && !db.ErrNoRows(err) {
					rChan <- err
					return
				}
				rChan <- artists
			}()
		}

		if !misc.InSlice(CollectionGetExcludeProperties, exclude) {
			go func() {
				properties, err := GetProperties(h, col.ID)
				if err != nil && !db.ErrNoRows(err) {
					rChan <- err
					return
				}
				rChan <- properties
			}()
		}

		if !misc.InSlice(CollectionGetExcludeExtras, exclude) {
			go func() {
				extras, err := db.GetCollectionExtras(h, col.ID)
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
			case *[]db.CollectionPrefix:
				prefixes := []string{}
				for i := range *result {
					prefix := (*result)[i]
					prefixes = append(prefixes, prefix.Prefix)
				}
				compCollection.Prefixes = prefixes
			case *[]db.CollectionAddress:
				addresses := []string{}
				for i := range *result {
					address := (*result)[i]
					addresses = append(addresses, address.Address)
				}
				compCollection.Addresses = addresses
			case *[]db.CollectionAsset:
				assets := []uint64{}
				for i := range *result {
					asset := (*result)[i]
					assets = append(assets, asset.AsaID)
				}
				compCollection.Assets = assets
			case *[]db.CollectionExcludedAsset:
				assets := []uint64{}
				for i := range *result {
					asset := (*result)[i]
					assets = append(assets, asset.AsaID)
				}
				compCollection.ExcludedAssets = assets
			case *[]db.CollectionArtist:
				artists := []string{}
				for i := range *result {
					artist := (*result)[i]
					artists = append(artists, artist.Address)
				}
				compCollection.Artists = artists
			case *[]Property:
				compCollection.Properties = *result
			case *[]db.CollectionExtras:
				extras := map[string]string{}
				for i := range *result {
					extra := (*result)[i]
					extras[extra.Key] = extra.Value
				}

				compCollection.Extras = extras
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
				msg += err.Error() + "\n"
			}

			return nil, errors.E(op, fmt.Errorf(msg))
		}
		collections = append(collections, compCollection)
	}

	return &collections, nil
}

func GetCollectionCriteria[H db.Handle](h H, collectionID string) (*Collection, []string, error) {
	const op errors.Op = "GetCollectionCriteria"
	collection := Collection{}

	buffer := 5
	rChan := make(chan interface{}, buffer)
	defer close(rChan)

	go func() {
		wallets, err := db.GetCollectionCreatorWallets(h, collectionID)
		if err != nil && !db.ErrNoRows(err) {
			rChan <- err
			return
		}
		rChan <- wallets
	}()

	go func() {
		prefixes, err := db.GetCollectionPrefixes(h, collectionID)
		if err != nil && !db.ErrNoRows(err) {
			rChan <- err
			return
		}
		rChan <- prefixes
	}()

	go func() {
		addresses, err := db.GetCollectionAddresses(h, collectionID)
		if err != nil && !db.ErrNoRows(err) {
			rChan <- err
			return
		}

		rChan <- addresses
	}()

	go func() {
		assets, err := db.GetCollectionAssets(h, collectionID)
		if err != nil && !db.ErrNoRows(err) {
			rChan <- err
			return
		}
		rChan <- assets
	}()

	go func() {
		excludedAssets, err := db.GetCollectionExcludedAssets(h, collectionID)
		if err != nil && !db.ErrNoRows(err) {
			rChan <- err
			return
		}
		rChan <- excludedAssets
	}()

	creators := []string{}
	var errs []error
	for i := 0; i < buffer; i++ {
		data := <-rChan
		switch result := data.(type) {
		case *[]db.ProviderAddress:
			for i := range *result {
				address := (*result)[i]
				creators = append(creators, address.Address)
			}
		case *[]db.CollectionPrefix:
			prefixes := []string{}
			for i := range *result {
				prefix := (*result)[i]
				prefixes = append(prefixes, prefix.Prefix)
			}
			collection.Prefixes = prefixes
		case *[]db.CollectionAddress:
			addresses := []string{}
			for i := range *result {
				address := (*result)[i]
				addresses = append(addresses, address.Address)
			}
			collection.Addresses = addresses
		case *[]db.CollectionAsset:
			assets := []uint64{}
			for i := range *result {
				asset := (*result)[i]
				assets = append(assets, asset.AsaID)
			}
			collection.Assets = assets
		case *[]db.CollectionExcludedAsset:
			assets := []uint64{}
			for i := range *result {
				asset := (*result)[i]
				assets = append(assets, asset.AsaID)
			}
			collection.ExcludedAssets = assets
		case error:
			errs = append(errs, result)
		}
	}

	if len(errs) > 0 {
		msg := ""
		for _, err := range errs {
			msg += err.Error() + "\n"
		}

		return nil, []string{}, errors.E(op, fmt.Errorf(msg))
	}

	return &collection, creators, nil
}
