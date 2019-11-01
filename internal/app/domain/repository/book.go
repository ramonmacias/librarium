package repository

import "github.com/ramonmacias/librarium/internal/app/domain/model"

type BookRepository interface {
	FindAll() ([]*model.Book, error)
	FindByID() (*model.Book, error)
	FindByISBN() (*model.Book, error)
	Save(*model.Book) error
	Delete(*model.Book) error
}
