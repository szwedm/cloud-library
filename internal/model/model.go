package model

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
