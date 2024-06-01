package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

type CollectionArtist struct {
	ID      string `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	Address string `structs:"address,omitempty" db:"address" json:"address,omitempty"`
}

func CollectionArtistTableKeys() []string {
	return []string{"id", "address"}
}

func GetCollectionArtist[H Handle](h H) (*[]CollectionArtist, error) {
	const op errors.Op = "GetCollectionArtist"
	query := fmt.Sprintf("select %s from %s.collection_artist", strings.Join(CollectionArtistTableKeys(), ","), arc53Database())

	var ccs []CollectionArtist
	err := h.Select(&ccs, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "CollectionArtist Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}
	return &ccs, nil
}

func GetCollectionArtistByCollection[H Handle](h H, id string) (*[]CollectionArtist, error) {
	const op errors.Op = "GetCollectionArtistByCollection"
	query := fmt.Sprintf("select %s from %s.collection_artist where id = ?", strings.Join(CollectionArtistTableKeys(), ","), arc53Database())

	var ccs []CollectionArtist
	err := h.Select(&ccs, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "CollectionArtist Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}
	return &ccs, nil
}

func DeleteCollectionArtists[H Handle](h H, id string) error {
	const op errors.Op = "DeleteCollectionArtists"
	query := fmt.Sprintf("delete from %s.collection_artist where id = ?", arc53Database())

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

// DeleteCollectionArtistsNotIn deletes collection artists that are not included in a list for a given collection
func DeleteCollectionArtistsNotIn[H Handle](h H, id string, addresses ...string) error {
	const op errors.Op = "DeleteCollectionArtistsNotIn"
	var err error
	data := append([]interface{}{id}, misc.ToInterfaceSlice(addresses)...)
	query := fmt.Sprintf("delete from %s.collection_artist where id = ?", arc53Database())

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
