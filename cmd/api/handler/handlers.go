package handler

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"math/rand"
	"net/http"
	"safedeal-backend-trainee/internal/ehttp"
	"safedeal-backend-trainee/internal/ftime"
	"safedeal-backend-trainee/internal/order"
	"safedeal-backend-trainee/internal/product"
	"strconv"
	"strings"
	"time"
)

const BottomLineValidID = 0

func (h *Handler) costOfDelivery(w http.ResponseWriter, r *http.Request) error {
	type destination struct {
		Address string `json:"destination"`
	}

	var d destination

	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		return ehttp.JSONUnmarshalErr(err)
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

	if product.ID == BottomLineValidID {
		msg := fmt.Sprintf("can't find product with id= %v", id)
		detail := fmt.Sprintf("%v: %v", msg, err)
		return ehttp.NotFoundErr(msg, detail)
	}

	price := calcPrice()

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
func calcPrice() int {
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
	type orderInfo struct {
		Address string    `json:"destination"`
		Time    time.Time `json:"time"`
	}

	var info orderInfo

	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		return ehttp.JSONUnmarshalErr(err)
	}

	id, err := getIDFromRequest(r)
	if err != nil {
		return err
	}

	product, err := h.productStorage.FindByID(id)
	if err != nil {
		detail := fmt.Sprintf("can't find product with productID= %v: %v", id, err)
		return ehttp.InternalServerErr(detail)
	}

	if product.ID == BottomLineValidID {
		msg := fmt.Sprintf("can't find product with id= %v", id)
		detail := fmt.Sprintf("%v: %v", msg, err)
		return ehttp.NotFoundErr(msg, detail)
	}

	order := NewOrder(product, info.Address, info.Time)

	err = h.orderStorage.Create(order)
	if err != nil {
		detail := fmt.Sprintf("can't can't create order with productID= %v: %v", order.ProductID, err)
		return ehttp.InternalServerErr(detail)
	}

	err = respondJSON(w, order)
	if err != nil {
		detail := fmt.Sprintf("can't respond json with order's info: %v", err)
		return ehttp.InternalServerErr(detail)
	}

	return nil
}

func NewOrder(p *product.Product, dest string, t time.Time) *order.Order {
	return &order.Order{
		ProductID:   p.ID,
		Name:        p.Name,
		From:        p.Place,
		Destination: dest,
		Time:        ftime.New(t),
	}
}

func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) error {
	orders, err := h.orderStorage.GetAll()
	if err != nil {
		detail := fmt.Sprintf("can't get all orders: %v", err)
		return ehttp.InternalServerErr(detail)
	}

	orders = removeExtraInfo(orders)

	err = respondJSON(w, orders)
	if err != nil {
		detail := fmt.Sprintf("can't respond json with all orders info: %v", err)
		return ehttp.InternalServerErr(detail)
	}

	return nil
}

func removeExtraInfo(old []*order.Order) []*order.Order {
	res := make([]*order.Order, len(old))
	copy(res, old)
	for _, o := range res {
		o.From = ""
		o.Destination = ""
		o.Time = nil
	}

	return res
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) error {
	orderID, err := getIDFromRequest(r)
	if err != nil {
		return err
	}

	order, err := h.orderStorage.FindByID(orderID)
	if err != nil {
		detail := fmt.Sprintf("can't find order with ID = %v: %v", orderID, err)
		return ehttp.InternalServerErr(detail)
	}

	if order.ID == BottomLineValidID {
		msg := fmt.Sprintf("can't find order with id= %v", orderID)
		detail := fmt.Sprintf("%v: %v", msg, err)
		return ehttp.NotFoundErr(msg, detail)
	}

	pr, err := h.productStorage.FindByID(order.ProductID)
	if err != nil {
		detail := fmt.Sprintf("can't find product: %v", err)
		return ehttp.InternalServerErr(detail)
	}

	if pr.ID == BottomLineValidID {
		msg := fmt.Sprintf("can't find product with id= %v", order.ProductID)
		detail := fmt.Sprintf("%v: %v", msg, err)
		return ehttp.NotFoundErr(msg, detail)
	}

	err = respondJSON(w, struct {
		ID          int64            `json:"id"`
		Product     product.Product  `json:"product"`
		From        string           `json:"from"`
		Destination string           `json:"destination"`
		Time        ftime.FormatTime `json:"time"`
	}{
		ID:          order.ID,
		Product:     *pr,
		From:        order.From,
		Destination: order.Destination,
		Time:        *order.Time,
	})
	if err != nil {
		detail := fmt.Sprintf("can't respond json with order's detailed info: %v", err)
		return ehttp.InternalServerErr(detail)
	}

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
