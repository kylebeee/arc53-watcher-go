package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type CommunityAssociate struct {
	ID        uint64  `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Address   string  `structs:"address,omitempty" db:"address" json:"address,omitempty"`
	Role      string  `structs:"role,omitempty" db:"role" json:"role,omitempty"`
	Confirmed *bool   `structs:"confirmed,omitempty" db:"confirmed" json:"confirmed,omitempty"`
	Txn       *string `structs:"txn,omitempty" db:"txn" json:"txn,omitempty"`
}

func CommunityAssociateTableKeys() []string {
	return []string{"id", "address", "role", "confirmed", "txn"}
}

func GetCommunityAssociates[H Handle](h H, id uint64) (*[]CommunityAssociate, error) {
	const op errors.Op = "GetCommunityAssociates"
	query := fmt.Sprintf("select %s from %s.community_associate where id = ?", strings.Join(CommunityAssociateTableKeys(), ","), arc53Database())

	var associates []CommunityAssociate
	err := h.Select(&associates, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Associates Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &associates, nil
}

func GetCommunityAssociate[H Handle](h H, id uint64, address string) (*CommunityAssociate, error) {
	const op errors.Op = "GetCommunityAssociate"
	query := fmt.Sprintf("select %s from %s.community_associate where id = ? and address = ?", strings.Join(CommunityAssociateTableKeys(), ","), arc53Database())

	var associate CommunityAssociate
	err := h.Get(&associate, query, id, address)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Associate Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &associate, nil
}

func GetCommunityAssociateByAddress[H Handle](h H, address string) (*[]CommunityAssociate, error) {
	const op errors.Op = "GetCommunityAssociateByAddress"
	query := fmt.Sprintf("select %s from %s.community_associate where address = ?", strings.Join(CommunityAssociateTableKeys(), ","), arc53Database())

	var associates []CommunityAssociate
	err := h.Select(&associates, query, address)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Collection Associate Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &associates, nil
}

func DeleteCommunityAssociate[H Handle](h H, id uint64, address string) error {
	const op errors.Op = "DeleteCommunityAssociate"
	query := fmt.Sprintf("delete from %s.community_associate where id = ? and address = ?", arc53Database())

	switch h := any(h).(type) {
	case *sqlx.Tx:
		_, err := h.Exec(query, id, address)
		if err != nil {
			return errors.E(pkg, op, err)
		}
	case *sqlx.DB:
		_, err := h.Exec(query, id, address)
		if err != nil {
			return errors.E(pkg, op, err)
		}
	}

	return nil
}

// DeleteCommunityAssociatesNotIn deletes collection associates that are not included in a list for a given community
func DeleteCommunityAssociatesNotIn[H Handle](h H, id uint64, addresses ...string) error {
	const op errors.Op = "DeleteCommunityAssociatesNotIn"
	var err error
	data := append([]interface{}{id}, misc.ToInterfaceSlice(addresses)...)
	query := fmt.Sprintf("delete from %s.community_associate where id = ?", arc53Database())

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
