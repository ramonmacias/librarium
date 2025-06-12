package http

import (
	"encoding/json"
	"errors"
	"librarium/internal/catalog"
	"log"
	"net/http"
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

}

func (cc *CatalogController) FindCatalogAssets(w http.ResponseWriter, r *http.Request) {

}
