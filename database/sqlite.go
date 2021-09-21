package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type SQLite struct {
	FileName string
	Database *sql.DB
}

func (db *SQLite) Update(q *Update) {
	protoQuery, columnOrder := db.GenerateQuery(q)
	values := make([]interface{}, len(columnOrder))
	for i, v := range columnOrder {
		if i == 0 {
			values[i] = q.Update
		} else {
			values[i] = q.GetValues()[v]
		}
	}
	tx, err := db.GetDatabaseReference().Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare(protoQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	stmt.Exec(values...)
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (db *SQLite) GetFileName() string {
	return db.FileName
}

func (db *SQLite) GetDatabaseReference() *sql.DB {
	return db.Database
}

func (db *SQLite) CloseDatabaseReference() {
	db.GetDatabaseReference().Close()
	db.Database = nil
}

func (db *SQLite) SetDatabaseReference(dbPath string) {
	database := GetDatabaseForFile(dbPath)
	db.FileName = dbPath
	db.Database = database
}

func (db SQLite) GetPlaceholderForDatabaseType() string {
	return "?"
}

func (db *SQLite) GenerateQuery(u *Update) (string, []string) {
	var (
		query         string
		querySkeleton string
		valueOrder    []string
	)

	placeholder := db.GetPlaceholderForDatabaseType()

	querySkeleton = fmt.Sprintf("UPDATE %s"+
		" SET %s=%s ", u.TableName, u.Column, placeholder)
	valueOrder = append(valueOrder, u.Column)

	whereBuilder := strings.Builder{}
	whereBuilder.WriteString(" WHERE ")
	uLen := len(u.GetValues())
	i := 0
	for k := range u.GetValues() { // keep track of order since maps aren't deterministic
		assertion := fmt.Sprintf("%s=%s ", k, placeholder)
		valueOrder = append(valueOrder, k)
		whereBuilder.WriteString(assertion)
		if uLen > 1 && i < uLen-1 {
			whereBuilder.WriteString("AND ")
		}
		i++
	}
	query = querySkeleton + strings.TrimSpace(whereBuilder.String()) + ";"
	return query, valueOrder
}