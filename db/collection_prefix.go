package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type CollectionPrefix struct {
	ID     string `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Prefix string `structs:"prefix,omitempty" db:"prefix" json:"prefix,omitempty"`
}

func CollectionPrefixTableKeys() []string {
	return []string{"id", "prefix"}
}

func GetCollectionPrefixes[H Handle](h H, id string) (*[]CollectionPrefix, error) {
	const op errors.Op = "GetCollectionPrefixes"
	query := fmt.Sprintf("select %s from %s.collection_prefix where id = ?", strings.Join(CollectionPrefixTableKeys(), ","), arc53Database())

	var prefixes []CollectionPrefix
	err := h.Select(&prefixes, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Prefixes Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &prefixes, nil
}

func DeleteCollectionPrefixes[H Handle](h H, id string) error {
	const op errors.Op = "DeleteCollectionPrefixes"
	query := fmt.Sprintf("delete from %s.collection_prefix where id = ?", arc53Database())

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

// DeleteCollectionPrefixesNotIn deletes collection prefixes that are not included in a list for a given collection
func DeleteCollectionPrefixesNotIn[H Handle](h H, id string, prefixes ...string) error {
	const op errors.Op = "DeleteCollectionPrefixesNotIn"
	var err error
	data := append([]interface{}{id}, misc.ToInterfaceSlice(prefixes)...)
	query := fmt.Sprintf("delete from %s.collection_prefix where id = ?", arc53Database())

	if len(prefixes) > 0 {
		qMarks := []rune(strings.Repeat("?, ", len(prefixes)))
		query += fmt.Sprintf(" and prefix not in (%s)", string(qMarks[0:len(qMarks)-2]))
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
