package handler

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"math/rand"
	"net/http"
	"safedeal-backend-trainee/cmd/api/render"
	"safedeal-backend-trainee/internal/order"
	"safedeal-backend-trainee/internal/product"
	"safedeal-backend-trainee/pkg/log/logger"
	"strconv"
	"strings"
	"time"
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
		r.Post("/products/{id}/cost-of-delivery", h.costOfDelivery)
	})

	return r
}

const BottomLineValidID = 0

func (h *Handler) costOfDelivery(w http.ResponseWriter, r *http.Request) {
	type destination struct {
		Address string `json:"destination"`
	}

	var d destination

	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		h.logger.Errorf("can't unmarshal input json: %v", err)
		render.HTTPError("", http.StatusBadRequest, w)

		return
	}

	id, err := IDFromParams(r)
	if err != nil {
		h.logger.Errorf("can't get ID from URL params: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)

		return
	}

	if id <= BottomLineValidID {
		h.logger.Errorf("don't valid id: %v", id)
		msg := fmt.Sprintf("incorrect id: %v", id)
		render.HTTPError(msg, http.StatusBadRequest, w)

		return
	}

	product, err := h.productStorage.FindByID(id)
	if err != nil {
		h.logger.Errorf("can't find product with id= %v: %v", id, err)
		render.HTTPError("", http.StatusInternalServerError, w)

		return
	}

	price := calcPrice(product, d.Address)

	err = respondJSON(w, map[string]interface{}{
		"from":        product.Place,
		"destination": d.Address,
		"price":       price,
	})
	if err != nil {
		h.logger.Errorf("can't respond json with delivery info: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)

		return
	}
}

// calcPrice возвращает цену доставки
// (цена - случайное число в диапазоне [min * fact, max * fact])
func calcPrice(product *product.Product, dest string) int {
	const min = 3
	const max = 20
	const fact = 100
	rand.Seed(time.Now().UnixNano())

	return (rand.Intn(max-min+1) + min) * fact
}

func IDFromParams(r *http.Request) (int64, error) {
	const IDIndex = 4 // space in first place

	str := r.URL.String()
	params := strings.Split(str, "/")

	id, err := strconv.ParseInt(params[IDIndex], 10, 64)
	if err != nil {
		return -1, errors.Wrap(err, "can't parse string to int for getting id from params")
	}

	return id, nil
}

func respondJSON(w http.ResponseWriter, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrapf(err, "can't marshal respond to json")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	c, err := w.Write(response)
	if err != nil {
		msg := fmt.Sprintf("can't write json data in respond, code: %v", c)
		return errors.Wrapf(err, msg)
	}

	return nil
}
