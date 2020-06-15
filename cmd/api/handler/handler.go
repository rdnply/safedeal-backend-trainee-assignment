package handler

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"safedeal-backend-trainee/internal/ehttp"
	"safedeal-backend-trainee/internal/order"
	"safedeal-backend-trainee/internal/product"
	"safedeal-backend-trainee/pkg/log/logger"
	"sync"
	"time"
)

type Handler struct {
	productStorage product.Storage
	orderStorage   order.Storage
	logger         logger.Logger
	visitors       visitors
}

type visitors struct {
	mu sync.Mutex
	m  map[string]*visitor
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func New(p product.Storage, o order.Storage, l logger.Logger) *Handler {
	return &Handler{
		productStorage: p,
		orderStorage:   o,
		logger:         l,
		visitors:       visitors{m: make(map[string]*visitor)},
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/products/{id}/cost-of-delivery", MWError(h.costOfDelivery, h.logger))
		r.Post("/products/{id}/order", MWError(h.createOrder, h.logger))
		r.Get("/orders", MWError(h.getOrders, h.logger))
		r.Get("/orders/{id}", MWError(h.getOrder, h.logger))
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

func (h *Handler) getVisitor(ip string) *rate.Limiter {
	h.visitors.mu.Lock()
	defer h.visitors.mu.Unlock()

	v, exists := h.visitors.m[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Every(1 * time.Minute), 10)
		h.visitors.m[ip] = &visitor{limiter, time.Now()}

		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (h *Handler) CleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		h.visitors.mu.Lock()
		for ip, v := range h.visitors.m {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(h.visitors.m, ip)
			}
		}
		h.visitors.mu.Unlock()
	}
}

func (h *Handler) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, http.StatusText(status), status)

			return
		}

		limiter := h.getVisitor(ip)
		if limiter.Allow() == false {
			status := http.StatusTooManyRequests
			http.Error(w, http.StatusText(status), status)

			return
		}

		next.ServeHTTP(w, r)
	})
}
