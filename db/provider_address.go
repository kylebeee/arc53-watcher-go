package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type ProviderAddress struct {
	// ID is the app id of the Provider Contract
	ID      uint64 `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Address string `structs:"address,omitempty" db:"address" json:"address,omitempty"`
}

func ProviderAddressTableKeys() []string {
	return []string{"id", "address"}
}

func GetAllProviderAddresses[H DBStruct](h H) (*[]ProviderAddress, error) {
	const op errors.Op = "GetAllProviderAddresses"
	query := fmt.Sprintf("select %s from %s.provider_address", strings.Join(ProviderAddressTableKeys(), ","), arc53Database())
	var list []ProviderAddress

	err := h.Select(&list, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "wallets not found")
		}
		return nil, errors.E(pkg, op, err)
	}

	if !(len(list) > 0) {
		return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, fmt.Errorf("wallets not found"))
	}

	return &list, nil
}

func GetAllProviderAddressesByType[H Handle](h H, t string) (*[]ProviderAddress, error) {
	const op errors.Op = "GetAllProviderAddressesByType"
	query := fmt.Sprintf("select %s from %s.provider_address where type = ?", strings.Join(ProviderAddressTableKeys(), ","), arc53Database())
	var list []ProviderAddress

	err := h.Select(&list, query, t)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "wallets not found")
		}
		return nil, errors.E(pkg, op, err)
	}

	if !(len(list) > 0) {
		return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, fmt.Errorf("wallets not found"))
	}

	return &list, nil

}

func GetProviderAddresses[H DBStruct](h H, id uint64) (*[]ProviderAddress, error) {
	const op errors.Op = "GetProviderAddresses"
	query := fmt.Sprintf("select %s from %s.provider_address where id = ?", strings.Join(ProviderAddressTableKeys(), ","), arc53Database())
	var list []ProviderAddress

	err := h.Select(&list, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "wallets not found")
		}
		return nil, errors.E(pkg, op, err)
	}

	if !(len(list) > 0) {
		return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, fmt.Errorf("wallets not found"))
	}

	return &list, nil
}

func GetProviderAddressesAddressesByAdjacentAddresses[H DBStruct](h H, addresses []string) (*[]string, error) {
	const op errors.Op = "GetProviderAddressesAddressesByAdjacentAddresses"
	query := fmt.Sprintf("select distinct(address) from %s.provider_address where id in (select id from %v.provider_address where address in (%s))", arc53Database(), arc53Database(), strings.Repeat("?, ", len(addresses))[0:(len(addresses)*3)-2])
	var list []string

	err := h.Select(&list, query, misc.ToInterfaceSlice(addresses)...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "wallets not found")
		}
		return nil, errors.E(pkg, op, err)
	}

	if !(len(list) > 0) {
		return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, fmt.Errorf("wallets not found"))
	}

	return &list, nil
}

func DeleteProviderAddressesByIDAndAddress[H DBStruct](h H, id uint64, address string) error {
	const op errors.Op = "DeleteProviderAddressesByIDAndAddress"
	query := fmt.Sprintf("delete from %s.provider_address where id = ? and address = ?", arc53Database())
	var err error

	switch h := any(h).(type) {
	case *sqlx.Tx:
		stmt, err := h.Prepare(query)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Prepare Query")
		}

		_, err = stmt.Exec(id, address)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	case *sqlx.DB:
		_, err = h.Exec(query, id, address)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	return nil
}

// DeleteProviderAddressNotIn deletes collections that are not included in a list for a given NFD
func DeleteProviderAddressNotIn[H Handle](h H, id uint64, addresses ...string) error {
	const op errors.Op = "DeleteProviderAddressNotIn"
	var err error
	data := append([]interface{}{id}, misc.ToInterfaceSlice(addresses)...)
	fmt.Println(data)
	query := fmt.Sprintf("delete from %s.provider_address where id = ?", arc53Database())

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
