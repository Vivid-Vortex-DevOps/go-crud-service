package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/model"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductHandler struct {
	svc    service.ProductService
	logger *slog.Logger
}

func NewProductHandler(svc service.ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, logger: logger}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{id}", h.getByID)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	return r
}

func (h *ProductHandler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/products/"+product.ID.String())
	writeJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) list(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))

	products, total, err := h.svc.List(r.Context(), page, size)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	writePaginated(w, products, total, page, size)
}

func (h *ProductHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(w, r)
	if err != nil {
		return
	}

	product, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(w, r)
	if err != nil {
		return
	}

	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(w, r)
	if err != nil {
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var notFound model.ErrNotFound
	var validation model.ErrValidation

	switch {
	case errors.As(err, &notFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.As(err, &validation):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		h.logger.ErrorContext(r.Context(), "unhandled error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func parseUUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "id")
	id, err := uuid.Parse(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product ID format")
	}
	return id, err
}
