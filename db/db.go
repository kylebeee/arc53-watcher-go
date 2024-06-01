package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/structs"
	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
)

const pkg errors.Pkg = "db"

type DBStruct interface {
	*sqlx.DB | *sqlx.Tx
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

type TableKeySlice string

const (
	TKSID    TableKeySlice = "id"
	TKSTiny  TableKeySlice = "tiny"
	TKSSmall TableKeySlice = "small"
	TKSFull  TableKeySlice = "full"
)

type Handle interface {
	*sqlx.DB | *sqlx.Tx
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Preparex(query string) (*sqlx.Stmt, error)
	Exec(query string, args ...any) (sql.Result, error)
}

// Insert inserts an object into the database
func Insert[H Handle, S DBObject](h H, s S) (int64, error) {
	const op errors.Op = "Insert"
	var (
		res    sql.Result
		err    error
		query  = ""
		keys   = ""
		values []interface{}
	)

	table := getTable(s)
	if table == "" {
		return 0, errors.E(pkg, op, errors.Database, fmt.Errorf("invalid table struct"))
	}
	// fmt.Println("insert on table: ", table)

	params := structs.Map(s)

	for k, v := range params {
		keys += fmt.Sprintf("%s, ", k)
		values = append(values, v)
	}
	q := strings.Repeat("?, ", len(values))
	query = fmt.Sprintf("insert into %s (%s) values (%s)", table, keys[0:len(keys)-2], q[0:len(q)-2])

	switch h := any(h).(type) {
	case *sqlx.Tx:
		stmt, err := h.Prepare(query)
		if err != nil {
			return 0, errors.E(pkg, op, errors.Database, err, "Failed to Prepare Query")
		}

		res, err = stmt.Exec(values...)
		if err != nil {
			return 0, errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	case *sqlx.DB:
		res, err = h.Exec(query, values...)
		if err != nil {
			return 0, errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.E(pkg, op, errors.Database, err, "Failed to Retrieve ID")
	}

	return id, nil
}

// Update updates an object in the database
func Update[H Handle, S DBObject](h H, s S, where map[string]interface{}) (*sql.Result, error) {
	const op errors.Op = "Update"
	var (
		res    sql.Result
		err    error
		query  = ""
		keys   = ""
		w      = ""
		values []interface{}
	)

	table := getTable(s)
	if table == "" {
		return nil, errors.E(pkg, op, errors.Database, fmt.Errorf("invalid table struct"))
	}
	// fmt.Println("update on table: ", table)

	params := structs.Map(s)

	for k, v := range params {
		keys += fmt.Sprintf("%s=?, ", k)
		values = append(values, v)
	}

	for k, v := range where {
		w += fmt.Sprintf("%s=? and ", k)
		values = append(values, v)
	}

	query = fmt.Sprintf("update %s set %s where %s", table, keys[0:len(keys)-2], w[0:len(w)-5])

	switch h := any(h).(type) {
	case *sqlx.Tx:
		stmt, err := h.Prepare(query)
		if err != nil {
			return nil, errors.E(pkg, op, errors.Database, err, "Failed to Prepare Query")
		}

		res, err = stmt.Exec(values...)
		if err != nil {
			return nil, errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	case *sqlx.DB:
		res, err = h.Exec(query, values...)
		if err != nil {
			return nil, errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	return &res, nil
}

// Upsert inserts into or updates an object in the database
func Upsert[H Handle, S DBObject](h H, s S) (int64, error) {
	const op errors.Op = "Upsert"
	var (
		res    sql.Result
		err    error
		query  = ""
		keys   = ""
		dupe   = ""
		values []interface{}
	)

	table := getTable(s)
	if table == "" {
		return 0, errors.E(pkg, op, errors.Database, fmt.Errorf("invalid table struct"))
	}

	params := structs.Map(s)
	_, id_preset := params["id"]
	if id_preset {
		return 0, errors.E(pkg, op, errors.Database, fmt.Errorf("dont use upsert with preset ids"))
	}

	for k, v := range params {
		keys += fmt.Sprintf("%s, ", k)
		dupe += fmt.Sprintf("%s = VALUES(%s), ", k, k)
		values = append(values, v)
	}

	q := strings.Repeat("?, ", len(values))

	query = fmt.Sprintf("insert into %s (%s) values (%s) on duplicate key update id=LAST_INSERT_ID(id), %s", table, keys[0:len(keys)-2], q[0:len(q)-2], dupe[0:len(dupe)-2])

	switch h := any(h).(type) {
	case *sqlx.Tx:
		stmt, err := h.Prepare(query)
		if err != nil {
			return 0, errors.E(pkg, op, errors.Database, err, "Failed to Prepare Query")
		}

		res, err = stmt.Exec(values...)
		if err != nil {
			return 0, errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	case *sqlx.DB:
		res, err = h.Exec(query, values...)
		if err != nil {
			return 0, errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.E(pkg, op, errors.Database, err, "Failed to Retrieve ID")
	}

	return id, nil
}

// Delete removes an object from the database
func Delete[H Handle, S DBObject](h H, s S) (*sql.Result, error) {
	const op errors.Op = "Delete"
	var (
		res   sql.Result
		err   error
		query = ""
	)

	table := getTable(s)
	if table == "" {
		return nil, errors.E(pkg, op, errors.Database, fmt.Errorf("invalid table struct"))
	}

	params := structs.Map(s)

	id, ok := params["id"]
	if !ok {
		return nil, errors.E(pkg, op, errors.Database, fmt.Errorf("an object id is required to delete by object"))
	}

	query = fmt.Sprintf("delete from %s where id = ?", table)

	switch h := any(h).(type) {
	case *sqlx.Tx:
		stmt, err := h.Prepare(query)
		if err != nil {
			return nil, errors.E(pkg, op, errors.Database, err, "Failed to Prepare Query")
		}

		res, err = stmt.Exec(id)
		if err != nil {
			return nil, errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	case *sqlx.DB:
		res, err = h.Exec(query, id)
		if err != nil {
			return nil, errors.E(pkg, op, errors.Database, err, "Failed to Execute Query")
		}
	}

	return &res, nil
}

func arc53Database() string {
	if os.Getenv("ENV") == "production" {
		return "arc53"
	}
	return "arc53_test"
}

func ErrNoRows(err error) bool {
	return err != nil && (err == sql.ErrNoRows || err.(*errors.Error).Kind == errors.DatabaseResultNotFound)
}
