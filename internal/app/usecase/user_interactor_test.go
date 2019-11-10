package usecase_test

import (
	"testing"

	"github.com/ramonmacias/librarium/internal/app/domain/service"
	"github.com/ramonmacias/librarium/internal/app/interface/persistence/memory"
	"github.com/ramonmacias/librarium/internal/app/usecase"
)

var (
	userInteractor usecase.UserInteractor
)

func init() {
	userController := memory.NewUserController()
	userInteractor = usecase.NewUserInteractor(
		userController,
		service.NewUserService(userController),
	)
}

func TestEmptyUserList(t *testing.T) {
	users, err := userInteractor.ListUser()
	if err != nil {
		t.Errorf("Shouldn't be an error but got %v", err)
	}
	if len(users) != 0 {
		t.Errorf("Should be an empty list but got a list with %d items", len(users))
	}
}

func TestNotEmptyBookList(t *testing.T) {
	users, err := userInteractor.ListUser()
	if err != nil {
		t.Errorf("Shouldn't be an error but got %v", err)
	}
	if len(users) != 0 {
		t.Errorf("Should be an empty list but got a list with %d items", len(users))
	}

	userInteractor.RegisterUser("test@test.com", "testName", "testLastName")
	users, err = userInteractor.ListUser()
	if err != nil {
		t.Errorf("Shouldn't be an err but got %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Should be a list with only one item but got %d items", len(users))
	}

	RemoveUser(t)
}

func RemoveUser(t *testing.T) {
	users, _ := userInteractor.ListUser()
	if len(users) != 1 {
		t.Errorf("Should be a list with only one item but got %d items", len(users))
	}
	err := userInteractor.RemoveUser(users[0].ID)
	if err != nil {
		t.Errorf("Shouldn't be an error but got %v", err)
	}

	users, _ = userInteractor.ListUser()
	if len(users) != 0 {
		t.Errorf("After remove the user the list should be empty but got %d items", len(users))
	}
}

func TestFindUser(t *testing.T) {
	userInteractor.RegisterUser("test@test.com", "testName", "testLastName")
	users, _ := userInteractor.ListUser()
	user, err := userInteractor.FindByID(users[0].ID)
	if err != nil {
		t.Errorf("Shouldn't be an error but got %v", err)
	}
	if user.Name != "testName" {
		t.Errorf("The name should be testName but got %s", user.Name)
	}
	if user.Email != "test@test.com" {
		t.Errorf("The email should be test@test.com but got %s", user.Email)
	}
	if user.LastName != "testLastName" {
		t.Errorf("The LastName should be testLastName but got %s", user.LastName)
	}

	user, err = userInteractor.FindByID("noUserID")
	if user != nil || err != nil {
		t.Errorf("No user should return a user an error nil but got user %v err %v", user, err)
	}
}
