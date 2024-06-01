package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
)

type CommunityJson struct {
	ID        uint64 `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Data      string `structs:"data,omitempty" db:"data" json:"data,omitempty"`
	Malformed *bool  `structs:"malformed,omitempty" db:"malformed" json:"malformed,omitempty"`
}

func CommunityJsonTableKeys() []string {
	return []string{"id", "data", "malformed"}
}

func GetCommunityJson[H Handle](h H, id uint64) (*CommunityJson, error) {
	const op errors.Op = "GetCollectionJson"
	query := fmt.Sprintf("select %s from %s.community_json where id = ?", strings.Join(CommunityJsonTableKeys(), ","), arc53Database())

	var json CommunityJson
	err := h.Get(&json, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Json Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &json, nil
}

func DeleteCommunityJson[H Handle](h H, id uint64) error {
	const op errors.Op = "DeleteCommunityJson"
	query := fmt.Sprintf("delete from %s.community_json where id = ?", arc53Database())

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
