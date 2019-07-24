package usecase

import (
	"github.com/ramonmacias/librarium/internal/app/domain/model"

	"github.com/google/uuid"
	"github.com/ramonmacias/librarium/internal/app/domain/repository"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
)

type UserUsecase interface {
	ListUser() ([]User, error)
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

type userUsecase struct {
	repo    repository.UserRepository
	service *service.UserService
}

func NewUserUsecase(repo repository.UserRepository, service *service.UserService) *userUsecase {
	return &userUsecase{
		repo:    repo,
		service: service,
	}
}

func (u *userUsecase) ListUser() ([]*User, error) {
	users, err := u.repo.FindAll()
	if err != nil {
		return nil, err
	}
	return toUser(users), nil
}

func (u *userUsecase) RegisterUser(email, name, lastName string) error {
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

func (u *userUsecase) RemoveUser(id string) error {
	return nil
}

func (u *userUsecase) FindByID(id string) (*User, error) {
	return nil, nil
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
