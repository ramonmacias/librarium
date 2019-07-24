package memory

import (
	"fmt"
	"sync"

	"github.com/ramonmacias/librarium/internal/app/domain/model"
)

type userRepository struct {
	mu    *sync.Mutex
	users map[string]*User
}

type User struct {
	ID       string
	Email    string
	Name     string
	LastName string
}

func NewUserRepository() *userRepository {
	return &userRepository{
		mu:    &sync.Mutex{},
		users: map[string]*User{},
	}
}

func (r *userRepository) FindAll() ([]*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	users := make([]*model.User, len(r.users))
	i := 0
	for _, user := range r.users {
		users[i] = model.NewUser(user.ID, user.Email, user.Name, user.LastName)
		i++
	}
	return users, nil
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range r.users {
		if user.Email == email {
			return model.NewUser(user.ID, user.Email, user.Name, user.LastName), nil
		}
	}
	return nil, nil
}

func (r *userRepository) FindByID(id string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("User with id: %s not found", id)
	}
	return model.NewUser(user.ID, user.Email, user.Name, user.LastName), nil
}

func (r *userRepository) Save(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.GetID()] = &User{
		ID:       user.GetID(),
		Email:    user.GetEmail(),
		Name:     user.GetName(),
		LastName: user.GetLastName(),
	}
	return nil
}

func (r *userRepository) Delete(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.users, user.GetID())

	return nil
}
