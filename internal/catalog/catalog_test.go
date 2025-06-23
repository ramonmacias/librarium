package catalog_test

import (
	"errors"
	"fmt"
	"os"
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
				assert.Equal(t, catalog.AssetCategoryBook, asset.Category)
				book, ok := asset.Info.(*catalog.Book)
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
				assert.Equal(t, catalog.AssetCategoryMagazine, asset.Category)
				mag, ok := asset.Info.(*catalog.Magazine)
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
				assert.Equal(t, catalog.AssetCategoryNewsPaper, asset.Category)
				newspaper, ok := asset.Info.(*catalog.NewsPaper)
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
				assert.Equal(t, catalog.AssetCategoryDVD, asset.Category)
				dvd, ok := asset.Info.(*catalog.DVD)
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
				assert.Equal(t, catalog.AssetCategoryCD, asset.Category)
				cd, ok := asset.Info.(*catalog.CD)
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
				assert.Equal(t, catalog.AssetCategoryVideoGame, asset.Category)
				game, ok := asset.Info.(*catalog.VideoGame)
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

func TestUnmarshallCreateAssetRequest(t *testing.T) {
	testCases := map[string]struct {
		data        func() []byte
		expectedErr error
		assertAsset func(req *catalog.CreateAssetRequest)
	}{
		"it should unmarhsall a book asset": {
			data: func() []byte {
				buf, err := os.ReadFile("testdata/book.json")
				assert.Nil(t, err)
				return buf
			},
			assertAsset: func(req *catalog.CreateAssetRequest) {
				assert.NotNil(t, req)
				assert.Equal(t, catalog.AssetCategoryBook, req.Category)
				book, ok := req.Asset.(*catalog.Book)
				assert.True(t, ok)
				assert.Equal(t, "The Lord Of The Rings The Two Towers", book.Title)
				assert.Equal(t, "J.R.R Tolkien", book.Author)
				assert.Equal(t, "George Allen & Unwin", book.Publisher)
				assert.Equal(t, "978-0261102385", book.ISBN)
				assert.Equal(t, 352, book.PageCount)
				publishedAt, err := time.Parse(time.RFC3339, "2020-05-20T12:30:30Z")
				assert.Nil(t, err)
				assert.Equal(t, publishedAt, book.PublishedAt)
			},
		},
		"it should unmarshall a magazine asset": {
			data: func() []byte {
				buf, err := os.ReadFile("testdata/magazine.json")
				assert.Nil(t, err)
				return buf
			},
			assertAsset: func(req *catalog.CreateAssetRequest) {
				assert.NotNil(t, req)
				assert.Equal(t, catalog.AssetCategoryMagazine, req.Category)
				magazine, ok := req.Asset.(*catalog.Magazine)
				assert.True(t, ok)
				assert.Equal(t, "National Geographic", magazine.Title)
				assert.Equal(t, "March 2022", magazine.Issue)
				assert.Equal(t, "National Geographic Partners", magazine.Publisher)
				assert.Equal(t, 128, magazine.PageCount)
				publishedAt, err := time.Parse(time.RFC3339, "2020-05-20T12:30:30Z")
				assert.Nil(t, err)
				assert.Equal(t, publishedAt, magazine.PublishedAt)
			},
		},
		"it should unmarshall a news paper asset": {
			data: func() []byte {
				buf, err := os.ReadFile("testdata/news_paper.json")
				assert.Nil(t, err)
				return buf
			},
			assertAsset: func(req *catalog.CreateAssetRequest) {
				assert.NotNil(t, req)
				assert.Equal(t, catalog.AssetCategoryNewsPaper, req.Category)
				newsPaper, ok := req.Asset.(*catalog.NewsPaper)
				assert.True(t, ok)
				assert.Equal(t, "The New York Times", newsPaper.Title)
				assert.Equal(t, "Sunday Edition", newsPaper.Edition)
				assert.Equal(t, "The New York Times Company", newsPaper.Publisher)
				assert.Equal(t, 94, newsPaper.PageCount)
				publishedAt, err := time.Parse(time.RFC3339, "2020-05-20T12:30:30Z")
				assert.Nil(t, err)
				assert.Equal(t, publishedAt, newsPaper.PublishedAt)
			},
		},
		"it should unmarshall a DVD asset": {
			data: func() []byte {
				buf, err := os.ReadFile("testdata/dvd.json")
				assert.Nil(t, err)
				return buf
			},
			assertAsset: func(req *catalog.CreateAssetRequest) {
				assert.NotNil(t, req)
				assert.Equal(t, catalog.AssetCategoryDVD, req.Category)
				dvd, ok := req.Asset.(*catalog.DVD)
				assert.True(t, ok)
				assert.Equal(t, "The Lord of the Rings: The Fellowship of the Ring (Extended Edition)", dvd.Title)
				assert.Equal(t, "Peter Jackson", dvd.Director)
				assert.Equal(t, "Barrie M. Osborne", dvd.Producer)
				assert.Equal(t, 228, dvd.DurationMin)
				assert.Equal(t, "Region 1", dvd.RegionCode)
				releasedAt, err := time.Parse(time.RFC3339, "2020-05-20T12:30:30Z")
				assert.Nil(t, err)
				assert.Equal(t, releasedAt, dvd.ReleasedAt)
			},
		},
		"it should unmarshall a CD asset": {
			data: func() []byte {
				buf, err := os.ReadFile("testdata/cd.json")
				assert.Nil(t, err)
				return buf
			},
			assertAsset: func(req *catalog.CreateAssetRequest) {
				assert.NotNil(t, req)
				assert.Equal(t, catalog.AssetCategoryCD, req.Category)
				cd, ok := req.Asset.(*catalog.CD)
				assert.True(t, ok)
				assert.Equal(t, "Thriller", cd.Title)
				assert.Equal(t, "Michael Jackson", cd.Artist)
				assert.Equal(t, "Epic Records", cd.Label)
				assert.Equal(t, 9, cd.TrackCount)
				assert.Equal(t, 42, cd.DurationMin)
				releasedAt, err := time.Parse(time.RFC3339, "2020-05-20T12:30:30Z")
				assert.Nil(t, err)
				assert.Equal(t, releasedAt, cd.ReleasedAt)
			},
		},
		"it should unmarshall a video game asset": {
			data: func() []byte {
				buf, err := os.ReadFile("testdata/video_game.json")
				assert.Nil(t, err)
				return buf
			},
			assertAsset: func(req *catalog.CreateAssetRequest) {
				assert.NotNil(t, req)
				assert.Equal(t, catalog.AssetCategoryVideoGame, req.Category)
				videoGame, ok := req.Asset.(*catalog.VideoGame)
				assert.True(t, ok)
				assert.Equal(t, "The Last of Us", videoGame.Title)
				assert.Equal(t, "Naughty Dog", videoGame.Developer)
				assert.Equal(t, "PlayStation 3", videoGame.Platform)
				assert.Equal(t, "Action-adventure", videoGame.Genre)
				assert.Equal(t, "+18", videoGame.AgeRating)
				releasedAt, err := time.Parse(time.RFC3339, "2020-05-20T12:30:30Z")
				assert.Nil(t, err)
				assert.Equal(t, releasedAt, videoGame.ReleasedAt)
			},
		},
		"it should return an error for a non expected asset": {
			data: func() []byte {
				buf, err := os.ReadFile("testdata/non_expected_asset.json")
				assert.Nil(t, err)
				return buf
			},
			assertAsset: func(req *catalog.CreateAssetRequest) {},
			expectedErr: fmt.Errorf("unknown category: %s", "FAIL"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := &catalog.CreateAssetRequest{}
			err := req.UnmarshalJSON(tc.data())
			assert.Equal(t, tc.expectedErr, err)
			tc.assertAsset(req)
		})
	}
}
