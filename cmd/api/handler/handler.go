package handler

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"net/http"
	"safedeal-backend-trainee/internal/ehttp"
	"safedeal-backend-trainee/internal/order"
	"safedeal-backend-trainee/internal/product"
	"safedeal-backend-trainee/pkg/log/logger"
)

type Handler struct {
	productStorage product.Storage
	orderStorage   order.Storage
	logger         logger.Logger
}

func New(p product.Storage, o order.Storage, l logger.Logger) *Handler {
	return &Handler{
		productStorage: p,
		orderStorage:   o,
		logger:         l,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/products/{id}/cost-of-delivery", MWError(h.costOfDelivery, h.logger))
		r.Post("/products/{id}/order", MWError(h.createOrder, h.logger))
		r.Get("/orders", MWError(h.getOrders, h.logger))
	})

	return r
}

type handlerFunc func(http.ResponseWriter, *http.Request) error

func MWError(h handlerFunc, l logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			e, ok := err.(ehttp.HTTPError)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if e.Detail != "" {
				l.Errorf(e.Detail)
			}

			w.WriteHeader(e.StatusCode)

			if e.Msg != "" {
				w.Header().Add("Content-Type", "application/json")

				out, err := json.Marshal(e)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// no need to handle error here
				_, _ = w.Write(out)
			}
		}
	}
}
