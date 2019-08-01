package postgres

import (
	"ramonmacias/samples/cleanArch/app/domain/model"

	"github.com/jinzhu/gorm"
)

type userController struct {
	db *gorm.DB
}

type User struct {
	gorm.Model
	Email    string
	Name     string
	LastName string
}

func NewUserController() *userController {
	return &userController{}
}

func (r userController) FindAll() ([]*model.User, error) {
	return nil, nil
}

func (r userController) FindByEmail(email string) (*model.User, error) {
	return nil, nil
}

func (r userController) FindByID(id string) (*model.User, error) {
	return nil, nil
}

func (r userController) Save(user *model.User) error {
	return nil
}

func (r userController) Delete(user *model.User) error {
	return nil
}
