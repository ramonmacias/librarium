package postgres

import (
	"log"
	"strconv"

	"github.com/ramonmacias/librarium/internal/app/domain/model"

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

func NewUserController(db *gorm.DB) *userController {
	return &userController{
		db: db,
	}
}

func (r userController) FindAll() ([]*model.User, error) {
	var fetchedUsers []User
	if err := r.db.Find(&fetchedUsers).Error; err != nil {
		return nil, err
	}
	users := make([]*model.User, len(fetchedUsers))
	i := 0
	for _, user := range fetchedUsers {
		users[i] = model.NewUser(string(user.ID), user.Email, user.Name, user.LastName)
		i++
	}
	return users, nil
}

func (r userController) FindByEmail(email string) (*model.User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return model.NewUser(string(user.ID), user.Email, user.Name, user.LastName), nil
}

func (r userController) FindByID(id string) (*model.User, error) {
	log.Printf("Finding a user by ID: %s", id)
	var user User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return model.NewUser(strconv.FormatUint(uint64(user.ID), 10), user.Email, user.Name, user.LastName), nil
}

func (r userController) Save(user *model.User) error {
	log.Println("Save method postgres")
	return r.db.Save(&User{
		Email:    user.GetEmail(),
		Name:     user.GetName(),
		LastName: user.GetLastName(),
	}).Error
}

func (r userController) Delete(user *model.User) error {
	log.Printf("User ID: %s", user.GetID())
	return r.db.Where("id = ?", user.GetID()).Delete(&User{}).Error
}
