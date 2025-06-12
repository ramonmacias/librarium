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
	"encoding/json"
	"errors"
	"fmt"
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
	Category  AssetCategory // Classification of the asset (e.g., Book, DVD)
	Info      any           // Holds the concrete asset data (e.g., a Book struct)
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
		a.Category = AssetCategoryBook
		a.Info = v
	case *Magazine:
		a.Category = AssetCategoryMagazine
		a.Info = v
	case *NewsPaper:
		a.Category = AssetCategoryNewsPaper
		a.Info = v
	case *DVD:
		a.Category = AssetCategoryDVD
		a.Info = v
	case *CD:
		a.Category = AssetCategoryCD
		a.Info = v
	case *VideoGame:
		a.Category = AssetCategoryVideoGame
		a.Info = v
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
	// DeleteAsset removes the asset linked to the given ID from the catalog.
	// It returns an error in case of failure.
	DeleteAsset(id *Asset) error
	// GetAsset retrieves the asset linked to the given ID.
	// It returns nil, nil in case the asset cannot be found.
	// It returns an error in case of failure.
	GetAsset(id uuid.UUID) (*Asset, error)
	// FindAssets looks for the assets already inserted in the database.
	// Returns an empty slice and no error in case of no asset found.
	// It returns an error if something fails.
	FindAssets() ([]*Asset, error)
}

// CreateAssetRequest decodes
type CreateAssetRequest struct {
	Category AssetCategory   `json:"category"` // e.g., "book", "magazine", etc.
	Data     json.RawMessage `json:"data"`     // Raw JSON to decode later
	Asset    any             `json:"-"`        // Will hold the decoded asset after unmarshall
}

// UnmarshalJSON overrides the json unmarshaller so we can handle the dynamic type conversion
// of assets in the json layer.
func (r *CreateAssetRequest) UnmarshalJSON(data []byte) error {
	// Alias to avoid recursion
	type Alias CreateAssetRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch r.Category {
	case AssetCategoryBook:
		var book Book
		if err := json.Unmarshal(r.Data, &book); err != nil {
			return err
		}
		r.Asset = book
	case AssetCategoryMagazine:
		var mag Magazine
		if err := json.Unmarshal(r.Data, &mag); err != nil {
			return err
		}
		r.Asset = mag
	case AssetCategoryNewsPaper:
		var news NewsPaper
		if err := json.Unmarshal(r.Data, &news); err != nil {
			return err
		}
		r.Asset = news
	case AssetCategoryDVD:
		var dvd DVD
		if err := json.Unmarshal(r.Data, &dvd); err != nil {
			return err
		}
		r.Asset = dvd
	case AssetCategoryCD:
		var cd CD
		if err := json.Unmarshal(r.Data, &cd); err != nil {
			return err
		}
		r.Asset = cd
	case AssetCategoryVideoGame:
		var game VideoGame
		if err := json.Unmarshal(r.Data, &game); err != nil {
			return err
		}
		r.Asset = game
	default:
		return fmt.Errorf("unknown category: %s", r.Category)
	}

	return nil
}
