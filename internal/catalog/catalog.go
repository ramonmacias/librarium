// Package catalog provides types and functions to manage the library's item catalog.
//
// It defines a generic Asset model that can represent a variety of media types
// available in the library, including books, magazines, newspapers, DVDs, CDs,
// and video games. Each specific media type has its own struct representation.
//
// The package includes:
//   - The Asset struct: a wrapper that categorizes and holds any catalog item.
//   - Media-specific types: Book, Magazine, NewsPaper, DVD, CD, and VideoGame.
//   - AssetCategory enum for classifying catalog items.
//   - A BuildAsset factory function that instantiates an Asset from specific media data.
//
// This abstraction allows the system to treat different types of catalog items
// uniformly while preserving their specific metadata.
package catalog

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// AssetCategory defines the different types of asset that we might have
// in the library's catalog
type AssetCategory string

const (
	// AssetCategoryBook identifies the asset type of a book
	AssetCategoryBook AssetCategory = "BOOK"
	// AssetCategoryMagazine identifies the asset type of a magazine
	AssetCategoryMagazine AssetCategory = "MAGAZINE"
	// AssetCategoryNewsPaper identifies the asset type of a news paper
	AssetCategoryNewsPaper AssetCategory = "NEWS_PAPER"
	// AssetCategoryDVD identifies the asset type of a dvd
	AssetCategoryDVD AssetCategory = "DVD"
	// AssetCategoryCD identifies the asset type of a cd
	AssetCategoryCD AssetCategory = "CD"
	// AssetCategoryVideoGame identifies the asset type of a video game
	AssetCategoryVideoGame AssetCategory = "VIDEO_GAME"
)

// Asset represents a generic item in the library catalog.
// It can hold any type of media such as books, magazines, DVDs, etc.
type Asset struct {
	ID        uuid.UUID     // Unique identifier for the asset
	CreatedAt time.Time     // Timestamp of when the asset was created
	UpdatedAt time.Time     // Timestamp of the last update to the asset
	category  AssetCategory // Classification of the asset (e.g., Book, DVD)
	info      any           // Holds the concrete asset data (e.g., a Book struct)
}

func (a *Asset) Category() AssetCategory {
	return a.category
}

func (a *Asset) Info() any {
	return a.info
}

// BuildAsset creates a new library catalog asset using the provided
// info to specify the concret asset.
// If the info provided doesn't belong to any of the accepted asset category
// the function will return an error.
// If success it will return a new pointer to the Asset created and no error.
// This method only accept pointers to type, otherwise it will return an error.
func BuildAsset(info any) (*Asset, error) {
	a := &Asset{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	switch v := info.(type) {
	case *Book:
		a.category = AssetCategoryBook
		a.info = v
	case *Magazine:
		a.category = AssetCategoryMagazine
		a.info = v
	case *NewsPaper:
		a.category = AssetCategoryNewsPaper
		a.info = v
	case *DVD:
		a.category = AssetCategoryDVD
		a.info = v
	case *CD:
		a.category = AssetCategoryCD
		a.info = v
	case *VideoGame:
		a.category = AssetCategoryVideoGame
		a.info = v
	default:
		return nil, errors.New("asset category not allowed")
	}
	return a, nil
}

// Book represents a book available in the library catalog.
type Book struct {
	Title       string    // Title of the book
	Author      string    // Author of the book
	Publisher   string    // Publisher of the book
	ISBN        string    // International Standard Book Number
	PageCount   int       // Number of pages in the book
	PublishedAt time.Time // Date the book was published
}

// Magazine represents a magazine issue in the library catalog.
type Magazine struct {
	Title       string    // Title of the magazine
	Issue       string    // Specific issue identifier (e.g., "May 2025")
	Publisher   string    // Publisher of the magazine
	PublishedAt time.Time // Date the magazine was published
	PageCount   int       // Number of pages in the magazine
}

// NewsPaper represents a newspaper edition in the library catalog.
type NewsPaper struct {
	Title       string    // Title of the newspaper
	Edition     string    // Specific edition identifier (e.g., "Morning Edition")
	Publisher   string    // Publisher of the newspaper
	PublishedAt time.Time // Date the newspaper was published
	PageCount   int       // Number of pages in the newspaper
}

// DVD represents a digital video disc in the library catalog.
type DVD struct {
	Title       string    // Title of the DVD
	Director    string    // Director of the film
	Producer    string    // Producer of the film
	DurationMin int       // Duration of the film in minutes
	RegionCode  string    // DVD region code (e.g., "Region 1")
	ReleasedAt  time.Time // Date the DVD was released
}

// CD represents a compact disc in the library catalog.
type CD struct {
	Title       string    // Title of the album or CD
	Artist      string    // Main performing artist or group
	Label       string    // Record label
	TrackCount  int       // Number of tracks on the CD
	DurationMin int       // Total duration in minutes
	ReleasedAt  time.Time // Date the CD was released
}

// VideoGame represents a video game item in the library catalog.
type VideoGame struct {
	Title      string    // Title of the video game
	Developer  string    // Company or person that developed the game
	Platform   string    // Platform the game runs on (e.g., "PlayStation", "PC")
	Genre      string    // Genre of the game (e.g., "Action", "RPG")
	ReleasedAt time.Time // Date the game was released
	AgeRating  string    // Age rating (e.g., "E", "T", "M")
}

// Repository defines all the interactions between the catalog domain and the persistence layer
type Repository interface {
	// CreateAsset inserts the provided asset into the database.
	// It returns an error in case of failure.
	CreateAsset(asset *Asset) error
	// FindAssets looks for the assets already inserted in the database.
	// Returns an empty slice and no error in case of no asset found.
	// It returns an error if something fails.
	FindAssets() ([]*Asset, error)
}
