package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/ramonmacias/librarium/internal/app/domain/model"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
	"github.com/ramonmacias/librarium/internal/app/interface/persistence/memory"
	"github.com/ramonmacias/librarium/internal/app/interface/persistence/postgres"
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
	memoryBookInteractor   usecase.BookInteractor
	postgresBookInteractor usecase.BookInteractor
)

func init() {
	memoryBookInteractor = usecase.NewBookInteractor(
		*memory.NewBookController(),
		service.NewBookService(memory.NewBookController()),
	)
	db := postgres.NewClient(os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_DATABASE"), os.Getenv("POSTGRES_PASSWORD")).Connect().DB()
	postgresBookInteractor = usecase.NewBookInteractor(
		*postgres.NewBookController(db),
		service.NewBookService(postgres.NewBookController(db)),
	)
}

func ListAllBooks(w http.ResponseWriter, r *http.Request) {
	var err error
	var books []model.Book

	switch r.Header.Get(customPersistenceHeader) {
	case "memory":
		books, err = memoryBookInteractor.ListBooks()
	case "postgres":
		books, err = postgresBookInteractor.ListBooks()
	default:
		err = fmt.Errorf("Persistence type not available")
	}

	if err != nil {
		log.Printf("Error while try to find all the books: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	var err error
	bookRequest := &BookRequestBody{}
	json.NewDecoder(r.Body).Decode(bookRequest)
	defer r.Body.Close()

	switch r.Header.Get(customPersistenceHeader) {
	case "memory":
		err = memoryBookInteractor.RegisterBook(bookRequest)
	case "postgres":
		err = postgresBookInteractor.RegisterBook(bookRequest)
	default:
		err = fmt.Errorf("Persistence type not available")
	}
	if err != nil {
		log.Printf("Error while try to register a new book: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func RemoveBook(w http.ResponseWriter, r *http.Request) {
	var err error

	switch r.Header.Get(customPersistenceHeader) {
	case "memory":
		err = memoryBookInteractor.RemoveBook(mux.Vars(r)["id"])
	case "postgres":
		err = postgresBookInteractor.RemoveBook(mux.Vars(r)["id"])
	default:
		err = fmt.Errorf("Persistence type not available")
	}

	if err != nil {
		log.Printf("Error while try to remove a book: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func FindBookByID(w http.ResponseWriter, r *http.Request) {
	var err error
	var book model.Book

	switch r.Header.Get(customPersistenceHeader) {
	case "memory":
		book, err = memoryBookInteractor.FindByID(mux.Vars(r)["id"])
	case "postgres":
		book, err = postgresBookInteractor.FindByID(mux.Vars(r)["id"])
	default:
		err = fmt.Errorf("Persistence type not available")
	}

	if err != nil {
		log.Printf("Error trying to find a book: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if book == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}
