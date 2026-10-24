package pricing

import (
	"net/http"

	"github.com/oche/diagram/pkg/response"
)

type Handler struct {
	detector *Detector
}

func NewHandler(detector *Detector) *Handler {
	return &Handler{detector: detector}
}

// Get returns the localized pricing catalog. Public endpoint.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	country := ""
	if h.detector != nil {
		country = h.detector.Detect(r)
	}

	// Collapse "NG" to the Nigeria catalog; everything else uses the global catalog.
	catalog := Build("")
	if country == RegionNigeria {
		catalog = Build(RegionNigeria)
	}
	if country != "" {
		catalog.Country = country
	}

	response.JSON(w, http.StatusOK, catalog)
}
