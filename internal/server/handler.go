package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/szwedm/cloud-library/internal/dbmodel"
	"github.com/szwedm/cloud-library/internal/model"
	"github.com/szwedm/cloud-library/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type booksHandler struct {
	storage storage.Books
}

type usersHandler struct {
	storage storage.Users
}

func newBooksHandler(b storage.Books) *booksHandler {
	return &booksHandler{
		storage: b,
	}
}

func newUsersHandler(b storage.Users) *usersHandler {
	return &usersHandler{
		storage: b,
	}
}

func (h *booksHandler) getBooks(w http.ResponseWriter, r *http.Request) {
	dtos, err := h.storage.GetBooks()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	books := make([]model.Book, 0)
	for _, dto := range dtos {
		books = append(books, model.BookFromDTO(dto))
	}

	body, err := json.Marshal(books)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, body)
}

func (h *booksHandler) getBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("book id is required"))
		return
	}

	dto, err := h.storage.GetBookByID(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	book := model.BookFromDTO(dto)

	body, err := json.Marshal(book)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, body)
}

func (h *booksHandler) createBook(w http.ResponseWriter, r *http.Request) {
	id := uuid.NewString()

	var book model.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err)
		r.Body.Close()
		return
	}
	defer r.Body.Close()

	book.Id = id

	dto := model.DTOFromBook(book)
	id, err := h.storage.CreateBook(dto)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	type response struct {
		Msg string `json:"message"`
	}
	resp := response{Msg: "book created with id: " + id}

	body, err := json.Marshal(resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, body)
}

func (h *booksHandler) updateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("book id is required"))
		return
	}

	var book model.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err)
		r.Body.Close()
		return
	}
	defer r.Body.Close()

	dto := dbmodel.BookDTO{
		Id: vars["id"],
	}
	if book.Title != "" {
		dto.Title = book.Title
	}
	if book.Author != "" {
		dto.Author = book.Author
	}
	if book.Subject != "" {
		dto.Subject = book.Subject
	}

	err := h.storage.UpdateBook(dto)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	type response struct {
		Msg string `json:"message"`
	}
	resp := response{Msg: "book updated"}

	body, err := json.Marshal(resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, body)
}

func (h *booksHandler) deleteBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("book id is required"))
		return
	}

	err := h.storage.DeleteBookByID(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	type response struct {
		Msg string `json:"message"`
	}
	resp := response{Msg: "book deleted"}

	body, err := json.Marshal(resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, body)
}

func (h *usersHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	dtos, err := h.storage.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	users := make([]model.User, 0)
	for _, dto := range dtos {
		users = append(users, model.UserFromDTO(dto))
	}

	body, err := json.Marshal(users)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, body)
}

func (h *usersHandler) getUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}

	dto, err := h.storage.GetUserByID(vars["id"])
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, fmt.Errorf("user with id: %s not found, %w", vars["id"], err))
			return
		}
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	user := model.UserFromDTO(dto)

	body, err := json.Marshal(user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, body)
}

func (h *usersHandler) createUser(w http.ResponseWriter, r *http.Request) {
	id := uuid.NewString()

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err)
		r.Body.Close()
		return
	}
	defer r.Body.Close()

	if user.Role != model.UserRoleAdministrator && user.Role != model.UserRoleReader {
		respondWithError(w, http.StatusBadRequest, errors.New("wrong user role"))
		return
	}

	if _, err := h.storage.GetUserByUsername(user.Username); err == nil {
		respondWithError(w, http.StatusConflict, errors.New("username already exists"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	user.Id = id
	user.Password = string(hashedPassword)

	dto := model.DTOFromUser(user)
	id, err = h.storage.CreateUser(dto)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	type response struct {
		Msg string `json:"message"`
	}
	resp := response{Msg: "user created with id: " + id}

	body, err := json.Marshal(resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, body)
}

func (h *usersHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		r.Body.Close()
		return
	}
	defer r.Body.Close()

	dto, err := h.storage.GetUserByID(vars["id"])
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, fmt.Errorf("user with id: %s not found, %w", vars["id"], err))
			return
		}
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if user.Username != "" {
		if _, err := h.storage.GetUserByUsername(user.Username); err == nil {
			respondWithError(w, http.StatusConflict, errors.New("username already exists"))
			return
		}
		dto.Username = user.Username
	}
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}
		dto.Password = string(hashedPassword)
	}
	if user.Role != "" {
		dto.Role = user.Role
	}

	err = h.storage.UpdateUser(dto)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	type response struct {
		Msg string `json:"message"`
	}
	resp := response{Msg: "user updated"}

	body, err := json.Marshal(resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, body)
}

func (h *usersHandler) deleteUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}

	err := h.storage.DeleteUserByID(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	type response struct {
		Msg string `json:"message"`
	}
	resp := response{Msg: "user deleted"}

	body, err := json.Marshal(resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, body)
}
