package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/kylebeee/arc53-watcher-go/errors"
)

type CollectionSettings struct {
	ID         string  `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	HideBlanks *string `structs:"hide_blanks,omitempty" db:"hide_blanks" json:"hide_blanks,omitempty"`
}

func CollectionSettingsTableKeys() []string {
	return []string{"id", "hide_blanks"}
}

func GetCollectionSettings[H Handle](h H, id string) (*CollectionSettings, error) {
	const op errors.Op = "GetCollectionSettings"
	query := fmt.Sprintf("select %s from %s.collection_settings where id = ?", strings.Join(CollectionSettingsTableKeys(), ","), arc53Database())

	var settings CollectionSettings
	err := h.Get(&settings, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Settings Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &settings, nil
}

func GetAllCollectionSettings[H Handle](h H) (*CollectionSettings, error) {
	const op errors.Op = "GetAllCollectionSettings"
	query := fmt.Sprintf("select %s from %s.collection_settings", strings.Join(CollectionSettingsTableKeys(), ","), arc53Database())

	var settings CollectionSettings
	err := h.Select(&settings, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Settings Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &settings, nil
}
