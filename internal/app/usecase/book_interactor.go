package usecase

import (
	"github.com/ramonmacias/librarium/internal/app/domain/model"
	"github.com/ramonmacias/librarium/internal/app/domain/repository"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
)

type BookInteractor interface {
	ListBooks() ([]*model.Book, error)
	RegisterBook(book model.Book) error
	UpdateBook(book model.Book) error
	RemoveBook(id string) error
	FindByID(id string) (*model.Book, error)
}

type bookInteractor struct {
	repo    repository.BookRepository
	service *service.BookService
}

func NewBookInteractor(repo repository.BookRepository, service *service.BookService) *bookInteractor {
	return &bookInteractor{
		repo:    repo,
		service: service,
	}
}

func (b *bookInteractor) ListBooks() ([]*model.Book, error) {
	books, err := b.repo.FindAll()
	if err != nil {
		return nil, err
	}
	return books, nil
}

func (b *bookInteractor) RegisterBook(book model.Book) error {
	if err := b.service.Duplicated(book.GetISBN()); err != nil {
		return err
	}
	return b.repo.Save(&book)
}

func (b *bookInteractor) UpdateBook(book model.Book) error {
	return b.repo.Save(&book)
}

func (b *bookInteractor) RemoveBook(id string) error {
	return b.repo.Delete(id)
}

func (b *bookInteractor) FindByID(id string) (*model.Book, error) {
	return b.repo.FindByID(id)
}
