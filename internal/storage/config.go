package storage

import (
	"fmt"
	"os"
)

type config struct {
	dbname   string
	user     string
	password string
	host     string
	port     string
}

func NewConfig() *config {
	return &config{
		dbname:   os.Getenv("APP_DB_NAME"),
		user:     os.Getenv("APP_DB_USER"),
		password: os.Getenv("APP_DB_PASSWORD"),
		host:     os.Getenv("APP_DB_HOST"),
		port:     os.Getenv("APP_DB_PORT"),
	}
}

func (c *config) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		c.host, c.port, c.dbname, c.user, c.password)
}
