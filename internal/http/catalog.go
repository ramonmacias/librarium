package http

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"librarium/internal/catalog"
	"librarium/internal/query"
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

func (cc *CatalogController) Create(w http.ResponseWriter, r *http.Request) {
	createAssetReq, err := DecodeRequest[catalog.CreateAssetRequest](r)
	if err != nil {
		log.Println("error decoding request while login", err)
		WriteResponse(w, http.StatusBadRequest, errors.New("error decoding asset request"))
		return
	}

	asset, err := catalog.BuildAsset(createAssetReq.Asset)
	if err != nil {
		log.Println("error building asset catalog", err)
		WriteResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := cc.catalogRepository.CreateAsset(asset); err != nil {
		log.Println("error creating asset catalog", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error creating asset catalog"))
		return
	}

	WriteResponse(w, http.StatusOK, struct {
		ID string `json:"id"`
	}{ID: asset.ID.String()})
}

func (cc *CatalogController) Delete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 {
		WriteResponse(w, http.StatusBadRequest, errors.New("error getting asset ID from the url"))
		return
	}
	assetID, err := uuid.Parse(parts[2])
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, errors.New("invalid asset ID format, expected UUID"))
		return
	}

	asset, err := cc.catalogRepository.GetAsset(assetID)
	if err != nil {
		log.Println("error getting asset catalog while deleting", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting asset catalog"))
		return
	}
	if asset == nil {
		WriteResponse(w, http.StatusNotFound, errors.New("asset not found"))
		return
	}

	if err := cc.catalogRepository.DeleteAsset(asset.ID); err != nil {
		log.Println("error deleting asset catalog", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error deleting asset catalog"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cc *CatalogController) Find(w http.ResponseWriter, r *http.Request) {
	pagination, err := query.PaginationFromHTTPRequest(r)
	if err != nil {
		log.Println("error getting pagination from the request", err)
		WriteResponse(w, http.StatusBadRequest, err)
		return
	}

	assets, err := cc.catalogRepository.FindAssets(
		query.FiltersFromHTTPRequest(r),
		query.SortingFromHTTPRequest(r),
		pagination,
	)
	if err != nil {
		log.Println("error finding catalog assets", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error finding catalog assets"))
		return
	}

	WriteResponse(w, http.StatusOK, assets)
}
