package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gorilla/mux"
	"github.com/szwedm/cloud-library/internal/storage"
)

const UUIDRegex string = `[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`

type server struct {
	router       *mux.Router
	booksHandler *booksHandler
	usersHandler *usersHandler
	authHandler  *authHandler
}

func NewServer(booksStorage storage.Books, usersStorage storage.Users) *server {
	return &server{
		router:       mux.NewRouter(),
		booksHandler: newBooksHandler(booksStorage),
		usersHandler: newUsersHandler(usersStorage),
		authHandler:  newAuthHandler(usersStorage),
	}
}

func (s *server) registerBookPaths() {
	s.router.HandleFunc("/books", s.middleware(s.booksHandler.getBooks)).Methods("GET")
	s.router.HandleFunc("/books", s.middleware(s.booksHandler.createBook)).Methods("POST")
	s.router.HandleFunc("/books/{id:"+UUIDRegex+"}", s.middleware(s.booksHandler.getBookByID)).Methods("GET")
	s.router.HandleFunc("/books/{id:"+UUIDRegex+"}", s.middleware(s.booksHandler.updateBook)).Methods("PUT")
	s.router.HandleFunc("/books/{id:"+UUIDRegex+"}", s.middleware(s.booksHandler.deleteBookByID)).Methods("DELETE")
}

func (s *server) registerUserPaths() {
	s.router.HandleFunc("/users", s.middleware(s.usersHandler.getUsers)).Methods("GET")
	s.router.HandleFunc("/users", s.usersHandler.createUser).Methods("POST")
	s.router.HandleFunc("/users/{id:"+UUIDRegex+"}", s.middleware(s.usersHandler.getUserByID)).Methods("GET")
	s.router.HandleFunc("/users/{id:"+UUIDRegex+"}", s.middleware(s.usersHandler.updateUser)).Methods("PUT")
	s.router.HandleFunc("/users/{id:"+UUIDRegex+"}", s.middleware(s.usersHandler.deleteUserByID)).Methods("DELETE")
}

func (s *server) registerAuthPaths() {
	s.router.HandleFunc("/signin", s.authHandler.signin).Methods("POST")
}

func (s *server) middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			respondWithError(w, http.StatusUnauthorized, errors.New("malformed token"))
			return
		}
		jwtToken := authHeader[1]
		token, err := jwt.Parse(jwtToken, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
			}
			return []byte(os.Getenv("APP_JWT_SIGN_KEY")), nil
		})
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), "props", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

	}
}

func (s *server) Run() {
	s.registerBookPaths()
	s.registerUserPaths()
	s.registerAuthPaths()
	log.Fatal(http.ListenAndServe(":8080", s.router))
}
