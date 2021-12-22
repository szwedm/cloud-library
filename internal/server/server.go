package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/szwedm/cloud-library/internal/storage"
)

type server struct {
	router       *mux.Router
	booksHandler *booksHandler
	usersHandler *usersHandler
}

func NewServer(booksStorage storage.Books, usersStorage storage.Users) *server {
	return &server{
		router:       mux.NewRouter(),
		booksHandler: newBooksHandler(booksStorage),
		usersHandler: newUsersHandler(usersStorage),
	}
}

func (s *server) registerBookPaths() {
	s.router.HandleFunc("/books", s.booksHandler.getBooks)
}

func (s *server) Run() {
	s.registerBookPaths()
	log.Fatal(http.ListenAndServe(":8080", s.router))
}
