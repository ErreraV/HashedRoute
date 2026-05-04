package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type api struct {
	cfg      GatewayConfig
	gw       *client.Gateway
	contract *client.Contract
}

type createShipmentReq struct {
	ID          string `json:"id"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Customer    string `json:"customer"`
	Carrier     string `json:"carrier"`
}

type patchStatusReq struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

type errResp struct {
	Error string `json:"error"`
}

func jsonWrite(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *api) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		jsonWrite(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/shipments", a.listShipments)
		r.Post("/shipments", a.createShipment)
		r.Get("/shipments/{id}", a.getShipment)
		r.Patch("/shipments/{id}/status", a.patchStatus)
	})
	return r
}

func cors(next http.Handler) http.Handler {
	allowed := "http://localhost:5173"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *api) createShipment(w http.ResponseWriter, r *http.Request) {
	var body createShipmentReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonWrite(w, http.StatusBadRequest, errResp{Error: "invalid JSON body"})
		return
	}
	if body.ID == "" {
		jsonWrite(w, http.StatusBadRequest, errResp{Error: "id is required"})
		return
	}
	_, err := a.contract.SubmitTransaction(
		"CreateShipment",
		body.ID,
		body.Origin,
		body.Destination,
		body.Customer,
		body.Carrier,
	)
	if err != nil {
		jsonWrite(w, http.StatusBadRequest, errResp{Error: err.Error()})
		return
	}
	jsonWrite(w, http.StatusCreated, map[string]string{"id": body.ID})
}

func (a *api) getShipment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	raw, err := a.contract.EvaluateTransaction("GetShipment", id)
	if err != nil {
		jsonWrite(w, http.StatusNotFound, errResp{Error: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func (a *api) listShipments(w http.ResponseWriter, r *http.Request) {
	raw, err := a.contract.EvaluateTransaction("ListShipments")
	if err != nil {
		jsonWrite(w, http.StatusBadGateway, errResp{Error: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func (a *api) patchStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body patchStatusReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonWrite(w, http.StatusBadRequest, errResp{Error: "invalid JSON body"})
		return
	}
	if body.Status == "" {
		jsonWrite(w, http.StatusBadRequest, errResp{Error: "status is required"})
		return
	}
	raw, err := a.contract.SubmitTransaction("UpdateStatus", id, body.Status, body.Notes)
	if err != nil {
		jsonWrite(w, http.StatusBadRequest, errResp{Error: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if len(raw) > 0 {
		_, _ = w.Write(raw)
	} else {
		_, _ = w.Write([]byte("{}"))
	}
}
