package postgres

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/ramonmacias/librarium/internal/app/domain/model"
)

type bookController struct {
	db *gorm.DB
}

type Book struct {
	gorm.Model
	Title  string
	ISBN   string
	Price  float64
	UserID uint
}

func (b Book) GetID() string {
	return fmt.Sprint(b.ID)
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

//TODO need to be able to get this User from a connection into database
func (b Book) GetUser() *model.User {
	return nil
}

func NewBookController(db *gorm.DB) *bookController {
	return &bookController{
		db: db,
	}
}

func (r bookController) FindAll() ([]model.Book, error) {
	var fetchedBooks []Book
	if err := r.db.Find(&fetchedBooks).Error; err != nil {
		return nil, err
	}
	books := make([]model.Book, len(fetchedBooks))
	i := 0
	for _, book := range fetchedBooks {
		books[i] = book
		i++
	}
	return books, nil
	// TODO this one not works, but I think it should work I don't know why
	// return fetchedBooks, nil
}

func (r bookController) FindByID(id string) (model.Book, error) {
	return nil, nil
}

func (r bookController) FindByISBN(ISBN string) (model.Book, error) {
	return nil, nil
}

func (r bookController) Save(book model.Book) error {
	return nil
}

func (r bookController) Delete(id string) error {
	return nil
}
