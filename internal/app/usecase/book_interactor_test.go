package usecase_test

import (
	"testing"

	"github.com/ramonmacias/librarium/internal/app/domain/model"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
	"github.com/ramonmacias/librarium/internal/app/interface/persistence/memory"
	"github.com/ramonmacias/librarium/internal/app/usecase"
)

type FakeBookModel struct {
	ID    string
	Title string
	ISBN  string
	Price float64
	User  *model.User
}

func (f FakeBookModel) GetID() string {
	return f.ID
}

func (f FakeBookModel) GetTitle() string {
	return f.Title
}

func (f FakeBookModel) GetISBN() string {
	return f.ISBN
}

func (f FakeBookModel) GetPrice() float64 {
	return f.Price
}

func (f FakeBookModel) GetUser() *model.User {
	return f.User
}

var (
	bookInteractor usecase.BookInteractor
)

func init() {
	// Here I'm broking one of the rule of clean architecture, this layer shouldn'
	// know anything about the outter layer, on this case interface layer
	bookController := memory.NewBookController()
	bookInteractor = usecase.NewBookInteractor(
		bookController,
		service.NewBookService(bookController),
	)
}

func TestEmptyListBooks(t *testing.T) {
	books, err := bookInteractor.ListBooks()
	if len(books) != 0 {
		t.Errorf("Should return an empty list, but got: %d", len(books))
	}
	if err != nil {
		t.Errorf("Should not return an error but got err: %v", err)
	}
}

func TestNotEmptyListBooks(t *testing.T) {
	books, err := bookInteractor.ListBooks()
	if len(books) != 0 {
		t.Errorf("Should return an empty list, but got: %d", len(books))
	}
	if err != nil {
		t.Errorf("Should not return an error but got err: %v", err)
	}

	err = bookInteractor.RegisterBook(FakeBookModel{
		Title: "Test Title",
		ISBN:  "testIsbn",
		Price: 34.4,
	})

	books, err = bookInteractor.ListBooks()
	if len(books) != 1 {
		t.Errorf("Should return a list with one item, but got list with %d items", len(books))
	}
	if err != nil {
		t.Errorf("Should not return an error but got err: %v", err)
	}

	RemovingBooks(t)
}

func RemovingBooks(t *testing.T) {
	books, err := bookInteractor.ListBooks()

	err = bookInteractor.RemoveBook(books[0].GetID())
	if err != nil {
		t.Errorf("Should not return an error but got err: %v", err)
	}

	books, _ = bookInteractor.ListBooks()
	if len(books) != 0 {
		t.Errorf("Should return an empty list, but got: %d", len(books))
	}
	if len(books) != 0 {
		t.Errorf("Should return an empty list, but got: %d", len(books))
	}
}

func TestFindAndUpdateBook(t *testing.T) {
	bookInteractor.RegisterBook(FakeBookModel{
		Title: "Test Title",
		ISBN:  "testIsbn",
		Price: 34.4,
	})

	books, _ := bookInteractor.ListBooks()

	if books[0].GetTitle() != "Test Title" {
		t.Errorf("Should get Test Tile but got %s", books[0].GetTitle())
	}
	if books[0].GetISBN() != "testIsbn" {
		t.Errorf("Should get testIsbn but got %s", books[0].GetISBN())
	}
	if books[0].GetPrice() != 34.4 {
		t.Errorf("Should get 34.4 but got %f", books[0].GetPrice())
	}

	err := bookInteractor.UpdateBook(FakeBookModel{
		ID:    books[0].GetID(),
		Title: "Another Test Title",
		ISBN:  "Another testISBN",
		Price: 35.5,
	})
	if err != nil {
		t.Errorf("Should not be an error but got %v", err)
	}

	book, err := bookInteractor.FindByID(books[0].GetID())
	if err != nil {
		t.Errorf("Should not be an error but got: %v", err)
	}
	if book.GetTitle() != "Another Test Title" {
		t.Errorf("Should get Test Tile but got %s", books[0].GetTitle())
	}
	if book.GetISBN() != "Another testISBN" {
		t.Errorf("Should get testIsbn but got %s", books[0].GetISBN())
	}
	if book.GetPrice() != 35.5 {
		t.Errorf("Should get 34.4 but got %f", books[0].GetPrice())
	}
}
