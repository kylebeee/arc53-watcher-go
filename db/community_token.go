package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type CommunityToken struct {
	ID        uint64  `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	AssetID   uint64  `structs:"asset_id,omitempty" db:"asset_id" json:"asset_id"`
	Image     *string `structs:"image,omitempty" db:"image" json:"image,omitempty"`
	Integrity *string `structs:"image_integrity,omitempty" db:"image_integrity" json:"image_integrity,omitempty"`
	Mime      *string `structs:"image_mimetype,omitempty" db:"image_mimetype" json:"image_mimetype,omitempty"`
}

func CommunityTokenTableKeys() []string {
	return []string{"id", "asset_id", "image", "image_integrity", "image_mimetype"}
}

func GetCommunityTokens[H Handle](h H, id uint64) (*[]CommunityToken, error) {
	const op errors.Op = "GetCommunityTokens"
	query := fmt.Sprintf("select %s from %s.community_token where id = ?", strings.Join(CommunityTokenTableKeys(), ","), arc53Database())

	var assets []CommunityToken
	err := h.Select(&assets, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Assets Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &assets, nil
}

func GetCommunityTokenByAsaID[H Handle](h H, asaID uint64) (*CommunityToken, error) {
	const op errors.Op = "GetCommunityTokenByAsaID"
	query := fmt.Sprintf("select %s from %s.community_token where asset_id = ?", strings.Join(CommunityTokenTableKeys(), ","), arc53Database())

	var asset CommunityToken
	err := h.Get(&asset, query, asaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Assets Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &asset, nil
}

func DeleteCommunityTokens[H Handle](h H, id uint64) error {
	const op errors.Op = "DeleteCommunityTokens"
	query := fmt.Sprintf("delete from %s.community_token where id = ?", arc53Database())

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

// DeleteCommunityTokensNotIn deletes collection assets that are not included in a list for a given collection
func DeleteCommunityTokensNotIn[H Handle](h H, id uint64, asaIDs ...uint64) error {
	const op errors.Op = "DeleteCommunityTokensNotIn"
	var err error
	data := append([]interface{}{id}, misc.ToInterfaceSlice(asaIDs)...)
	query := fmt.Sprintf("delete from %s.community_token where id = ?", arc53Database())

	if len(asaIDs) > 0 {
		qMarks := []rune(strings.Repeat("?, ", len(asaIDs)))
		query += fmt.Sprintf(" and asset_id not in (%s)", string(qMarks[0:len(qMarks)-2]))
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
