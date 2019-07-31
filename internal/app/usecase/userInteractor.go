package usecase

import (
	"github.com/ramonmacias/librarium/internal/app/domain/model"

	"github.com/google/uuid"
	"github.com/ramonmacias/librarium/internal/app/domain/repository"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
)

type UserInteractor interface {
	ListUser() ([]*User, error)
	RegisterUser(email, name, lastName string) error
	RemoveUser(id string) error
	FindByID(id string) (*User, error)
}

type User struct {
	ID       string
	Email    string
	Name     string
	LastName string
}

type userInteractor struct {
	repo    repository.UserRepository
	service *service.UserService
}

func NewUserInteractor(repo repository.UserRepository, service *service.UserService) *userInteractor {
	return &userInteractor{
		repo:    repo,
		service: service,
	}
}

func (u *userInteractor) ListUser() ([]*User, error) {
	users, err := u.repo.FindAll()
	if err != nil {
		return nil, err
	}
	return toUser(users), nil
}

func (u *userInteractor) RegisterUser(email, name, lastName string) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	if err := u.service.Duplicated(email); err != nil {
		return err
	}
	user := model.NewUser(uid.String(), email, name, lastName)
	if err := u.repo.Save(user); err != nil {
		return err
	}
	return nil
}

func (u *userInteractor) RemoveUser(id string) error {
	user, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	return u.repo.Delete(user)
}

func (u *userInteractor) FindByID(id string) (*User, error) {
	user, err := u.repo.FindByID(id)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, nil
	}
	return &User{
		ID:       user.GetID(),
		Name:     user.GetName(),
		Email:    user.GetEmail(),
		LastName: user.GetLastName(),
	}, nil
}

func toUser(users []*model.User) []*User {
	res := make([]*User, len(users))
	for i, user := range users {
		res[i] = &User{
			ID:    user.GetID(),
			Email: user.GetEmail(),
		}
	}
	return res
}
