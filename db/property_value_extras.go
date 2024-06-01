package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type PropertyValueExtras struct {
	ID    string `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Name  string `structs:"name,omitempty" db:"name" json:"name,omitempty"`
	Key   string `structs:"mkey,omitempty" db:"mkey" json:"key,omitempty"`
	Value string `structs:"mvalue,omitempty" db:"mvalue" json:"value,omitempty"`
}

func PropertyValueExtrasTableKeys() []string {
	return []string{"id", "name", "mkey", "mvalue"}
}

func GetPropertyValueExtras[H Handle](h H, id string) (*[]PropertyValueExtras, error) {
	const op errors.Op = "GetPropertyValueExtras"
	query := fmt.Sprintf("select %s from %s.property_value_extras where id = ?", strings.Join(PropertyValueExtrasTableKeys(), ","), arc53Database())

	var extras []PropertyValueExtras
	err := h.Select(&extras, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Property Value Extras Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &extras, nil
}

func GetPropertyValueExtrasByName[H Handle](h H, id string, name string) (*[]PropertyValueExtras, error) {
	const op errors.Op = "GetPropertyValueExtrasByName"
	query := fmt.Sprintf("select %s from %s.property_value_extras where id = ? and name = ?", strings.Join(PropertyValueExtrasTableKeys(), ","), arc53Database())

	var extras []PropertyValueExtras
	err := h.Select(&extras, query, id, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Property Value Extras Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &extras, nil
}

func DeletePropertyValueExtras[H Handle](h H, id string) error {
	const op errors.Op = "DeletePropertyValueExtras"
	query := fmt.Sprintf("delete from %s.property_value_extras where id = ?", arc53Database())

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

// DeletePropertyValueExtrasNotIn deletes properties that are not included in a list for a given NFD
func DeletePropertyValueExtrasNotIn[H Handle](h H, id string, name string, keys ...string) error {
	const op errors.Op = "DeletePropertyValueExtrasNotIn"
	var err error
	data := append([]interface{}{id, name}, misc.ToInterfaceSlice(keys)...)
	query := fmt.Sprintf("delete from %s.property_value_extras where id = ? and name = ?", arc53Database())

	if len(keys) > 0 {
		qMarks := []rune(strings.Repeat("?, ", len(keys)))
		query += fmt.Sprintf(" and mkey not in (%s)", string(qMarks[0:len(qMarks)-2]))
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
