package catalog_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"librarium/internal/catalog"
)

func TestBuildAsset(t *testing.T) {
	testCases := map[string]struct {
		info        any
		expectedErr error
		assertAsset func(t *testing.T, asset *catalog.Asset)
	}{
		"it should build an asset if the provided info is a book": {
			info: &catalog.Book{
				Title:       "The Lord Of The Rings The Two Towers",
				Author:      "J.R.R Tolkien",
				Publisher:   "George Allen & Unwin",
				ISBN:        "978-0261102385",
				PageCount:   352,
				PublishedAt: time.Date(1954, time.November, 11, 0, 0, 0, 0, time.UTC),
			},
			assertAsset: func(t *testing.T, asset *catalog.Asset) {
				assert.NotEmpty(t, asset.ID)
				assert.NotEmpty(t, asset.CreatedAt)
				assert.NotEmpty(t, asset.UpdatedAt)
				assert.Equal(t, catalog.AssetCategoryBook, asset.Category())
				book, ok := asset.Info().(*catalog.Book)
				assert.True(t, ok)
				assert.Equal(t, "The Lord Of The Rings The Two Towers", book.Title)
			},
		},
		"it should build an asset if the provided info is a magazine": {
			info: &catalog.Magazine{
				Title:       "National Geographic",
				Issue:       "March 2022",
				Publisher:   "National Geographic Partners",
				PublishedAt: time.Date(2022, time.March, 1, 0, 0, 0, 0, time.UTC),
				PageCount:   128,
			},
			assertAsset: func(t *testing.T, asset *catalog.Asset) {
				assert.NotEmpty(t, asset.ID)
				assert.NotEmpty(t, asset.CreatedAt)
				assert.NotEmpty(t, asset.UpdatedAt)
				assert.Equal(t, catalog.AssetCategoryMagazine, asset.Category())
				mag, ok := asset.Info().(*catalog.Magazine)
				assert.True(t, ok)
				assert.Equal(t, "National Geographic", mag.Title)
			},
		},
		"it should build an asset if the provided info is a news paper": {
			info: &catalog.NewsPaper{
				Title:       "The New York Times",
				Edition:     "Sunday Edition",
				Publisher:   "The New York Times Company",
				PublishedAt: time.Date(2023, time.December, 17, 0, 0, 0, 0, time.UTC),
				PageCount:   94,
			},
			assertAsset: func(t *testing.T, asset *catalog.Asset) {
				assert.NotEmpty(t, asset.ID)
				assert.NotEmpty(t, asset.CreatedAt)
				assert.NotEmpty(t, asset.UpdatedAt)
				assert.Equal(t, catalog.AssetCategoryNewsPaper, asset.Category())
				newspaper, ok := asset.Info().(*catalog.NewsPaper)
				assert.True(t, ok)
				assert.Equal(t, "The New York Times", newspaper.Title)
			},
		},
		"it should build an asset if the provided info is a DVD": {
			info: &catalog.DVD{
				Title:       "The Lord of the Rings: The Fellowship of the Ring (Extended Edition)",
				Director:    "Peter Jackson",
				Producer:    "Barrie M. Osborne",
				DurationMin: 228,
				RegionCode:  "Region 1",
				ReleasedAt:  time.Date(2002, time.November, 12, 0, 0, 0, 0, time.UTC),
			},
			assertAsset: func(t *testing.T, asset *catalog.Asset) {
				assert.NotEmpty(t, asset.ID)
				assert.NotEmpty(t, asset.CreatedAt)
				assert.NotEmpty(t, asset.UpdatedAt)
				assert.Equal(t, catalog.AssetCategoryDVD, asset.Category())
				dvd, ok := asset.Info().(*catalog.DVD)
				assert.True(t, ok)
				assert.Equal(t, "The Lord of the Rings: The Fellowship of the Ring (Extended Edition)", dvd.Title)
			},
		},
		"it should build an asset if the provided info is a CD": {
			info: &catalog.CD{
				Title:       "Thriller",
				Artist:      "Michael Jackson",
				Label:       "Epic Records",
				TrackCount:  9,
				DurationMin: 42,
				ReleasedAt:  time.Date(1982, time.November, 30, 0, 0, 0, 0, time.UTC),
			},
			assertAsset: func(t *testing.T, asset *catalog.Asset) {
				assert.NotEmpty(t, asset.ID)
				assert.NotEmpty(t, asset.CreatedAt)
				assert.NotEmpty(t, asset.UpdatedAt)
				assert.Equal(t, catalog.AssetCategoryCD, asset.Category())
				cd, ok := asset.Info().(*catalog.CD)
				assert.True(t, ok)
				assert.Equal(t, "Thriller", cd.Title)
			},
		},
		"it should build an asset if the provided info is a video game": {
			info: &catalog.VideoGame{
				Title:      "The Last of Us",
				Developer:  "Naughty Dog",
				Platform:   "PlayStation 3",
				Genre:      "Action-adventure",
				ReleasedAt: time.Date(2013, time.June, 14, 0, 0, 0, 0, time.UTC),
				AgeRating:  "+18",
			},
			assertAsset: func(t *testing.T, asset *catalog.Asset) {
				assert.NotEmpty(t, asset.ID)
				assert.NotEmpty(t, asset.CreatedAt)
				assert.NotEmpty(t, asset.UpdatedAt)
				assert.Equal(t, catalog.AssetCategoryVideoGame, asset.Category())
				game, ok := asset.Info().(*catalog.VideoGame)
				assert.True(t, ok)
				assert.Equal(t, "The Last of Us", game.Title)
			},
		},
		"it should return an error if the provided info is not a pointer": {
			info: catalog.Book{
				Title:       "The Lord Of The Rings The Two Towers",
				Author:      "J.R.R Tolkien",
				Publisher:   "George Allen & Unwin",
				ISBN:        "978-0261102385",
				PageCount:   352,
				PublishedAt: time.Date(1954, time.November, 11, 0, 0, 0, 0, time.UTC),
			},
			expectedErr: errors.New("asset category not allowed"),
			assertAsset: func(t *testing.T, asset *catalog.Asset) {},
		},
		"it should return an error if the provided info is not in any of the ones expected": {
			info:        struct{ Title string }{Title: "Random title"},
			expectedErr: errors.New("asset category not allowed"),
			assertAsset: func(t *testing.T, asset *catalog.Asset) {},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			asset, err := catalog.BuildAsset(tc.info)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertAsset(t, asset)
		})
	}
}
