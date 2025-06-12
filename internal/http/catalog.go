package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"librarium/internal/catalog"
)

// CatalogController holds all the dependencies needed to
// handle all the http requests related with the catalog domain.
type CatalogController struct {
	catalogRepository catalog.Repository
}

// NewCatalogController builds a new catalog controller to handle http requests
// using the given data, all the params received are mandatory.
// It returns an error if some mandatory data is missing.
func NewCatalogController(catalogRepository catalog.Repository) (*CatalogController, error) {
	if catalogRepository == nil {
		return nil, errors.New("catalog repository is mandatory")
	}

	return &CatalogController{catalogRepository}, nil
}

func (cc *CatalogController) CreateCatalogAsset(w http.ResponseWriter, r *http.Request) {
	createAssetReq := &catalog.CreateAssetRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(createAssetReq); err != nil {
		log.Println("error decoding request while login", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding asset request")
		return
	}

	asset, err := catalog.BuildAsset(createAssetReq.Asset)
	if err != nil {
		log.Println("error building asset catalog", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error building asset catalog")
		return
	}

	if err := cc.catalogRepository.CreateAsset(asset); err != nil {
		log.Println("error creating asset catalog", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error creating asset catalog")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct{ ID string }{ID: asset.ID.String()})
}

func (cc *CatalogController) DeleteCatalogAsset(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error getting asset ID from the url")
		return
	}
	assetID, err := uuid.Parse(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid asset ID format, expected UUID")
		return
	}

	asset, err := cc.catalogRepository.GetAsset(assetID)
	if err != nil {
		log.Println("error getting asset catalog while deleting", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error getting asset catalog")
		return
	}

	if err := cc.catalogRepository.DeleteAsset(asset.ID); err != nil {
		log.Println("error deleting asset catalog", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error deleting asset catalog")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cc *CatalogController) FindCatalogAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := cc.catalogRepository.FindAssets()
	if err != nil {
		log.Println("error finding catalog assets", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error finding catalog assets")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(assets)
}
