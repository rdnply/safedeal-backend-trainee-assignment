package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"safedeal-backend-trainee/internal/ftime"
	"safedeal-backend-trainee/internal/order"
	"safedeal-backend-trainee/internal/product"
	"safedeal-backend-trainee/pkg/log/logger"
	"strings"
	"testing"
	"time"
)

type mockProductStorage struct {
	p *product.Product
	product.Storage
}

func (m mockProductStorage) FindByID(id int64) (*product.Product, error) {
	return m.p, nil
}

type mockOrderStorage struct {
	o  *order.Order
	oo []*order.Order
	order.Storage
}

func (m mockOrderStorage) Create(o *order.Order) error {
	o.ID = m.o.ID
	return nil
}

func (m mockOrderStorage) GetAll() ([]*order.Order, error) {
	return m.oo, nil
}

func (m mockOrderStorage) FindByID(id int64) (*order.Order, error) {
	return m.o, nil
}

type mockLogger struct {
	logger.Logger
}

func (m mockLogger) Debugf(format string, args ...interface{}) {}
func (m mockLogger) Infof(format string, args ...interface{})  {}
func (m mockLogger) Warnf(format string, args ...interface{})  {}
func (m mockLogger) Errorf(format string, args ...interface{}) {}
func (m mockLogger) Fatalf(format string, args ...interface{}) {}
func (m mockLogger) Panicf(format string, args ...interface{}) {}

func respContains(in string, want string) bool {
	if in == "" {
		return want == ""
	}

	return strings.Contains(in, want)
}

