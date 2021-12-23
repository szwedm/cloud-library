package main

import (
	"fmt"

	"github.com/szwedm/cloud-library/internal/server"
	"github.com/szwedm/cloud-library/internal/storage"
)

func main() {
	fmt.Println("Let's get started!")

	cfg := storage.NewConfig()
	db := storage.NewPostgres(cfg.ConnectionString())
	defer db.CloseConnection()

	db.TestConnection()

	srv := server.NewServer(db.NewBooksStorage(), db.NewUsersStorage())
	srv.Run()
}
