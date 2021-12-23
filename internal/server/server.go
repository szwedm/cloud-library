package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/szwedm/cloud-library/internal/storage"
)

const UUIDRegex string = `^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}$`

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
	s.router.HandleFunc("/books", s.booksHandler.getBooks).Methods("GET")
	s.router.HandleFunc("/books", s.booksHandler.createBook).Methods("POST")
	s.router.HandleFunc("/books/{id:"+UUIDRegex+"}", s.booksHandler.getBookByID).Methods("GET")
	s.router.HandleFunc("/books/{id:"+UUIDRegex+"}", s.booksHandler.updateBook).Methods("PUT")
	s.router.HandleFunc("/books/{id:"+UUIDRegex+"}", s.booksHandler.deleteBookByID).Methods("DELETE")
}

func (s *server) registerUserPaths() {
	s.router.HandleFunc("/users", s.usersHandler.getUsers).Methods("GET")
	s.router.HandleFunc("/users", s.usersHandler.createUser).Methods("POST")
	s.router.HandleFunc("/users/{id:"+UUIDRegex+"}", s.usersHandler.getUserByID).Methods("GET")
	s.router.HandleFunc("/users/{id:"+UUIDRegex+"}", s.usersHandler.updateUser).Methods("PUT")
	s.router.HandleFunc("/users/{id:"+UUIDRegex+"}", s.usersHandler.deleteUserByID).Methods("DELETE")
}

func (s *server) Run() {
	s.registerBookPaths()
	s.registerUserPaths()
	log.Fatal(http.ListenAndServe(":8080", s.router))
}
