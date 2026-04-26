package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	// Catalog Management RabbitMQ topics
	catalogExchange = "catalog_events"

	// Catalog topics
	cmdCatalogCreate = "cmd.catalog.catalog.create"
	cmdCatalogUpdate = "cmd.catalog.catalog.update"
	cmdCatalogDelete = "cmd.catalog.catalog.delete"
	queryCatalogGet  = "query.catalog.catalog.get"
	queryCatalogList = "query.catalog.catalog.list"

	// Category topics
	cmdCategoryCreate = "cmd.catalog.category.create"
	cmdCategoryUpdate = "cmd.catalog.category.update"
	cmdCategoryDelete = "cmd.catalog.category.delete"
	queryCategoryGet  = "query.catalog.category.get"
	queryCategoryList = "query.catalog.category.list"

	// Specification topics
	cmdSpecCreate = "cmd.catalog.specification.create"
	cmdSpecUpdate = "cmd.catalog.specification.update"
	cmdSpecDelete = "cmd.catalog.specification.delete"
	querySpecGet  = "query.catalog.specification.get"
	querySpecList = "query.catalog.specification.list"

	// Offering topics
	cmdOfferingCreate = "cmd.catalog.offering.create"
	cmdOfferingUpdate = "cmd.catalog.offering.update"
	cmdOfferingDelete = "cmd.catalog.offering.delete"
	queryOfferingGet  = "query.catalog.offering.get"
	queryOfferingList = "query.catalog.offering.list"

	catalogRPCTimeout = 10 * time.Second
)

// CatalogHandler handles Catalog Management API endpoints
type CatalogHandler struct {
	rpcClient RPCClient
}

// NewCatalogHandler creates a new CatalogHandler
func NewCatalogHandler(client RPCClient) *CatalogHandler {
	return &CatalogHandler{rpcClient: client}
}

// RegisterRoutes registers all catalog routes
func (h *CatalogHandler) RegisterRoutes(mux *http.ServeMux) {
	// Catalog routes
	mux.HandleFunc("GET /api/catalogs", h.ListCatalogs)
	mux.HandleFunc("POST /api/catalogs", h.CreateCatalog)
	mux.HandleFunc("GET /api/catalogs/{id}", h.GetCatalog)
	mux.HandleFunc("PUT /api/catalogs/{id}", h.UpdateCatalog)
	mux.HandleFunc("DELETE /api/catalogs/{id}", h.DeleteCatalog)

	// Category routes
	mux.HandleFunc("GET /api/categories", h.ListCategories)
	mux.HandleFunc("POST /api/categories", h.CreateCategory)
	mux.HandleFunc("GET /api/categories/{id}", h.GetCategory)
	mux.HandleFunc("PUT /api/categories/{id}", h.UpdateCategory)
	mux.HandleFunc("DELETE /api/categories/{id}", h.DeleteCategory)

	// Specification routes
	mux.HandleFunc("GET /api/specifications", h.ListSpecifications)
	mux.HandleFunc("POST /api/specifications", h.CreateSpecification)
	mux.HandleFunc("GET /api/specifications/{id}", h.GetSpecification)
	mux.HandleFunc("PUT /api/specifications/{id}", h.UpdateSpecification)
	mux.HandleFunc("DELETE /api/specifications/{id}", h.DeleteSpecification)

	// Offering routes
	mux.HandleFunc("GET /api/offerings", h.ListOfferings)
	mux.HandleFunc("POST /api/offerings", h.CreateOffering)
	mux.HandleFunc("GET /api/offerings/{id}", h.GetOffering)
	mux.HandleFunc("PUT /api/offerings/{id}", h.UpdateOffering)
	mux.HandleFunc("DELETE /api/offerings/{id}", h.DeleteOffering)
}

// --- Catalog Handlers ---

