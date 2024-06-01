package nfd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/db"
	"github.com/kylebeee/arc53-watcher-go/db/compound"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
	"github.com/kylebeee/arc53-watcher-go/uuid"
)

const NFDMainNetRegistryAppID uint64 = 760937186

type NFDProvider struct {
	DB      *sqlx.DB
	Algod   *algod.Client
	SyncMap *sync.Map
}

func (p *NFDProvider) Init(dbConn *sqlx.DB, algodClient *algod.Client) error {
	const op errors.Op = "NFDProvider.Init"

	p.DB = dbConn
	p.Algod = algodClient

	appIDs, err := db.GetAllProvidersByType(p.DB, "nfd")
	if err != nil {
		return errors.E(op, err)
	}

	for i := range *appIDs {
		appID := (*appIDs)[i]
		p.SyncMap.Store(appID, struct{}{})
	}

	return nil
}

func (p *NFDProvider) Process(stxn types.SignedTxnInBlock, round uint64) error {

	txnsToProcess := append([]types.SignedTxnWithAD{stxn.SignedTxnWithAD}, misc.ListInner(&stxn.SignedTxnWithAD)...)

	for i := range txnsToProcess {
		txn := txnsToProcess[i].Txn
		txAppID := uint64(txn.ApplicationFields.ApplicationID)

		_, exists := p.SyncMap.Load(txAppID)
		if exists {
			// update on existing NFD
			err := p.SyncNFDByAppID(uint64(txn.ApplicationFields.ApplicationID), round)
			if err != nil {
				return err
			}
			return nil
		}

		isMint := txn.Type == types.ApplicationCallTx && txAppID == NFDMainNetRegistryAppID && len(txn.ApplicationFields.ApplicationArgs) > 0 && string(txn.ApplicationFields.ApplicationArgs[0]) == "mint"
		if isMint {

			var appID uint64
			for _, innerTxn := range stxn.EvalDelta.InnerTxns {
				if innerTxn.Txn.Type == types.ApplicationCallTx {
					appID = uint64(innerTxn.Txn.ApplicationID)
					break
				}
			}

			// new NFD
			p.SyncMap.Store(appID, struct{}{})
			err := p.SyncNFDByAppID(appID, round)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// helpers
func (p *NFDProvider) SyncNFDByAppID(appID uint64, currentBlock uint64) error {
	const op errors.Op = "SyncNFDByAppID"
	var new bool = false
	var communitySet bool = false

	fmt.Println("Syncing NFD: ", appID)

	_, err := db.GetProvider(p.DB, appID)
	if err != nil && !db.ErrNoRows(err) {
		return errors.E(op, err)
	} else if db.ErrNoRows(err) {
		new = true
	}

	fmt.Println("Getting NFD data")

	nfdProperties, err := GetNFDData(p.Algod, context.Background(), appID)
	if err != nil {
		return errors.E(op, err)
	}

	fmt.Println(nfdProperties)

	tx, err := p.DB.Beginx()
	if err != nil {
		return errors.E(op, err)
	}

	// jsonDataIndent, err := json.MarshalIndent(nfdProperties, "", "  ")
	// if err != nil {
	// 	tx.Rollback()
	// 	return errors.E(op, err)
	// }
	// fmt.Println(string(jsonDataIndent))

	dniAddresses := []string{}
	addresses := map[string]db.ProviderAddress{}

	if !new {
		preexistingAddresses, err := db.GetProviderAddresses(p.DB, appID)
		if err != nil && !db.ErrNoRows(err) {
			tx.Rollback()
			return errors.E(op, err)
		}

		if preexistingAddresses != nil {
			for _, address := range *preexistingAddresses {
				addresses[address.Address] = address
			}
		}
	}

	for key, value := range nfdProperties.UserDefined {
		switch key {
		case "akitacommunity", "project":
			{
				communitySet = true
				err = p.processCommunity(tx, appID, []byte(value))
				if err != nil {
					tx.Rollback()
					return errors.E(op, err)
				}
			}
		}
	}

	for key, value := range nfdProperties.Verified {
		switch key {
		case "caAlgo":
			{
				vaddresses := misc.UniqueSlice(strings.Split(value, ","))

				for _, address := range vaddresses {
					dniAddresses = append(dniAddresses, address)
					nfdWallet := &db.ProviderAddress{ID: appID, Address: address}

					_, walletExists := addresses[address]
					if !walletExists {
						_, err = db.Insert(tx, nfdWallet)
						if err != nil {
							tx.Rollback()
							return errors.E(op, err)
						}
					}
				}
			}
		}
	}

	// delete wallets not in list
	err = db.DeleteProviderAddressNotIn(tx, appID, dniAddresses...)
	if err != nil {
		tx.Rollback()
		return errors.E(op, err)
	}

	if !communitySet {
		_, err = db.GetCommunity(p.DB, appID)
		if err != nil && !db.ErrNoRows(err) {
			tx.Rollback()
			return errors.E(op, err)
		} else if !db.ErrNoRows(err) {
			err = compound.DeleteCommunity(tx, appID)
			if err != nil {
				tx.Rollback()
				return errors.E(op, err)
			}
		}
	}

	// insert or update db with NFD data
	if new {
		_, err = db.Insert(tx, &db.Provider{ID: appID, Type: "nfd"})
		if err != nil {
			tx.Rollback()
			return errors.E(op, err)
		}
	}

	fmt.Printf("Committing changes for %v\n", appID)
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return errors.E(op, err)
	}

	return nil
}

func (p *NFDProvider) processCommunity(tx *sqlx.Tx, nfdID uint64, data []byte) error {
	const op errors.Op = "ProcessCommunity"

	if strings.HasPrefix(string(data), "ipfs://") {
		ipfsdata, err := GetIPFSData(string(data))
		if err != nil {
			return errors.E(op, err)
		}
		data = ipfsdata
	}

	commJson := &db.CommunityJson{
		ID:   nfdID,
		Data: string(data),
	}

	prevJson, err := db.GetCommunityJson(p.DB, nfdID)
	if err != nil && !db.ErrNoRows(err) {
		return errors.E(op, err)
	} else if db.ErrNoRows(err) {
		_, err = db.Insert(tx, commJson)
		if err != nil {
			return errors.E(op, err)
		}
	} else {
		if prevJson.Data == string(data) {
			return nil
		}

		_, err = db.Update(tx, commJson, map[string]interface{}{"id": nfdID})
		if err != nil {
			return errors.E(op, err)
		}
	}

	communityData := compound.Community{}
	err = json.Unmarshal(data, &communityData)
	if err != nil {
		fmt.Println(err)

		commJson.Malformed = misc.PointerBool(true)
		_, err := db.Update(tx, commJson, map[string]interface{}{"id": nfdID})
		if err != nil {
			return errors.E(op, err)
		}

		return nil
	}

	comm := communityData.Community
	comm.ID = nfdID

	_, err = db.GetCommunity(p.DB, nfdID)
	if err != nil && !db.ErrNoRows(err) {
		return errors.E(op, err)
	} else if db.ErrNoRows(err) {
		_, err = db.Insert(tx, comm)
		if err != nil {
			return errors.E(op, err)
		}
	}

	err = p.processTokens(tx, nfdID, communityData.Tokens)
	if err != nil {
		return errors.E(op, err)
	}

	err = p.processAssociates(tx, nfdID, communityData.Associates)
	if err != nil {
		return errors.E(op, err)
	}

	err = p.processCollections(tx, nfdID, communityData.Collections)
	if err != nil {
		return errors.E(op, err)
	}

	err = p.processFaq(tx, nfdID, communityData.Faq)
	if err != nil {
		return errors.E(op, err)
	}

	err = p.processExtras(tx, nfdID, communityData.Extras)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (p *NFDProvider) processTokens(tx *sqlx.Tx, nfdID uint64, tokensData []db.CommunityToken) error {
	const op errors.Op = "ProcessTokens"

	tokenKeys := map[uint64]db.CommunityToken{}
	tokens, err := db.GetCommunityTokens(p.DB, nfdID)
	if err != nil && !db.ErrNoRows(err) {
		return errors.E(op, err)
	}

	if tokens != nil {
		for _, token := range *tokens {
			tokenKeys[token.AssetID] = token
		}
	}

	dniKeys := []uint64{}
	for _, t := range tokensData {
		token := t

		// we dont need to trigger the asset cache processor here because *if*
		// the asset is actually created by this community it will be picked up while
		// the system processes the verified wallets created assets

		_, exists := tokenKeys[token.AssetID]
		dniKeys = append(dniKeys, token.AssetID)
		token.ID = nfdID
		if !exists {
			_, err = db.Insert(tx, &token)
			if err != nil {
				return errors.E(op, err)
			}
		} else {
			_, err = db.Update(tx, &token, map[string]interface{}{"id": nfdID, "asset_id": token.AssetID})
			if err != nil {
				return errors.E(op, err)
			}
		}
	}

	err = db.DeleteCommunityTokensNotIn(tx, nfdID, dniKeys...)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (p *NFDProvider) processAssociates(tx *sqlx.Tx, nfdID uint64, associateData []db.CommunityAssociate) error {
	const op errors.Op = "ProcessTokens"

	associateKeys := map[string]db.CommunityAssociate{}
	associates, err := db.GetCommunityAssociates(p.DB, nfdID)
	if err != nil && !db.ErrNoRows(err) {
		return errors.E(op, err)
	}

	for _, associate := range *associates {
		associateKeys[associate.Address] = associate
	}

	dniKeys := []string{}
	for _, a := range associateData {
		associate := a

		_, exists := associateKeys[associate.Address]
		dniKeys = append(dniKeys, associate.Address)
		if !exists {
			associate.ID = nfdID
			_, err = db.Insert(tx, &associate)
			if err != nil {
				return errors.E(op, err)
			}
		}
	}

	err = db.DeleteCommunityAssociatesNotIn(tx, nfdID, dniKeys...)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (p *NFDProvider) processCollections(tx *sqlx.Tx, nfdID uint64, collectionsData []compound.Collection) error {
	const op errors.Op = "ProcessCollections"

	collectionKeys := map[string]compound.Collection{}
	collections, err := compound.GetCollectionsByNFDID(p.DB, nfdID)
	if err != nil {
		if err.(*errors.Error).Kind != errors.DatabaseResultNotFound {
			return errors.E(op, err)
		}

		// even if we get no results back, the map will be empty
		// and will be caught by the check for pre-existing collections
	}

	for i := range *collections {
		col := (*collections)[i]
		collectionKeys[col.Name] = col
	}

	dniCollection := []string{}
	for _, col := range collectionsData {
		pre, exists := collectionKeys[col.Name]
		if exists {
			dniCollection = append(dniCollection, pre.ID)

			preJson, err := json.Marshal(pre)
			if err != nil {
				return errors.E(op, err)
			}

			colJson, err := json.Marshal(col)
			if err != nil {
				return errors.E(op, err)
			}

			if string(preJson) == string(colJson) {
				continue
			}

			// update
			_, err = db.Update(tx, col.Collection, map[string]interface{}{"id": pre.ID})
			if err != nil {
				return errors.E(op, err)
			}

			// insert or update prefixes
			// dni = Delete not in
			dniPrefixes := []string{}
			for _, prefix := range col.Prefixes {
				dniPrefixes = append(dniPrefixes, prefix)

				if !misc.InSlice(prefix, pre.Prefixes) {
					_, err = db.Insert(tx, &db.CollectionPrefix{ID: pre.ID, Prefix: prefix})
					if err != nil {
						return errors.E(op, err)
					}
				}
			}

			// delete not in prefix list
			err = db.DeleteCollectionPrefixesNotIn(tx, pre.ID, dniPrefixes...)
			if err != nil {
				return errors.E(op, err)
			}

			dniAddresses := []string{}
			for _, address := range col.Addresses {
				dniAddresses = append(dniAddresses, address)

				if !misc.InSlice(address, pre.Addresses) {
					_, err = db.Insert(tx, &db.CollectionAddress{ID: pre.ID, Address: address})
					if err != nil {
						return errors.E(op, err)
					}
				}
			}

			// delete not in address list
			err = db.DeleteCollectionAddressesNotIn(tx, pre.ID, dniAddresses...)
			if err != nil {
				return errors.E(op, err)
			}

			// insert or update assets
			// dni = Delete not in
			dniAssets := []uint64{}
			for _, asset := range col.Assets {
				dniAssets = append(dniAssets, asset)

				if !misc.InSlice(asset, pre.Assets) {
					_, err = db.Insert(tx, &db.CollectionAsset{ID: pre.ID, AsaID: asset})
					if err != nil {
						return errors.E(op, err)
					}
				}
			}

			// delete not in asset list
			err = db.DeleteCollectionAssetsNotIn(tx, pre.ID, dniAssets...)
			if err != nil {
				return errors.E(op, err)
			}

			// insert or update excluded assets
			// dni = Delete not in
			dniExcludedAssets := []uint64{}
			for _, asset := range col.ExcludedAssets {
				dniExcludedAssets = append(dniExcludedAssets, asset)

				if !misc.InSlice(asset, pre.ExcludedAssets) {
					_, err = db.Insert(tx, &db.CollectionExcludedAsset{ID: pre.ID, AsaID: asset})
					if err != nil {
						return errors.E(op, err)
					}
				}
			}

			// delete not in excluded asset list
			err = db.DeleteCollectionExcludedAssetsNotIn(tx, pre.ID, dniExcludedAssets...)
			if err != nil {
				return errors.E(op, err)
			}

			dniArtists := []string{}
			for _, address := range col.Artists {
				dniArtists = append(dniArtists, address)

				if !misc.InSlice(address, pre.Artists) {
					_, err = db.Insert(tx, &db.CollectionArtist{ID: pre.ID, Address: address})
					if err != nil {
						return errors.E(op, err)
					}
				}
			}

			// delete not in artist list
			err = db.DeleteCollectionArtistsNotIn(tx, pre.ID, dniArtists...)
			if err != nil {
				return errors.E(op, err)
			}

			propertyKeys := map[string]compound.Property{}
			for _, property := range pre.Properties {
				propertyKeys[property.Name] = property
			}

			// dni = Delete not in
			dniProperties := []string{}
			for _, prop := range col.Properties {
				preProp, propExists := propertyKeys[prop.Name]
				if propExists {
					prop.ID = preProp.ID
					prop.CollectionID = preProp.CollectionID

					dniProperties = append(dniProperties, prop.ID)

					_, err = db.Update(tx, prop.Property, map[string]interface{}{"id": preProp.ID})
					if err != nil {
						return errors.E(op, err)
					}

					propValueNames := []string{}
					for _, value := range preProp.Values {
						propValueNames = append(propValueNames, value.Name)
					}

					// insert or update properties values
					dniPropertiesValues := []string{}
					for _, value := range prop.Values {
						dniPropertiesValues = append(dniPropertiesValues, value.Name)
						value.ID = preProp.ID

						if misc.InSlice(value.Name, propValueNames) {
							_, err = db.Update(tx, value.PropertyValue, map[string]interface{}{"id": preProp.ID, "name": value.Name})
							if err != nil {
								return errors.E(op, err)
							}
						} else {
							_, err = db.Insert(tx, value.PropertyValue)
							if err != nil {
								return errors.E(op, err)
							}
						}

						propValueExtrasKeys := []string{}
						for _, pvalue := range preProp.Values {
							if pvalue.Name == value.Name {
								for extraKey := range pvalue.Extras {
									propValueExtrasKeys = append(propValueExtrasKeys, extraKey)
								}
							}
						}

						fmt.Println("propValueExtrasKeys: ", propValueExtrasKeys)

						// update or insert properties values extras
						dniPropertiesValuesExtras := []string{}
						for extraKey, extraValue := range value.Extras {
							dniPropertiesValuesExtras = append(dniPropertiesValuesExtras, extraKey)

							extra := &db.PropertyValueExtras{
								ID:    preProp.ID,
								Name:  value.Name,
								Key:   extraKey,
								Value: extraValue,
							}

							if misc.InSlice(extraKey, propValueExtrasKeys) {
								_, err = db.Update(tx, extra, map[string]interface{}{"id": preProp.ID, "name": value.Name, "mkey": extra.Key})
								if err != nil {
									return errors.E(op, err)
								}
							} else {
								_, err = db.Insert(tx, extra)
								if err != nil {
									return errors.E(op, err)
								}
							}
						}

						// delete not in properties values extras list
						err = db.DeletePropertyValueExtrasNotIn(tx, preProp.ID, value.Name, dniPropertiesValuesExtras...)
						if err != nil {
							return errors.E(op, err)
						}
					}

					// delete not in properties values list
					err = db.DeletePropertyValueNotIn(tx, preProp.ID, dniPropertiesValues...)
					if err != nil {
						return errors.E(op, err)
					}

				} else {
					// insert prop that didnt exist before
					prop.ID = uuid.New(uuid.Property)
					prop.CollectionID = pre.ID

					dniProperties = append(dniProperties, prop.ID)

					_, err = db.Insert(tx, prop.Property)
					if err != nil {
						return errors.E(op, err)
					}

					for _, value := range prop.Values {
						value.ID = prop.ID

						_, err = db.Insert(tx, value.PropertyValue)
						if err != nil {
							return errors.E(op, err)
						}

						for extraKey, extraValue := range value.Extras {
							fmt.Println("inserting extra: ", prop.ID, " - ", extraKey, " - ", extraValue)

							extra := &db.PropertyValueExtras{
								ID:    prop.ID,
								Name:  value.Name,
								Key:   extraKey,
								Value: extraValue,
							}

							_, err = db.Insert(tx, extra)
							if err != nil {
								return errors.E(op, err)
							}
						}
					}
				}
			}

			for _, prop := range pre.Properties {
				if !misc.InSlice(prop.ID, dniProperties) {
					// delete prop values & meta that should no longer exist
					err = db.DeletePropertyValues(tx, prop.ID)
					if err != nil {
						return errors.E(op, err)
					}

					err = db.DeletePropertyValueExtras(tx, prop.ID)
					if err != nil {
						return errors.E(op, err)
					}
				}
			}

			err = db.DeletePropertyNotIn(tx, pre.ID, dniProperties...)
			if err != nil {
				return errors.E(op, err)
			}

			preExtras := []string{}
			for extraKey := range pre.Extras {
				preExtras = append(preExtras, extraKey)
			}

			// insert or update extras
			// dni = Delete not in
			dniExtras := []string{}
			for extraKey, extraValue := range col.Extras {
				dniExtras = append(dniExtras, extraKey)

				extra := &db.CollectionExtras{
					ID:    pre.ID,
					Key:   extraKey,
					Value: extraValue,
				}

				if misc.InSlice(extraKey, preExtras) {
					_, err = db.Update(tx, extra, map[string]interface{}{"id": pre.ID, "mkey": extra.Key})
					if err != nil {
						return errors.E(op, err)
					}
				} else {
					_, err = db.Insert(tx, extra)
					if err != nil {
						return errors.E(op, err)
					}
				}
			}

			// delete not in extras list
			err = db.DeleteCollectionExtrasNotIn(tx, pre.ID, dniExtras...)
			if err != nil {
				return errors.E(op, err)
			}
		} else {
			// insert
			col.ID = uuid.New(uuid.Collection)
			col.ProviderID = nfdID
			dniCollection = append(dniCollection, col.ID)

			// collection
			_, err = db.Insert(tx, col.Collection)
			if err != nil {
				return errors.E(op, err)
			}

			// prefixes
			for _, prefix := range col.Prefixes {
				_, err = db.Insert(tx, &db.CollectionPrefix{ID: col.ID, Prefix: prefix})
				if err != nil {
					return errors.E(op, err)
				}
			}

			// asset
			for _, asset := range col.Assets {
				_, err = db.Insert(tx, &db.CollectionAsset{ID: col.ID, AsaID: asset})
				if err != nil {
					return errors.E(op, err)
				}
			}

			// excluded_asset
			for _, asset := range col.ExcludedAssets {
				_, err = db.Insert(tx, &db.CollectionExcludedAsset{ID: col.ID, AsaID: asset})
				if err != nil {
					return errors.E(op, err)
				}
			}

			// artist
			for _, artist := range col.Artists {
				_, err = db.Insert(tx, &db.CollectionArtist{ID: col.ID, Address: artist})
				if err != nil {
					return errors.E(op, err)
				}
			}

			// properties
			for _, prop := range col.Properties {
				prop.ID = uuid.New(uuid.Property)
				prop.CollectionID = col.ID

				_, err = db.Insert(tx, prop.Property)
				if err != nil {
					return errors.E(op, err)
				}

				// property values
				for _, value := range prop.Values {
					value.ID = prop.ID

					_, err = db.Insert(tx, value.PropertyValue)
					if err != nil {
						return errors.E(op, err)
					}

					for extraKey, extraValue := range value.Extras {

						extra := &db.PropertyValueExtras{
							ID:    prop.ID,
							Name:  value.Name,
							Key:   extraKey,
							Value: extraValue,
						}

						_, err = db.Insert(tx, extra)
						if err != nil {
							return errors.E(op, err)
						}
					}
				}
			}

			// extras
			for extraKey, extraValue := range col.Extras {
				_, err = db.Insert(tx, &db.CollectionExtras{ID: col.ID, Key: extraKey, Value: extraValue})
				if err != nil {
					return errors.E(op, err)
				}
			}
		}
	}

	err = db.DeleteCollectionNotIn(tx, nfdID, dniCollection...)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (p *NFDProvider) processFaq(tx *sqlx.Tx, nfdID uint64, faqData []db.CommunityFaq) error {
	const op errors.Op = "ProcessFaq"
	var err error

	err = db.DeleteCommunityFaq(tx, nfdID)
	if err != nil {
		return errors.E(op, err)
	}

	for i, faq := range faqData {
		nfaq := faq

		nfaq.ID = nfdID
		nfaq.Ordering = misc.Pointer(uint64(i))

		_, err = db.Insert(tx, &nfaq)
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

func (p *NFDProvider) processExtras(tx *sqlx.Tx, nfdID uint64, extrasData []db.CommunityExtras) error {
	const op errors.Op = "ProcessExtras"

	extraKeys := map[string]db.CommunityExtras{}
	extras, err := db.GetCommunityExtras(p.DB, nfdID)
	if err != nil && !db.ErrNoRows(err) {
		return errors.E(op, err)
	}

	for _, extra := range *extras {
		extraKeys[extra.Key] = extra
	}

	dniKeys := []string{}
	for _, extra := range extrasData {
		nextra := extra

		dniKeys = append(dniKeys, extra.Key)
		pre, exists := extraKeys[extra.Key]
		if exists {
			_, err = db.Update(tx, &nextra, map[string]interface{}{"id": pre.ID})
			if err != nil {
				return errors.E(op, err)
			}
		} else {

			nextra.ID = nfdID
			_, err = db.Insert(tx, &nextra)
			if err != nil {
				return errors.E(op, err)
			}
		}
	}

	err = db.DeleteCommunityExtrasNotIn(tx, nfdID, dniKeys...)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func GetIPFSData(url string) ([]byte, error) {
	const op errors.Op = "GetIPFSData"
	if !strings.HasPrefix(url, "ipfs://") {
		fmt.Printf("invalid ipfs url: %s\n", url)
		// return nil, errors.E(op, fmt.Errorf("invalid ipfs url: %s", url))
	}

	url = strings.Replace(url, "ipfs://", "https://ipfs.algonode.xyz/ipfs/", 1)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if resp.StatusCode != http.StatusOK || len(body) == 0 {
		return nil, errors.E(op, fmt.Errorf("ipfs request failed: %s", resp.Status))
	}

	return body, nil
}
