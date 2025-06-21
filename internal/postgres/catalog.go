package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"librarium/internal/catalog"
	"librarium/internal/query"

	"github.com/google/uuid"
)

type catalogRepository struct {
	db *sql.DB
}

// NewCatalogRepository builds a new catalog.Repository implemented in postgres.
// It returns an error if the provided db connection is nil.
func NewCatalogRepository(db *sql.DB) (catalog.Repository, error) {
	if db == nil {
		return nil, errors.New("error while building catalog repository, db is nil")
	}
	return &catalogRepository{db}, nil
}

// CreateAsset inserts the provided asset into the database.
// It returns an error in case of failure.
func (cr *catalogRepository) CreateAsset(asset *catalog.Asset) error {
	infoJsonBlob, err := json.Marshal(asset.Info)
	if err != nil {
		return fmt.Errorf("error marshalling asset catalog info %w", err)
	}

	_, err = cr.db.Exec(
		"INSERT INTO assets (id, category, created_at, updated_at, info) VALUES ($1, $2, $3, $4, $5)",
		asset.ID.String(),
		asset.Category,
		asset.CreatedAt,
		asset.UpdatedAt,
		infoJsonBlob,
	)
	if err != nil {
		return fmt.Errorf("error inserting asset catalog in postgres %w", err)
	}

	return nil
}

// DeleteAsset removes the asset linked to the given ID from the catalog.
// It returns an error in case of failure.
func (cr *catalogRepository) DeleteAsset(id uuid.UUID) error {
	_, err := cr.db.Exec(
		"DELETE FROM assets WHERE id = $1",
		id.String(),
	)
	if err != nil {
		return fmt.Errorf("error deleting asset catalog with id %s err %w", id.String(), err)
	}

	return nil
}

// GetAsset retrieves the asset linked to the given ID.
// It returns nil, nil in case the asset cannot be found.
// It returns an error in case of failure.
func (cr *catalogRepository) GetAsset(id uuid.UUID) (*catalog.Asset, error) {
	asset := &catalog.Asset{}
	buf := []byte{}

	err := cr.db.QueryRow("SELECT id, category, created_at, updated_at, info FROM assets WHERE id = $1", id.String()).Scan(&asset.ID, &asset.Category, &asset.CreatedAt, &asset.UpdatedAt, &buf)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting catalog asset, with id %s err %w", id.String(), err)
	}
	asset.Info, err = decodeAssetInfo(asset.Category, buf)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

// FindAssets looks for the assets already inserted in the database.
// Returns an empty slice and no error in case of no asset found.
// It returns an error if something fails.
func (cr *catalogRepository) FindAssets(filters query.Filters, sorting *query.Sorting, pagination *query.Pagination) ([]*catalog.Asset, error) {
	rows, err := cr.db.Query("SELECT id, category, created_at, updated_at, info FROM assets WHERE $1 $2 $3", query.SQLFilterBy(filters, map[string]string{
		"id":       "assets.id",
		"category": "assets.category",
	}), query.SQLPaginateBy(pagination), query.SQLSortBy([]query.Sorting{*sorting}, map[string]string{
		"id":       "assets.id",
		"category": "assets.category",
	}))
	if err != nil {
		return nil, fmt.Errorf("error querying for finding assets %w", err)
	}
	defer rows.Close()

	assets := []*catalog.Asset{}
	for rows.Next() {
		asset := &catalog.Asset{}
		buf := []byte{}

		if err := rows.Scan(&asset.ID, &asset.Category, &asset.CreatedAt, &asset.UpdatedAt, &buf); err != nil {
			return nil, fmt.Errorf("error scanning while finding assets %w", err)
		}

		asset.Info, err = decodeAssetInfo(asset.Category, buf)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while going through the asset rows %w", err)
	}
	return assets, nil
}

var (
	unmarshalTargets = map[catalog.AssetCategory]func() any{
		catalog.AssetCategoryBook:      func() any { return &catalog.Book{} },
		catalog.AssetCategoryMagazine:  func() any { return &catalog.Magazine{} },
		catalog.AssetCategoryCD:        func() any { return &catalog.CD{} },
		catalog.AssetCategoryDVD:       func() any { return &catalog.DVD{} },
		catalog.AssetCategoryNewsPaper: func() any { return &catalog.NewsPaper{} },
		catalog.AssetCategoryVideoGame: func() any { return &catalog.VideoGame{} },
	}
)

func decodeAssetInfo(category catalog.AssetCategory, buf []byte) (any, error) {
	constructor, ok := unmarshalTargets[category]
	if !ok {
		return nil, fmt.Errorf("unsupported asset category: %s", category)
	}

	target := constructor()
	if err := json.Unmarshal(buf, target); err != nil {
		return nil, err
	}

	return target, nil
}
