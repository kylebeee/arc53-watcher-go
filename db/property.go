package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type Property struct {
	ID           string `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	CollectionID string `structs:"collection_id,omitempty" db:"collection_id" json:"collection_id,omitempty"`
	Name         string `structs:"name,omitempty" db:"name" json:"name,omitempty"`
}

func PropertyTableKeys() []string {
	return []string{"id", "collection_id", "name"}
}

func GetProperties[H Handle](h H, collectionID string) (*[]Property, error) {
	const op errors.Op = "GetProperties"
	query := fmt.Sprintf("select %s from %s.property where collection_id = ?", strings.Join(PropertyTableKeys(), ","), arc53Database())

	var properties []Property
	err := h.Select(&properties, query, collectionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Properties Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &properties, nil
}

func GetPropertiesByName[H Handle](h H, collectionID string, name string) (*[]Property, error) {
	const op errors.Op = "GetProperties"
	query := fmt.Sprintf("select %s from %s.property where collection_id = ? and name = ?", strings.Join(PropertyTableKeys(), ","), arc53Database())

	var properties []Property
	err := h.Select(&properties, query, collectionID, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Properties Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &properties, nil
}

func GetPropertiesWhereNameIn[H Handle](h H, collectionID string, names ...string) (*[]Property, error) {
	const op errors.Op = "GetProperties"
	query := fmt.Sprintf("select %s from %s.property where collection_id = ? and name in (%s)", strings.Join(PropertyTableKeys(), ","), arc53Database(), strings.Repeat("?, ", len(names))[0:(len(names)*3)-2])

	var properties []Property
	err := h.Select(&properties, query, append([]interface{}{collectionID}, misc.ToInterfaceSlice(names)...)...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Properties Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &properties, nil
}

func DeleteCollectionProperties[H Handle](h H, collectionID string) error {
	const op errors.Op = "DeleteCollectionProperties"
	query := fmt.Sprintf("delete from %s.property where collection_id = ?", arc53Database())

	switch h := any(h).(type) {
	case *sqlx.Tx:
		stmt, err := h.Prepare(query)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Prepare Query")
		}

		_, err = stmt.Exec(collectionID)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	case *sqlx.DB:
		_, err := h.Exec(query, collectionID)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	return nil
}

// DeletePropertyNotIn deletes properties that are not included in a list for a given provider ID
func DeletePropertyNotIn[H Handle](h H, collectionID string, ids ...string) error {
	const op errors.Op = "DeletePropertyNotIn"
	var err error
	data := append([]interface{}{collectionID}, misc.ToInterfaceSlice(ids)...)
	query := fmt.Sprintf("delete from %s.property where collection_id = ?", arc53Database())

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
