package storage

import (
	"database/sql"

	"github.com/szwedm/cloud-library/internal/dbmodel"
)

const UsersTable = "users"

type users struct {
	db *sql.DB
}

func (u *users) GetUsers() ([]dbmodel.UserDTO, error) {
	stmt := "SELECT * FROM " + UsersTable
	rows, err := u.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	dtos := make([]dbmodel.UserDTO, 0)
	for rows.Next() {
		var dto dbmodel.UserDTO
		if err := rows.Scan(&dto.Id, &dto.Username, &dto.Password, &dto.Role); err != nil {
			return nil, err
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

func (u *users) GetUserByID(id string) (dbmodel.UserDTO, error) {
	stmt := "SELECT * FROM " + UsersTable + " WHERE id=$1"
	row := u.db.QueryRow(stmt, id)

	var dto dbmodel.UserDTO
	err := row.Scan(&dto.Id, &dto.Username, &dto.Password, &dto.Role)
	if err != nil {
		return dbmodel.UserDTO{}, err
	}
	return dto, nil
}

func (u *users) CreateUser(dto dbmodel.UserDTO) (string, error) {
	stmt := "INSERT INTO " + UsersTable + "(id, username, password, role) " +
		"VALUES($1, $2, $3, $4) RETURNING id"
	row := u.db.QueryRow(stmt, dto.Id, dto.Username, dto.Password, dto.Role)

	var newUserID string
	err := row.Scan(&newUserID)
	if err != nil {
		return "", err
	}
	return newUserID, nil
}

func (u *users) UpdateUser(dto dbmodel.UserDTO) error {
	stmt := "UPDATE " + UsersTable + " SET username=$1, password=$2, role=$3 WHERE id=$4"
	_, err := u.db.Exec(stmt, dto.Username, dto.Password, dto.Role, dto.Id)
	return err
}

func (u *users) DeleteUserByID(id string) error {
	stmt := "DELETE FROM " + UsersTable + " WHERE id=$1"
	_, err := u.db.Exec(stmt, id)
	return err
}