func TestCostOfDeliveryCorrect(t *testing.T) {
	json := []byte(`{"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50"}`)
	req, err := http.NewRequest("POST", "/api/v1/products/1/cost-of-delivery", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	place := "Тверской бульвар, 25"

	p := &product.Product{
		ID:    1,
		Place: place,
	}

	mockProductStorage.p = p

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.costOfDelivery, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("costOfDelivery handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `{"destination":"Большая Садовая, 302-бис, пятый этаж, кв. № 50","from":"Тверской бульвар, 25"`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("costOfDelivery handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestCostOfDeliveryNotFound(t *testing.T) {
	json := []byte(`{"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50"}`)
	req, err := http.NewRequest("POST", "/api/v1/products/1/cost-of-delivery", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	p := &product.Product{
		ID: 0, // zero value => can't find product in storage
	}

	mockProductStorage.p = p

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.costOfDelivery, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("costOfDelivery handler returned wrong status code: got %v, want %v",
			status, http.StatusNotFound)
	}

	expected := `{"error":"can't find product with id= 1"}`
	if rr.Body.String() != expected {
		t.Errorf("costOfDelivery handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestCostOfDeliveryIncorrectID(t *testing.T) {
	json := []byte(`{"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50"}`)
	req, err := http.NewRequest("POST", "/api/v1/products/-1/cost-of-delivery", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	place := "Тверской бульвар, 25"

	p := &product.Product{
		ID:    1,
		Place: place,
	}

	mockProductStorage.p = p

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.costOfDelivery, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("costOfDelivery handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}

	expected := fmt.Sprintf("{\"error\":\"incorrect id: %v\"}", -1)
	if rr.Body.String() != expected {
		t.Errorf("costOfDelivery handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestCostOfDeliveryIncorrectJSON(t *testing.T) {
	// body contains incorrect json(missing a open bracket)
	json := []byte(`"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50"}`)
	req, err := http.NewRequest("POST", "/api/v1/products/1/cost-of-delivery", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	place := "Тверской бульвар, 25"

	p := &product.Product{
		ID:    1,
		Place: place,
	}

	mockProductStorage.p = p

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.costOfDelivery, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("costOfDelivery handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}

	expected := ""
	if rr.Body.String() != expected {
		t.Errorf("costOfDelivery handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestCreateOrderCorrect(t *testing.T) {
	json := []byte(`{"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50", "time" : "2020-06-15T13:30:00Z"}`)
	req, err := http.NewRequest("POST", "/api/v1/products/1/order", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	place := "Тверской бульвар, 25"

	p := &product.Product{
		ID:    1,
		Name:  "Название",
		Place: place,
	}

	o := &order.Order{
		ID: 5,
	}

	mockProductStorage.p = p
	mockOrderStorage.o = o

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.createOrder, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("createOrder handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `{"id":5,"product_id":1,"name":"Название","from":"Тверской бульвар, 25",` +
		`"destination":"Большая Садовая, 302-бис, пятый этаж, кв. № 50","time":"2020-06-15T13:30:00Z"}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("createOrder handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestCreateOrderIncorrectID(t *testing.T) {
	json := []byte(`{"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50", "time" : "2020-06-15T13:30:00Z"}`)
	req, err := http.NewRequest("POST", "/api/v1/products/-1/order", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	place := "Тверской бульвар, 25"

	p := &product.Product{
		ID:    1,
		Place: place,
	}

	mockProductStorage.p = p

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.createOrder, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("createOrder handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}

	expected := fmt.Sprintf("{\"error\":\"incorrect id: %v\"}", -1)
	if rr.Body.String() != expected {
		t.Errorf("createOrder handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestCreateOrderIncorrectJSON(t *testing.T) {
	// body contains incorrect json(missing a open bracket)
	json := []byte(`"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50", "time" : "2020-06-15T13:30:00Z"}`)
	req, err := http.NewRequest("POST", "/api/v1/products/1/order", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	place := "Тверской бульвар, 25"

	p := &product.Product{
		ID:    1,
		Place: place,
	}

	mockProductStorage.p = p

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.createOrder, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("createOrder handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}

	expected := ""
	if rr.Body.String() != expected {
		t.Errorf("createOrder handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestCreateOrderNotFound(t *testing.T) {
	json := []byte(`{"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50", "time" : "2020-06-15T13:30:00Z"}`)
	req, err := http.NewRequest("POST", "/api/v1/products/1/order", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	p := &product.Product{
		ID: 0, // zero value => can't find product in storage
	}

	o := &order.Order{
		ID: 5,
	}

	mockProductStorage.p = p
	mockOrderStorage.o = o

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.createOrder, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("createOrder handler returned wrong status code: got %v, want %v",
			status, http.StatusNotFound)
	}

	expected := `{"error":"can't find product with id= 1"}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("createOrder handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetOrdersCorrect(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/orders", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	orders := []*order.Order{
		{ID: 1, ProductID: 1, Name: "Первое название"},
		{ID: 2, ProductID: 1, Name: "Второе название"},
	}

	mockOrderStorage.oo = orders

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.getOrders, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("getOrders handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `[{"id":1,"product_id":1,"name":"Первое название"},{"id":2,"product_id":1,"name":"Второе название"}]`
	if rr.Body.String() != expected {
		t.Errorf("getOrders handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetOrderCorrect(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/orders/1", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	place := "Большой Патриарший пер., 7, строение 1"

	p := &product.Product{
		ID:     1,
		Name:   "Сноуборд",
		Width:  40.5,
		Length: 143,
		Height: 20,
		Weight: 3.3,
		Place:  place,
	}

	str := "2020-06-17T15:30:00Z"
	tt, _ := time.Parse(ftime.Layout, str)
	time := ftime.New(tt)

	o := &order.Order{
		ID:          2,
		ProductID:   1,
		Name:        "Сноуборд",
		From:        place,
		Destination: "Большая Садовая, 302-бис, пятый этаж, кв. № 50",
		Time:        time,
	}

	mockProductStorage.p = p
	mockOrderStorage.o = o

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.getOrder, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("getOrder handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `{"id":2,"product":{"id":1,"name":"Сноуборд","width":40.5,"length":143,"height":20,"weight":3.3,` +
		`"place":"Большой Патриарший пер., 7, строение 1"},"from":"Большой Патриарший пер., 7, строение 1",` +
		`"destination":"Большая Садовая, 302-бис, пятый этаж, кв. № 50","time":"2020-06-17T15:30:00Z"}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getOrder handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetOrderNotFoundOrder(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/orders/1", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	place := "Большой Патриарший пер., 7, строение 1"

	p := &product.Product{
		ID:     1,
		Name:   "Сноуборд",
		Width:  40.5,
		Length: 143,
		Height: 20,
		Weight: 3.3,
		Place:  place,
	}

	o := &order.Order{
		ID: 0, // zero value => can't find order in storage
	}

	mockProductStorage.p = p
	mockOrderStorage.o = o

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.getOrder, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("getOrder handler returned wrong status code: got %v, want %v",
			status, http.StatusNotFound)
	}

	expected := `{"error":"can't find order with id= 1"}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getOrder handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetOrderNotFoundProduct(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/orders/1", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	mockProductStorage := new(mockProductStorage)
	mockOrderStorage := new(mockOrderStorage)

	place := "Большой Патриарший пер., 7, строение 1"

	p := &product.Product{
		ID: 0, // zero value => can't find product in storage
	}

	str := "2020-06-17T15:30:00Z"
	tt, _ := time.Parse(ftime.Layout, str)
	time := ftime.New(tt)

	o := &order.Order{
		ID:          2,
		ProductID:   1,
		Name:        "Сноуборд",
		From:        place,
		Destination: "Большая Садовая, 302-бис, пятый этаж, кв. № 50",
		Time:        time,
	}

	mockProductStorage.p = p
	mockOrderStorage.o = o

	h := New(mockProductStorage, mockOrderStorage, l)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MWError(h.getOrder, l))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("getOrder handler returned wrong status code: got %v, want %v",
			status, http.StatusNotFound)
	}

	expected := `{"error":"can't find product with id= 1"}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getOrder handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}
