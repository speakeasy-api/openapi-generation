package ecommerce

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gorilla/mux"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/middleware"
)

func Ptr[T any](t T) *T {
	return &t
}

func HandleListProducts(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	scopes, scopesFound := middleware.OAuth2Scopes(r)
	if !scopesFound || !scopes.Has([]string{"products:read"}) {
		http.Error(rw, `{"error": "insufficient scopes"}`, http.StatusForbidden)
		return
	}

	cursorQ := r.URL.Query().Get("cursor")
	if cursorQ == "" {
		cursorQ = "0"
	}

	cursor, err := strconv.ParseUint(string(cursorQ), 10, 64)
	if err != nil {
		http.Error(rw, `{"error": "cursor not a uint64"}`, http.StatusBadRequest)
		return
	}

	if cursor%10 != 0 {
		http.Error(rw, `{"error": "cursor not a multiple of 10"}`, http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(rw)
	enc.SetIndent("", "  ")

	if cursor > 30 {
		err := enc.Encode(ProductList{Products: []*Product{}})
		if err != nil {
			http.Error(rw, `{"error": "could not encode response"}`, http.StatusInternalServerError)
		}
		return
	}

	next := cursor + 10
	var nextCursorBody *string
	if next <= 30 {
		nextCursorBody = Ptr(fmt.Sprintf("%d", next))
	}

	// Seed cannot be 0 otherwise faker picks a random one
	faker := gofakeit.New(cursor + 1)

	products := make([]*Product, 10)
	for i := 0; i < 10; i++ {
		p := faker.Product()

		created := faker.DateRange(
			time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		).Truncate(24 * time.Hour)

		updated := faker.DateRange(
			created,
			time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		).Truncate(24 * time.Hour)

		products[i] = &Product{
			ID:        p.UPC,
			Name:      p.Name,
			Price:     p.Price,
			CreatedAt: created,
			UpdatedAt: updated,
		}
	}

	err = enc.Encode(ProductList{
		NextCursor: nextCursorBody,
		Products:   products,
	})
	if err != nil {
		http.Error(rw, `{"error": "could not encode response"}`, http.StatusInternalServerError)
	}
}

func HandleCreateProduct(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	scopes, scopesFound := middleware.OAuth2Scopes(r)
	if !scopesFound || !scopes.Has([]string{"products:create"}) {
		http.Error(rw, `{"error": "insufficient scopes"}`, http.StatusForbidden)
		return
	}

	enc := json.NewEncoder(rw)
	enc.SetIndent("", "  ")

	defer r.Body.Close()
	var p NewProductForm
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(rw, `{"error": "could not decode request"}`, http.StatusBadRequest)
		return
	}

	faker := gofakeit.New(0)

	now := time.Now().Truncate(24 * time.Second)
	if err := enc.Encode(Product{
		ID:          faker.ProductUPC(),
		Name:        p.Name,
		Price:       p.Price,
		Description: p.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		http.Error(rw, `{"error": "could not encode response"}`, http.StatusInternalServerError)
	}
}

func HandleFetchProduct(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	scopes, scopesFound := middleware.OAuth2Scopes(r)
	if !scopesFound || !scopes.Has([]string{"products:read"}) {
		http.Error(rw, `{"error": "insufficient scopes"}`, http.StatusForbidden)
		return
	}

	enc := json.NewEncoder(rw)
	enc.SetIndent("", "  ")

	vars := mux.Vars(r)
	rawID, ok := vars["id"]
	if !ok {
		http.Error(rw, `{"error": "{id} is required"}`, http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		http.Error(rw, `{"error": "{id} must a uint64"}`, http.StatusBadRequest)
		return
	}

	// Seed cannot be 0 otherwise faker picks a random one
	faker := gofakeit.New(productID + 1)
	p := faker.Product()
	created := faker.DateRange(
		time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	).Truncate(24 * time.Hour)
	updated := faker.DateRange(
		created,
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	).Truncate(24 * time.Hour)

	if err := enc.Encode(Product{
		ID:        rawID,
		Name:      p.Name,
		Price:     p.Price,
		CreatedAt: created,
		UpdatedAt: updated,
	}); err != nil {
		http.Error(rw, `{"error": "could not encode response"}`, http.StatusInternalServerError)
	}
}

func HandleUpdateProduct(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	scopes, scopesFound := middleware.OAuth2Scopes(r)
	if !scopesFound || !scopes.Has([]string{"products:read", "products:create"}) {
		http.Error(rw, `{"error": "insufficient scopes"}`, http.StatusForbidden)
		return
	}

	enc := json.NewEncoder(rw)
	enc.SetIndent("", "  ")

	vars := mux.Vars(r)
	rawID, ok := vars["id"]
	if !ok {
		http.Error(rw, `{"error": "{id} is required"}`, http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		http.Error(rw, `{"error": "{id} must a uint64"}`, http.StatusBadRequest)
		return
	}

	var form ProductUpdateForm
	err = json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		http.Error(rw, `{"error": "could not decode request"}`, http.StatusBadRequest)
		return
	}

	// Seed cannot be 0 otherwise faker picks a random one
	faker := gofakeit.New(productID + 1)
	p := faker.Product()
	created := faker.DateRange(
		time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	).Truncate(24 * time.Hour)
	updated := faker.DateRange(
		created,
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	).Truncate(24 * time.Hour)

	name := p.Name
	if form.Name != nil {
		name = *form.Name
	}
	description := p.Description
	if form.Description != nil {
		description = *form.Description
	}
	price := p.Price
	if form.Price != nil {
		price = *form.Price
	}

	if err := enc.Encode(Product{
		ID:          rawID,
		Name:        name,
		Price:       price,
		Description: description,
		CreatedAt:   created,
		UpdatedAt:   updated,
	}); err != nil {
		http.Error(rw, `{"error": "could not encode response"}`, http.StatusInternalServerError)
	}
}

func HandleDeleteProduct(rw http.ResponseWriter, r *http.Request) {
	scopes, scopesFound := middleware.OAuth2Scopes(r)
	if !scopesFound || !scopes.Has([]string{"products:delete"}) {
		rw.Header().Set("Content-Type", "application/json")
		http.Error(rw, `{"error": "insufficient scopes"}`, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	rawID, ok := vars["id"]
	if !ok {
		http.Error(rw, `{"error": "{id} is required"}`, http.StatusBadRequest)
		return
	}

	_, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		http.Error(rw, `{"error": "{id} must a uint64"}`, http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func HandleUpdateProductStock(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	scopes, scopesFound := middleware.OAuth2Scopes(r)
	if !scopesFound || !scopes.Has([]string{"admin"}) {
		http.Error(rw, `{"error": "insufficient scopes"}`, http.StatusForbidden)
		return
	}

	enc := json.NewEncoder(rw)
	enc.SetIndent("", "  ")

	vars := mux.Vars(r)
	rawID, ok := vars["id"]
	if !ok {
		http.Error(rw, `{"error": "{id} is required"}`, http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		http.Error(rw, `{"error": "{id} must a uint64"}`, http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var form ProductInventoryUpdateForm
	err = json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		http.Error(rw, `{"error": "could not decode request"}`, http.StatusBadRequest)
		return
	}

	// Seed cannot be 0 otherwise faker picks a random one
	faker := gofakeit.New(productID + 1)

	now := time.Now().Truncate(24 * time.Second)
	if err := enc.Encode(ProductInventoryStatus{
		ProductID: rawID,
		Quantity:  faker.IntRange(int(math.Abs(float64(form.QuantityDelta))), 100) + form.QuantityDelta,
		UpdatedAt: now,
	}); err != nil {
		http.Error(rw, `{"error": "could not encode response"}`, http.StatusInternalServerError)
	}
}
