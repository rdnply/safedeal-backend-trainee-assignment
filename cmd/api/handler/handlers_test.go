package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"safedeal-backend-trainee/internal/order"
	"safedeal-backend-trainee/internal/product"
	"safedeal-backend-trainee/pkg/log/logger"
	"strings"
	"testing"
)

type mockProductStorage struct {
	p *product.Product
	product.Storage
}

func (m mockProductStorage) FindByID(id int64) (*product.Product, error) {
	return m.p, nil
}

type mockOrderStorage struct {
	o *order.Order
	order.Storage
}

func (m mockOrderStorage) Create(o *order.Order) error {
	o.ID = m.o.ID
	return nil
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