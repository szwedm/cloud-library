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
	s.router.HandleFunc("/books", s.corsMiddleware(s.middleware(s.booksHandler.getBooks))).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/books", s.corsMiddleware(s.middleware(s.booksHandler.createBook))).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/books/{id:"+UUIDRegex+"}", s.corsMiddleware(s.middleware(s.booksHandler.getBookByID))).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/books/{id:"+UUIDRegex+"}", s.corsMiddleware(s.middleware(s.booksHandler.updateBook))).Methods("PUT", "OPTIONS")
	s.router.HandleFunc("/books/{id:"+UUIDRegex+"}", s.corsMiddleware(s.middleware(s.booksHandler.deleteBookByID))).Methods("DELETE", "OPTIONS")
}

func (s *server) registerUserPaths() {
	s.router.HandleFunc("/users", s.corsMiddleware(s.middleware(s.usersHandler.getUsers))).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/users", s.corsMiddleware(s.usersHandler.createUser)).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/users/{id:"+UUIDRegex+"}", s.corsMiddleware(s.middleware(s.usersHandler.getUserByID))).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/users/{id:"+UUIDRegex+"}", s.corsMiddleware(s.middleware(s.usersHandler.updateUser))).Methods("PUT", "OPTIONS")
	s.router.HandleFunc("/users/{id:"+UUIDRegex+"}", s.corsMiddleware(s.middleware(s.usersHandler.deleteUserByID))).Methods("DELETE", "OPTIONS")
}

func (s *server) registerAuthPaths() {
	s.router.HandleFunc("/signin", s.corsMiddleware(s.authHandler.signin)).Methods("POST", "OPTIONS")
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

func (s *server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (s *server) Run() {
	s.registerBookPaths()
	s.registerUserPaths()
	s.registerAuthPaths()
	log.Fatal(http.ListenAndServe(":8080", s.router))
}
