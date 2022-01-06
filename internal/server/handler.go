package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/szwedm/cloud-library/internal/dbmodel"
	"github.com/szwedm/cloud-library/internal/model"
	"github.com/szwedm/cloud-library/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

const MaxBookFileSize int64 = 10 << 20

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

func newUsersHandler(u storage.Users) *usersHandler {
	return &usersHandler{
		storage: u,
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
	props, _ := r.Context().Value("props").(jwt.MapClaims)
	if props["role"] != model.UserRoleAdministrator && props["role"] != model.UserRoleReader {
		respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	vars := mux.Vars(r)
	if vars["id"] == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("book id is required"))
		return
	}

	filePath := filepath.Clean(os.Getenv("APP_BOOKS_STORAGE_PATH") + string(os.PathSeparator) + vars["id"] + ".pdf")
	requestedFile, err := os.Open(filePath)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err)
		return
	}
	defer requestedFile.Close()

	fileInfo, _ := requestedFile.Stat()
	fileSize := fileInfo.Size()

	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))

	if _, err = io.Copy(w, requestedFile); err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
}

func (h *booksHandler) createBook(w http.ResponseWriter, r *http.Request) {
	props, _ := r.Context().Value("props").(jwt.MapClaims)
	if props["role"] != model.UserRoleAdministrator {
		respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxBookFileSize)
	err := r.ParseMultipartForm(MaxBookFileSize)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	id := uuid.NewString()
	dto := dbmodel.BookDTO{
		Id:      id,
		Title:   r.PostFormValue("title"),
		Author:  r.PostFormValue("author"),
		Subject: r.PostFormValue("subject"),
	}

	file, _, err := r.FormFile("bookFile")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	fileType := http.DetectContentType(buff)
	if fileType != "application/pdf" {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("only pdf files are supported, %w", err))
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	dst, err := os.Create(fmt.Sprintf("%s/%s.pdf", os.Getenv("APP_BOOKS_STORAGE_PATH"), id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	_, err = h.storage.CreateBook(dto)
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
	props, _ := r.Context().Value("props").(jwt.MapClaims)
	if props["role"] != model.UserRoleAdministrator {
		respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

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
	props, _ := r.Context().Value("props").(jwt.MapClaims)
	if props["role"] != model.UserRoleAdministrator {
		respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

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
	props, _ := r.Context().Value("props").(jwt.MapClaims)
	if props["role"] != model.UserRoleAdministrator {
		respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

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
	props, _ := r.Context().Value("props").(jwt.MapClaims)
	if props["role"] != model.UserRoleAdministrator {
		respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

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
	props, _ := r.Context().Value("props").(jwt.MapClaims)
	if props["role"] != model.UserRoleAdministrator {
		respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

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
	props, _ := r.Context().Value("props").(jwt.MapClaims)
	if props["role"] != model.UserRoleAdministrator {
		respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

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
