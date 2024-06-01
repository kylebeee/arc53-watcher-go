package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
)

type Community struct {
	ID      uint64 `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Version string `structs:"version,omitempty" db:"version" json:"version,omitempty"`
}

func CommunityTableKeys() []string {
	return []string{"id", "version"}
}

func IsCommunity[H Handle](h H, id uint64) (bool, error) {
	const op errors.Op = "IsCommunity"
	query := fmt.Sprintf("select exists(select id from %v.community where id = ?)", arc53Database())

	var exists bool
	err := h.Get(&exists, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.E(pkg, op, err)
	}

	return exists, nil
}

func GetCommunities[H Handle](h H, start, limit uint64) (*[]Community, error) {
	const op errors.Op = "GetCommunities"
	query := fmt.Sprintf("select %v from %v.community order by akta desc limit ?, ?", strings.Join(CommunityTableKeys(), ","), arc53Database())

	var communities []Community
	err := h.Select(&communities, query, start, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Community Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}
	return &communities, nil
}

func GetAllCommunityVerifiedAddresses[H Handle](h H) ([]string, error) {
	const op errors.Op = "GetAllCommunityVerifiedAddresses"
	query := fmt.Sprintf("select address from %v.provider_address where id in (select id from %v.community)", arc53Database(), arc53Database())

	var wallets []string
	err := h.Select(&wallets, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Community Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}
	return wallets, nil
}

func GetCommunity[H Handle](h H, id uint64) (*Community, error) {
	const op errors.Op = "GetCommunity"
	query := fmt.Sprintf("select %v from %v.community where id = ?", strings.Join(CommunityTableKeys(), ","), arc53Database())

	var community Community
	err := h.Get(&community, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Community Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &community, nil
}

func DeleteCommunity[H Handle](h H, id uint64) error {
	const op errors.Op = "DeleteCommunity"
	query := fmt.Sprintf("delete from %v.community where id = ?", arc53Database())

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
