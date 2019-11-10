package service_test

import (
	"testing"

	"github.com/ramonmacias/librarium/internal/app/domain/model"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
)

type FakeUserRepository struct{}

func (f FakeUserRepository) FindAll() ([]*model.User, error) {
	return nil, nil
}

func (f FakeUserRepository) FindByEmail(email string) (*model.User, error) {
	if email == "email_already_in_the_system@test.com" {
		return &model.User{}, nil
	}
	return nil, nil
}

func (f FakeUserRepository) FindByID(id string) (*model.User, error) {
	return nil, nil
}

func (f FakeUserRepository) Save(*model.User) error {
	return nil
}

func (f FakeUserRepository) Delete(*model.User) error {
	return nil
}

var (
	userService *service.UserService
)

func init() {
	userService = service.NewUserService(FakeUserRepository{})
}

func TestDuplicatedUser(t *testing.T) {
	var err error

	err = userService.Duplicated("email_not_exists@test.com")
	if err != nil {
		t.Errorf("Err should be nil, but we got err: %v", err)
	}

	err = userService.Duplicated("email_already_in_the_system@test.com")
	if err == nil {
		t.Error("Err shouldn't be nil")
	}
}
