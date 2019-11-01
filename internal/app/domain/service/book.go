package service

import (
	"fmt"

	"github.com/ramonmacias/librarium/internal/app/domain/repository"
)

type BookService struct {
	repo repository.BookRepository
}

func NewBookService(repo repository.BookRepository) *BookService {
	return &BookService{
		repo: repo,
	}
}

func (s *BookService) Duplicated(ISBN string) error {
	book, err := s.repo.FindByISBN(ISBN)
	if book != nil {
		return fmt.Errorf("%s already exists", ISBN)
	}
	return err
}