func (h *CatalogHandler) ListCatalogs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), catalogRPCTimeout)
	defer cancel()

	payload := map[string]any{} // Add filters if needed
	responseBytes, err := h.rpcClient.CallRPC(ctx, catalogExchange, queryCatalogList, payload, getHeaders(r))
	if err != nil {
		slog.Error("error listing catalogs", "error", err)
		http.Error(w, "Failed to list catalogs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

func (h *CatalogHandler) CreateCatalog(w http.ResponseWriter, r *http.Request) {
	h.handleCommand(w, r, cmdCatalogCreate, "creating catalog")
}

func (h *CatalogHandler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	h.handleQueryByID(w, r, queryCatalogGet, "getting catalog")
}

func (h *CatalogHandler) UpdateCatalog(w http.ResponseWriter, r *http.Request) {
	h.handleCommandWithID(w, r, cmdCatalogUpdate, "updating catalog")
}

func (h *CatalogHandler) DeleteCatalog(w http.ResponseWriter, r *http.Request) {
	h.handleDelete(w, r, cmdCatalogDelete, "deleting catalog")
}

// --- Category Handlers ---

func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), catalogRPCTimeout)
	defer cancel()

	payload := map[string]any{}
	responseBytes, err := h.rpcClient.CallRPC(ctx, catalogExchange, queryCategoryList, payload, getHeaders(r))
	if err != nil {
		slog.Error("error listing categories", "error", err)
		http.Error(w, "Failed to list categories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

func (h *CatalogHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	h.handleCommand(w, r, cmdCategoryCreate, "creating category")
}

func (h *CatalogHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	h.handleQueryByID(w, r, queryCategoryGet, "getting category")
}

func (h *CatalogHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	h.handleCommandWithID(w, r, cmdCategoryUpdate, "updating category")
}

func (h *CatalogHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	h.handleDelete(w, r, cmdCategoryDelete, "deleting category")
}

// --- Specification Handlers ---

func (h *CatalogHandler) ListSpecifications(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), catalogRPCTimeout)
	defer cancel()

	payload := map[string]any{}
	responseBytes, err := h.rpcClient.CallRPC(ctx, catalogExchange, querySpecList, payload, getHeaders(r))
	if err != nil {
		slog.Error("error listing specifications", "error", err)
		http.Error(w, "Failed to list specifications: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

func (h *CatalogHandler) CreateSpecification(w http.ResponseWriter, r *http.Request) {
	h.handleCommand(w, r, cmdSpecCreate, "creating specification")
}

func (h *CatalogHandler) GetSpecification(w http.ResponseWriter, r *http.Request) {
	h.handleQueryByID(w, r, querySpecGet, "getting specification")
}

func (h *CatalogHandler) UpdateSpecification(w http.ResponseWriter, r *http.Request) {
	h.handleCommandWithID(w, r, cmdSpecUpdate, "updating specification")
}

func (h *CatalogHandler) DeleteSpecification(w http.ResponseWriter, r *http.Request) {
	h.handleDelete(w, r, cmdSpecDelete, "deleting specification")
}

// --- Offering Handlers ---

func (h *CatalogHandler) ListOfferings(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), catalogRPCTimeout)
	defer cancel()

	payload := map[string]any{}
	responseBytes, err := h.rpcClient.CallRPC(ctx, catalogExchange, queryOfferingList, payload, getHeaders(r))
	if err != nil {
		slog.Error("error listing offerings", "error", err)
		http.Error(w, "Failed to list offerings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

func (h *CatalogHandler) CreateOffering(w http.ResponseWriter, r *http.Request) {
	h.handleCommand(w, r, cmdOfferingCreate, "creating offering")
}

func (h *CatalogHandler) GetOffering(w http.ResponseWriter, r *http.Request) {
	h.handleQueryByID(w, r, queryOfferingGet, "getting offering")
}

func (h *CatalogHandler) UpdateOffering(w http.ResponseWriter, r *http.Request) {
	h.handleCommandWithID(w, r, cmdOfferingUpdate, "updating offering")
}

func (h *CatalogHandler) DeleteOffering(w http.ResponseWriter, r *http.Request) {
	h.handleDelete(w, r, cmdOfferingDelete, "deleting offering")
}

// --- Helper Methods ---

func (h *CatalogHandler) handleCommand(w http.ResponseWriter, r *http.Request, topic, msg string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), catalogRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, catalogExchange, topic, payload, getHeaders(r))
	if err != nil {
		slog.Error("error "+msg, "error", err)
		http.Error(w, "Failed to "+msg+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(responseBytes)
}

func (h *CatalogHandler) handleCommandWithID(w http.ResponseWriter, r *http.Request, topic, msg string) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	payload["id"] = id

	ctx, cancel := context.WithTimeout(r.Context(), catalogRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, catalogExchange, topic, payload, getHeaders(r))
	if err != nil {
		slog.Error("error "+msg, "error", err)
		http.Error(w, "Failed to "+msg+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

func (h *CatalogHandler) handleQueryByID(w http.ResponseWriter, r *http.Request, topic, msg string) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": id}

	ctx, cancel := context.WithTimeout(r.Context(), catalogRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, catalogExchange, topic, payload, getHeaders(r))
	if err != nil {
		slog.Error("error "+msg, "error", err)
		http.Error(w, "Failed to "+msg+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

func (h *CatalogHandler) handleDelete(w http.ResponseWriter, r *http.Request, topic, msg string) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": id}

	ctx, cancel := context.WithTimeout(r.Context(), catalogRPCTimeout)
	defer cancel()

	_, err := h.rpcClient.CallRPC(ctx, catalogExchange, topic, payload, getHeaders(r))
	if err != nil {
		slog.Error("error "+msg, "error", err)
		http.Error(w, "Failed to "+msg+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
