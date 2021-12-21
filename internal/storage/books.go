package storage

import (
	"database/sql"

	"github.com/szwedm/cloud-library/internal/dbmodel"
)

const BooksTable = "books"

type books struct {
	db *sql.DB
}

func (b *books) GetBooks() ([]dbmodel.BookDTO, error) {
	stmt := "SELECT * FROM " + BooksTable
	rows, err := b.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	dtos := make([]dbmodel.BookDTO, 0)
	for rows.Next() {
		var dto dbmodel.BookDTO
		if err := rows.Scan(&dto); err != nil {
			return nil, err
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

func (b *books) GetBookByID(id string) (dbmodel.BookDTO, error) {
	stmt := "SELECT * FROM " + BooksTable + " WHERE id=$1"
	row := b.db.QueryRow(stmt, id)

	var dto dbmodel.BookDTO
	err := row.Scan(&dto)
	if err != nil {
		return dbmodel.BookDTO{}, err
	}
	return dto, nil
}

func (b *books) CreateBook(dto dbmodel.BookDTO) (string, error) {
	stmt := "INSERT INTO " + BooksTable + "(id, title, author, subject) " +
		"VALUES($1, $2, $3, $4) RETURNING id"
	row := b.db.QueryRow(stmt, dto.Id, dto.Title, dto.Author, dto.Subject)

	var newBookID string
	err := row.Scan(&newBookID)
	if err != nil {
		return "", err
	}
	return newBookID, nil
}

func (b *books) UpdateBook(dto dbmodel.BookDTO) error {
	stmt := "UPDATE " + BooksTable + " SET title=$1, author=$2, subject=$3 WHERE id=$4"
	_, err := b.db.Exec(stmt, dto.Title, dto.Author, dto.Subject, dto.Id)
	return err
}

func (b *books) DeleteBookByID(id string) error {
	stmt := "DELETE FROM " + BooksTable + " WHERE id=$1"
	_, err := b.db.Exec(stmt, id)
	return err
}
