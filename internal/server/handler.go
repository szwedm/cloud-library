package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/szwedm/cloud-library/internal/dbmodel"
	"github.com/szwedm/cloud-library/internal/model"
	"github.com/szwedm/cloud-library/internal/storage"
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

	dto := dbmodel.BookDTO{}
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
