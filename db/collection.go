package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type BlockchainNetwork string

const (
	Bitcoin  BlockchainNetwork = "bitcoin"
	Ethereum BlockchainNetwork = "ethereum"
	Algorand BlockchainNetwork = "algorand"
	Solana   BlockchainNetwork = "solana"
)

type Collection struct {
	ID          string             `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	ProviderID  uint64             `structs:"provider_id,omitempty" db:"provider_id" json:"provider_id,omitempty"`
	Name        string             `structs:"name,omitempty" db:"name" json:"name,omitempty"`
	Description *string            `structs:"description,omitempty" db:"description" json:"description,omitempty"`
	Banner      *uint64            `structs:"banner,omitempty" db:"banner" json:"banner,omitempty"`
	Avatar      *uint64            `structs:"avatar,omitempty" db:"avatar" json:"avatar,omitempty"`
	Network     *BlockchainNetwork `structs:"network,omitempty" db:"network" json:"network,omitempty"`
	Explicit    *bool              `structs:"explicit,omitempty" db:"explicit" json:"explicit,omitempty"`
}

func CollectionTableKeys() []string {
	return []string{"id", "provider_id", "name", "description", "banner", "avatar", "network", "explicit"}
}

func GetCollection[H Handle](h H, id string) (*Collection, error) {
	const op errors.Op = "GetCollection"
	query := fmt.Sprintf("select %s from %s.collection where id = ?", strings.Join(CollectionTableKeys(), ","), arc53Database())

	var c Collection
	err := h.Get(&c, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}
	return &c, nil
}

func GetCollectionsByProviderID[H Handle](h H, providerID uint64) (*[]Collection, error) {
	const op errors.Op = "GetCollectionsPaginated"
	query := fmt.Sprintf("select %s from %s.collection where provider_id = ?", strings.Join(CollectionTableKeys(), ","), arc53Database())

	var collections []Collection
	err := h.Select(&collections, query, providerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collections Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &collections, nil
}

func GetCollectionCreatorWallets[H Handle](h H, id string) (*[]ProviderAddress, error) {
	const op errors.Op = "GetCollectionCreatorWallets"
	query := fmt.Sprintf("select %s from %s.nfd_wallet where id = (select nfd_id from %s.collection where id = ?) and verified = 1 order by deposit desc", strings.Join(ProviderAddressTableKeys(), ","), arc53Database(), arc53Database())

	var wallets []ProviderAddress
	err := h.Select(&wallets, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Wallets Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &wallets, nil
}

// GetCollectionByAssetID is a query that returns a collection by asset ID, this is used for reverse lookup when someone is viewing a specific asset
func GetCollectionByAssetID[H Handle](h H, assetID uint64, creator string, unitName string) (*Collection, error) {
	const op errors.Op = "GetCollectionByAssetID"
	query := fmt.Sprintf("select %s from %s.collection where provider_id in(select id from %s.provider_address where address = ?) and ((exists(select id from %s.collection_prefix where %s.collection.id = id and left(?, char_length(prefix)) = prefix) and not exists(select id from %s.collection_excluded_asset where %s.collection.id = id and asa_id = ?)) or (exists(select id from %s.collection_asset where %s.collection.id = id and asa_id = ?)))", strings.Join(CollectionTableKeys(), ","), arc53Database(), arc53Database(), arc53Database(), arc53Database(), arc53Database(), arc53Database(), arc53Database(), arc53Database())
	var c Collection

	err := h.Get(&c, query, creator, unitName, assetID, assetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &c, nil
}

func GetCollectionsPaginated[H Handle](h H, start, limit uint64) (*[]Collection, error) {
	const op errors.Op = "GetCollectionsPaginated"
	query := fmt.Sprintf("select %s from %s.collection limit ?, ?", strings.Join(CollectionTableKeys(), ","), arc53Database())

	var collections []Collection
	err := h.Select(&collections, query, start, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collections Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &collections, nil
}

func DeleteCollectionsByNFDID[H Handle](h H, nfdID uint64) error {
	const op errors.Op = "DeleteCollectionsByNFDID"
	query := fmt.Sprintf("delete from %s.collection where nfd_id = ?", arc53Database())

	switch h := any(h).(type) {
	case *sqlx.Tx:
		stmt, err := h.Prepare(query)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Prepare Query")
		}

		_, err = stmt.Exec(nfdID)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	case *sqlx.DB:
		_, err := h.Exec(query, nfdID)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	return nil
}

// DeleteCollectionNotIn deletes collections that are not included in a list for a given NFD
func DeleteCollectionNotIn[H Handle](h H, nfdID uint64, ids ...string) error {
	const op errors.Op = "DeleteCollectionNotIn"
	var err error
	data := append([]interface{}{nfdID}, misc.ToInterfaceSlice(ids)...)
	query := fmt.Sprintf("delete from %s.collection where nfd_id = ?", arc53Database())

	if len(ids) > 0 {
		qMarks := []rune(strings.Repeat("?, ", len(ids)))
		query += fmt.Sprintf(" and id not in (%s)", string(qMarks[0:len(qMarks)-2]))
	}

	switch h := any(h).(type) {
	case *sqlx.Tx:
		stmt, err := h.Prepare(query)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Prepare Query")
		}

		_, err = stmt.Exec(data...)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	case *sqlx.DB:
		_, err = h.Exec(query, data...)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	return nil
}
