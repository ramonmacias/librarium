package memory

import (
	"sync"

	"github.com/ramonmacias/librarium/internal/app/domain/model"
)

type bookController struct {
	mu    *sync.Mutex
	books map[string]model.Book
}

func NewBookController() *bookController {
	return &bookController{
		mu:    &sync.Mutex{},
		books: map[string]model.Book{},
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

	r.books[book.GetID()] = book
	return nil
}

func (r bookController) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.books, id)

	return nil
}
