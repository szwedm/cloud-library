package server

import (
	"encoding/json"
	"net/http"

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
