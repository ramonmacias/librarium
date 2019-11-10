package service_test

import (
	"testing"

	"github.com/ramonmacias/librarium/internal/app/domain/model"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
)

type FakeBookModel struct{}

func (f FakeBookModel) GetID() string {
	return ""
}

func (f FakeBookModel) GetTitle() string {
	return ""
}

func (f FakeBookModel) GetISBN() string {
	return ""
}

func (f FakeBookModel) GetPrice() float64 {
	return 0
}

func (f FakeBookModel) GetUser() *model.User {
	return nil
}

type FakeBookRepository struct{}

func (f FakeBookRepository) FindAll() ([]model.Book, error) {
	return nil, nil
}

func (f FakeBookRepository) FindByID(id string) (model.Book, error) {
	return nil, nil
}

func (f FakeBookRepository) FindByISBN(ISBN string) (model.Book, error) {
	if ISBN == "IsbnMustExist" {
		return FakeBookModel{}, nil
	} else {
		return nil, nil
	}
}

func (f FakeBookRepository) Save(book model.Book) error {
	return nil
}

func (f FakeBookRepository) Delete(id string) error {
	return nil
}

var (
	bookService *service.BookService
)

func init() {
	bookService = service.NewBookService(&FakeBookRepository{})
}

func TestDuplicatedBook(t *testing.T) {
	var res error
	res = bookService.Duplicated("IsbnNotExists")
	if res != nil {
		t.Errorf("Duplicated should return nothing but returns %v", res)
	}
	res = bookService.Duplicated("IsbnMustExist")
	if res == nil {
		t.Error("Duplicated should returns an error but returns nothing")
	}
}
