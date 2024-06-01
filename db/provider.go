package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/kylebeee/arc53-watcher-go/errors"
)

type Provider struct {
	ID   uint64 `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Type string `structs:"type,omitempty" db:"type" json:"type,omitempty"`
}

func ProviderTableKeys() []string {
	return []string{"id", "type"}
}

func GetProvider[H Handle](h H, id uint64) (*Provider, error) {
	const op errors.Op = "GetAllProvidersByType"
	query := fmt.Sprintf("select %s from %s.provider where id = ?", strings.Join(ProviderTableKeys(), ","), arc53Database())
	var provider Provider

	err := h.Get(&provider, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "providers not found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &provider, nil
}

func GetAllProvidersByType[H Handle](h H, t string) (*[]Provider, error) {
	const op errors.Op = "GetAllProvidersByType"
	query := fmt.Sprintf("select %s from %s.provider where type = ?", strings.Join(ProviderTableKeys(), ","), arc53Database())
	var list []Provider

	err := h.Select(&list, query, t)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "providers not found")
		}
		return nil, errors.E(pkg, op, err)
	}

	if !(len(list) > 0) {
		return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, fmt.Errorf("providers not found"))
	}

	return &list, nil
}
