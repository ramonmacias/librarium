package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ramonmacias/librarium/internal/app/domain/model"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
	"github.com/ramonmacias/librarium/internal/app/interface/persistance/memory"
	"github.com/ramonmacias/librarium/internal/app/usecase"
)

type BookRequestBody struct {
	Title string  `json:"title"`
	ISBN  string  `json:"isbn"`
	Price float64 `json:"price"`
}

//TODO Thing more about this, it makes no sense
func (b BookRequestBody) GetID() string {
	return ""
}

func (b BookRequestBody) GetTitle() string {
	return b.Title
}

func (b BookRequestBody) GetISBN() string {
	return b.ISBN
}

func (b BookRequestBody) GetPrice() float64 {
	return b.Price
}

//TODO Thing more about this, it makes no sense
func (b BookRequestBody) GetUser() *model.User {
	return nil
}

var (
	memoryBookInteractor usecase.BookInteractor
)

func init() {
	memoryBookInteractor = usecase.NewBookInteractor(
		*memory.NewBookController(),
		service.NewBookService(memory.NewBookController()),
	)
}

func ListAllBooks(w http.ResponseWriter, r *http.Request) {
	books, err := memoryBookInteractor.ListBooks()
	if err != nil {
		log.Printf("Error while try to find all the books: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	bookRequest := &BookRequestBody{}
	json.NewDecoder(r.Body).Decode(bookRequest)
	defer r.Body.Close()

	err := memoryBookInteractor.RegisterBook(bookRequest)
	if err != nil {
		log.Printf("Error while try to register a new book: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func RemoveBook(w http.ResponseWriter, r *http.Request) {
}

func FindBookByID(w http.ResponseWriter, r *http.Request) {
}
