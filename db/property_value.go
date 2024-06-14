package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type PropertyValue struct {
	// ID is the id of the property
	ID                    string  `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Name                  string  `structs:"name,omitempty" db:"name" json:"name,omitempty"`
	Image                 *string `structs:"image,omitempty" db:"image" json:"image,omitempty"`
	ImageIntegrity        *string `structs:"image_integrity,omitempty" db:"image_integrity" json:"image_integrity,omitempty"`
	ImageMimeType         *string `structs:"image_mimetype,omitempty" db:"image_mimetype" json:"image_mimetype,omitempty"`
	AnimationURL          *string `structs:"animation_url,omitempty" db:"animation_url" json:"animation_url,omitempty"`
	AnimationURLIntegrity *string `structs:"animation_url_integrity,omitempty" db:"animation_url_integrity" json:"animation_url_integrity,omitempty"`
	AnimationURLMimeType  *string `structs:"animation_url_mimetype,omitempty" db:"animation_url_mimetype" json:"animation_url_mimetype,omitempty"`
}

func PropertyValueTableKeys() []string {
	return []string{"id", "name", "image", "image_integrity", "image_mimetype", "animation_url", "animation_url_integrity", "animation_url_mimetype"}
}

func GetPropertyValues[H Handle](h H, id string) (*[]PropertyValue, error) {
	const op errors.Op = "GetPropertyValues"
	query := fmt.Sprintf("select %s from %s.property_value where id = ?", strings.Join(PropertyValueTableKeys(), ","), arc53Database())

	var propertyValues []PropertyValue
	err := h.Select(&propertyValues, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Property Values Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &propertyValues, nil
}

func GetPropertyValueByName[H Handle](h H, id string, name string) (*PropertyValue, error) {
	const op errors.Op = "GetPropertyValueByName"
	query := fmt.Sprintf("select %s from %s.property_value where id = ? and name = ?", strings.Join(PropertyValueTableKeys(), ","), arc53Database())

	var propertyValue PropertyValue
	err := h.Get(&propertyValue, query, id, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Property Value Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &propertyValue, nil
}

func DeletePropertyValues[H Handle](h H, id string) error {
	const op errors.Op = "DeletePropertyValues"
	query := fmt.Sprintf("delete from %s.property_value where id = ?", arc53Database())

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

// DeletePropertyValueNotIn deletes properties that are not included in a list for a given provider ID
func DeletePropertyValueNotIn[H Handle](h H, id string, names ...string) error {
	const op errors.Op = "DeletePropertyValueNotIn"
	var err error
	data := append([]interface{}{id}, misc.ToInterfaceSlice(names)...)
	query := fmt.Sprintf("delete from %s.property_value where id = ?", arc53Database())

	if len(names) > 0 {
		qMarks := []rune(strings.Repeat("?, ", len(names)))
		query += fmt.Sprintf(" and name not in (%s)", string(qMarks[0:len(qMarks)-2]))
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
