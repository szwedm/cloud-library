package storage

import "github.com/szwedm/cloud-library/internal/dbmodel"

type Books interface {
	GetBooks() ([]dbmodel.BookDTO, error)
	GetBookByID(id string) (dbmodel.BookDTO, error)
	CreateBook(dto dbmodel.BookDTO) (string, error)
	UpdateBook(dto dbmodel.BookDTO) error
	DeleteBookByID(id string) error
}

type Users interface {
	GetUsers() ([]dbmodel.UserDTO, error)
	GetUserByID(id string) (dbmodel.UserDTO, error)
	CreateUser(dto dbmodel.UserDTO) (string, error)
	UpdateUser(dto dbmodel.UserDTO) error
	DeleteUserByID(id string) error
}
