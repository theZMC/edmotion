package httpapi

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/thezmc/edmotion/internal/admin"
	"github.com/thezmc/edmotion/internal/challenge"
)

type Params struct {
	Catalog               *challenge.Catalog
	Admin                 *admin.State
	RequestLimitPerMinute int
}

func NewRouter(params Params) (chi.Router, error) {
	if params.Admin == nil {
		return nil, fmt.Errorf("admin state is required")
	}

	if params.Catalog == nil {
		return nil, fmt.Errorf("challenge catalog is required")
	}

	r := chi.NewRouter()
	r.Use(defaultMiddlewares(params.Admin, params.RequestLimitPerMinute)...)

	params.Catalog.RegisterRoutes(r)

	admin.RegisterRoutes(r, params.Admin)

	return r, nil
}
