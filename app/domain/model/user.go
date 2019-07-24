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

func (u *User) GetID() string {
	return u.id
}

func (u *User) GetName() string {
	return u.name
}

func (u *User) GetEmail() string {
	return u.email
}

func (u *User) GetLastName() string {
	return u.lastName
}
