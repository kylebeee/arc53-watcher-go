package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
)

type CommunityFaq struct {
	ID       uint64  `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Question string  `structs:"q,omitempty" db:"q" json:"q,omitempty"`
	Answer   string  `structs:"a,omitempty" db:"a" json:"a,omitempty"`
	Ordering *uint64 `structs:"ordering,omitempty" db:"ordering" json:"ordering,omitempty"`
}

func CommunityFaqTableKeys() []string {
	return []string{"id", "q", "a", "ordering"}
}

func GetCommunityFaq[H Handle](h H, id, start, limit uint64) (*[]CommunityFaq, error) {
	const op errors.Op = "GetCommunityFaq"
	query := fmt.Sprintf("select %s from %s.community_faq where id = ? order by ordering asc limit ?, ?", strings.Join(CommunityFaqTableKeys(), ","), arc53Database())

	var faq []CommunityFaq
	err := h.Select(&faq, query, id, start, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "CommunityFaq Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &faq, nil
}

func DeleteCommunityFaq[H Handle](h H, id uint64) error {
	const op errors.Op = "DeleteCommunityFaq"
	query := fmt.Sprintf("delete from %s.community_faq where id = ?", arc53Database())
	var err error

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
		_, err = h.Exec(query, id)
		if err != nil {
			return errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	return nil
}
