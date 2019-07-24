package repository

import "github.com/ramonmacias/librarium/app/domain/model"

type UserRepository interface {
	FindAll() ([]*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByID(id string) (*model.User, error)
	Save(*model.User) error
	Delete(*model.User) error
}
