package db

import (
	"database/sql"
	"reflect"

	_ "github.com/mattn/go-sqlite3"

	"github.com/u6du/config"
	"github.com/u6du/ex"
)

const DriverName = "sqlite3"

type db string

func (d db) FileName() string {
	return string(d) + "." + DriverName
}
func (d db) Conn() *sql.DB {
	dbPath := config.File.Path(d.FileName())

	db, err := sql.Open(DriverName, dbPath)
	ex.Panic(err)
	return db
}

func (d db) With(f func(conn *sql.DB)) {
	conn := d.Conn()
	defer ex.Close(conn)
	f(conn)
}

func (d db) Exec(query string, args ...interface{}) sql.Result {
	conn := d.Conn()
	defer ex.Close(conn)
	r, err := conn.Exec(query, args...)
	ex.Warn(err)
	return r
}

// args = insert sql , insert sql args ...
func Db(name, create string, args ...interface{}) db {
	d := db(name)
	dbPath, isNew := config.File.PathIsNew(d.FileName())

	db, err := sql.Open(DriverName, dbPath)
	ex.Panic(err)

	if isNew {
		_, err := db.Exec(create)
		ex.Panic(err)

		argsLen := len(args)
		if argsLen > 0 {
			insert := args[0].(string)

			if argsLen > 1 {
				s, err := db.Prepare(insert)
				ex.Panic(err)

				for _, i := range args[1:] {
					t := reflect.TypeOf(i)
					switch t.Kind() {
					case reflect.Interface:
						li, _ := i.([]interface{})
						_, err = s.Exec(li...)
					default:
						_, err = s.Exec(i)
					}
					ex.Panic(err)

				}
			} else {
				_, err := db.Exec(insert)
				ex.Panic(err)
			}
		}
	}

	return d
}
