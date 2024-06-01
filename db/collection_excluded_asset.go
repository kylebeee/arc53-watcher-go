package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type CollectionExcludedAsset struct {
	ID    string `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	AsaID uint64 `structs:"asa_id,omitempty" db:"asa_id" json:"asa_id"`
}

func CollectionExcludedAssetTableKeys() []string {
	return []string{"id", "asa_id"}
}

func GetCollectionExcludedAssets[H Handle](h H, id string) (*[]CollectionExcludedAsset, error) {
	const op errors.Op = "GetCollectionExcludedAssets"
	query := fmt.Sprintf("select %s from %s.collection_excluded_asset where id = ?", strings.Join(CollectionExcludedAssetTableKeys(), ","), arc53Database())

	var assets []CollectionExcludedAsset
	err := h.Select(&assets, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Excluded Assets Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &assets, nil
}

func GetCollectionExcludedAssetByAsaID[H Handle](h H, asaID uint64) (*[]CollectionExcludedAsset, error) {
	const op errors.Op = "GetCollectionExcludedAssetByAsaID"
	query := fmt.Sprintf("select %s from %s.collection_excluded_asset where asa_id = ?", strings.Join(CollectionExcludedAssetTableKeys(), ","), arc53Database())

	var assets []CollectionExcludedAsset
	err := h.Select(&assets, query, asaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Excluded Assets Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &assets, nil
}

func DeleteCollectionExcludedAssets[H Handle](h H, id string) error {
	const op errors.Op = "DeleteCollectionExcludedAssets"
	query := fmt.Sprintf("delete from %s.collection_excluded_asset where id = ?", arc53Database())

	switch h := any(h).(type) {
	case *sqlx.Tx:
		stmt, err := h.Prepare(query)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Prepare Query")
		}

		_, err = stmt.Exec(id)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	case *sqlx.DB:
		_, err := h.Exec(query, id)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	return nil
}

// DeleteCollectionExcludedAssetsNotIn deletes collection excluded assets that are not included in a list for a given collection
func DeleteCollectionExcludedAssetsNotIn[H Handle](h H, id string, asaIDs ...uint64) error {
	const op errors.Op = "DeleteCollectionExcludedAssetsNotIn"
	var err error
	data := append([]interface{}{id}, misc.ToInterfaceSlice(asaIDs)...)
	query := fmt.Sprintf("delete from %s.collection_excluded_asset where id = ?", arc53Database())

	if len(asaIDs) > 0 {
		qMarks := []rune(strings.Repeat("?, ", len(asaIDs)))
		query += fmt.Sprintf(" and asa_id not in (%s)", string(qMarks[0:len(qMarks)-2]))
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
