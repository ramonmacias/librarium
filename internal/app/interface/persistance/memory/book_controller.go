package memory

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ramonmacias/librarium/internal/app/domain/model"
)

type Book struct {
	ID    string
	Title string
	ISBN  string
	Price float64
	User  *model.User
}

func (b Book) GetID() string {
	return b.ID
}

func (b Book) GetTitle() string {
	return b.Title
}

func (b Book) GetISBN() string {
	return b.ISBN
}

func (b Book) GetPrice() float64 {
	return b.Price
}

func (b Book) GetUser() *model.User {
	return b.User
}

type bookController struct {
	mu    *sync.Mutex
	books map[string]Book
}

func NewBookController() *bookController {
	return &bookController{
		mu:    &sync.Mutex{},
		books: map[string]Book{},
	}
}

func (r bookController) FindAll() ([]model.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	books := make([]model.Book, len(r.books))
	i := 0
	for _, book := range r.books {
		books[i] = book
		i++
	}
	return books, nil
}

func (r bookController) FindByID(id string) (model.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	book, ok := r.books[id]
	if !ok {
		return nil, nil
	}
	return book, nil
}

func (r bookController) FindByISBN(ISBN string) (model.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, book := range r.books {
		if book.GetISBN() == ISBN {
			return book, nil
		}
	}
	return nil, nil
}

func (r bookController) Save(book model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if book.GetID() != "" {
		r.books[book.GetID()] = Book{
			ID:    book.GetID(),
			Title: book.GetTitle(),
			ISBN:  book.GetISBN(),
			Price: book.GetPrice(),
			User:  book.GetUser(),
		}
	} else {
		uid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		r.books[uid.String()] = Book{
			ID:    uid.String(),
			Title: book.GetTitle(),
			ISBN:  book.GetISBN(),
			Price: book.GetPrice(),
			User:  book.GetUser(),
		}
	}

	return nil
}

func (r bookController) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.books, id)

	return nil
}
