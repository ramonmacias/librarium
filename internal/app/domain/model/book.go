package model

type Book interface {
	GetID() string
	GetTitle() string
	GetISBN() string
	GetPrice() float64
}
