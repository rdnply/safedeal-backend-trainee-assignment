package handler

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"math/rand"
	"net/http"
	"safedeal-backend-trainee/internal/ehttp"
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
		r.Post("/products/{id}/cost-of-delivery", MWError(h.costOfDelivery, h.logger))
		r.Post("/products/{id}/order", MWError(h.createOrder, h.logger))
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

			l.Errorf(e.Msg)
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

const BottomLineValidID = 0

func (h *Handler) costOfDelivery(w http.ResponseWriter, r *http.Request) error {
	type destination struct {
		Address string `json:"destination"`
	}

	var d destination

	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		return ehttp.JSONUmmarshalErr(err)
	}

	id, err := getIDFromRequest(r)
	if err != nil {
		return err
	}

	product, err := h.productStorage.FindByID(id)
	if err != nil {
		detail := fmt.Sprintf("can't find product with id= %v: %v", id, err)
		return ehttp.InternalServerErr(detail)
	}

	price := calcPrice(product, d.Address)

	err = respondJSON(w, map[string]interface{}{
		"from":        product.Place,
		"destination": d.Address,
		"price":       price,
	})
	if err != nil {
		detail := fmt.Sprintf("can't respond json with delivery info: %v", err)
		return ehttp.InternalServerErr(detail)
	}

	return nil
}

func getIDFromRequest(r *http.Request) (int64, error) {
	id, err := IDFromParams(r)
	if err != nil {
		detail := fmt.Sprintf("can't get ID from URL params: %v", err)
		return -1, ehttp.InternalServerErr(detail)
	}

	if id <= BottomLineValidID {
		return -1, ehttp.IncorrectID(id)
	}

	return id, nil
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

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) error {
	return nil
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
