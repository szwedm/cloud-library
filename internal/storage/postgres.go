package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type postgres struct {
	db      *sql.DB
	connStr string
}

func NewPostgres(connStr string) *postgres {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	return &postgres{
		db:      db,
		connStr: connStr,
	}
}

func (p *postgres) TestConnection() {
	err := p.db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to the database has been established.")
}
