package model

import "github.com/szwedm/cloud-library/internal/dbmodel"

type Book struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Subject string `json:"subject"`
}

const (
	UserRoleReader        string = "reader"
	UserRoleAdministrator string = "administrator"
)

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
}

func BookFromDTO(dto dbmodel.BookDTO) (b Book) {
	b = Book{
		Id:      dto.Id,
		Title:   dto.Title,
		Author:  dto.Author,
		Subject: dto.Subject,
	}
	return
}

func DTOFromBook(book Book) (dto dbmodel.BookDTO) {
	dto = dbmodel.BookDTO{
		Id:      book.Id,
		Title:   book.Title,
		Author:  book.Author,
		Subject: book.Subject,
	}
	return
}

func UserFromDTO(dto dbmodel.UserDTO) (u User) {
	u = User{
		Id:       dto.Id,
		Username: dto.Username,
		Password: dto.Password,
		Role:     dto.Role,
	}
	return
}

func DTOFromUser(user User) (dto dbmodel.UserDTO) {
	dto = dbmodel.UserDTO{
		Id:       user.Id,
		Username: user.Username,
		Password: user.Password,
		Role:     user.Role,
	}
	return
}
