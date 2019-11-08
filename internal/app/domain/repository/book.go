package repository

import "github.com/ramonmacias/librarium/internal/app/domain/model"

type BookRepository interface {
	FindAll() ([]model.Book, error)
	FindByID(id string) (model.Book, error)
	FindByISBN(ISBN string) (model.Book, error)
	Save(book model.Book) error
	Delete(id string) error
}
