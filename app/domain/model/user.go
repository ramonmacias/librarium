package model

type User struct {
	id       string
	email    string
	name     string
	lastName string
}

func NewUser(id, email, name, lastName string) *User {
	return &User{
		id:       id,
		email:    email,
		name:     name,
		lastName: lastName,
	}
}
