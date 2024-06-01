package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type CollectionAddress struct {
	ID      string `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Address string `structs:"address,omitempty" db:"address" json:"address,omitempty"`
}

func CollectionAddressTableKeys() []string {
	return []string{"id", "address"}
}

func GetAllCollectionAddress[H Handle](h H) (*[]CollectionAddress, error) {
	const op errors.Op = "GetCollectionAddress"
	query := fmt.Sprintf("select %s from %s.collection_address", strings.Join(CollectionAddressTableKeys(), ","), arc53Database())

	var ccs []CollectionAddress
	err := h.Select(&ccs, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "CollectionAddress Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}
	return &ccs, nil
}

func GetCollectionAddresses[H Handle](h H, id string) (*[]CollectionAddress, error) {
	const op errors.Op = "GetCollectionAddresses"
	query := fmt.Sprintf("select %s from %s.collection_address where id = ?", strings.Join(CollectionAddressTableKeys(), ","), arc53Database())

	var ccs []CollectionAddress
	err := h.Select(&ccs, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "CollectionAddress Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}
	return &ccs, nil
}

func DeleteCollectionAddresses[H Handle](h H, id string) error {
	const op errors.Op = "DeleteCollectionAddresses"
	query := fmt.Sprintf("delete from %s.collection_address where id = ?", arc53Database())

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

// DeleteCollectionAddressesNotIn deletes collection addresss that are not included in a list for a given collection
func DeleteCollectionAddressesNotIn[H Handle](h H, id string, addresses ...string) error {
	const op errors.Op = "DeleteCollectionAddressesNotIn"
	var err error
	data := append([]interface{}{id}, misc.ToInterfaceSlice(addresses)...)
	query := fmt.Sprintf("delete from %s.collection_address where id = ?", arc53Database())

	if len(addresses) > 0 {
		qMarks := []rune(strings.Repeat("?, ", len(addresses)))
		query += fmt.Sprintf(" and address not in (%s)", string(qMarks[0:len(qMarks)-2]))
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
