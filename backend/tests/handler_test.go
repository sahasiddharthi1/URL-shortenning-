package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/handler"
	"url-shortener/internal/service"
	"url-shortener/internal/store"

	"github.com/gorilla/mux"
)

func setupTestServer(t *testing.T) (*mux.Router, func()) {
	t.Helper()

	db, err := store.NewPostgresStore("localhost", "5432", "postgres", "postgres123", "urlshortener")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	urlService := service.NewURLService(db)
	urlHandler := handler.NewURLHandler(urlService)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/shorten", urlHandler.ShortenURL).Methods("POST")
	router.HandleFunc("/api/v1/stats/{code}", urlHandler.GetStats).Methods("GET")
	router.HandleFunc("/{code}", urlHandler.RedirectURL).Methods("GET")

	cleanup := func() {
		db.Close()
	}

	return router, cleanup
}

func TestShortenURL_Success(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]string{
		"long_url": "https://www.example.com/test",
	})

	req, _ := http.NewRequest("POST", "/api/v1/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response["short_url"] == "" {
		t.Error("Expected short_url in response")
	}
	if response["short_code"] == "" {
		t.Error("Expected short_code in response")
	}
}

func TestShortenURL_EmptyURL(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]string{
		"long_url": "",
	})

	req, _ := http.NewRequest("POST", "/api/v1/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestShortenURL_InvalidJSON(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest("POST", "/api/v1/shorten", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestRedirectURL_Found(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	// First, create a short URL
	body, _ := json.Marshal(map[string]string{
		"long_url": "https://www.example.com/redirect-test",
	})
	createReq, _ := http.NewRequest("POST", "/api/v1/shorten", bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResponse map[string]string
	json.Unmarshal(createRR.Body.Bytes(), &createResponse)
	shortCode := createResponse["short_code"]

	// Then, try to redirect
	redirectReq, _ := http.NewRequest("GET", "/"+shortCode, nil)
	redirectRR := httptest.NewRecorder()
	router.ServeHTTP(redirectRR, redirectReq)

	if redirectRR.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status 301, got %d", redirectRR.Code)
	}

	location := redirectRR.Header().Get("Location")
	if location != "https://www.example.com/redirect-test" {
		t.Errorf("Expected redirect to https://www.example.com/redirect-test, got %s", location)
	}
}

func TestRedirectURL_NotFound(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestGetStats_Found(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a short URL first
	body, _ := json.Marshal(map[string]string{
		"long_url": "https://www.example.com/stats-test",
	})
	createReq, _ := http.NewRequest("POST", "/api/v1/shorten", bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResponse map[string]string
	json.Unmarshal(createRR.Body.Bytes(), &createResponse)
	shortCode := createResponse["short_code"]

	// Get stats
	statsReq, _ := http.NewRequest("GET", "/api/v1/stats/"+shortCode, nil)
	statsRR := httptest.NewRecorder()
	router.ServeHTTP(statsRR, statsReq)

	if statsRR.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", statsRR.Code)
	}

	var statsResponse map[string]interface{}
	json.Unmarshal(statsRR.Body.Bytes(), &statsResponse)

	stats, ok := statsResponse["stats"].(map[string]interface{})
	if !ok {
		t.Error("Expected stats object in response")
	}

	if stats["short_code"] != shortCode {
		t.Errorf("Expected short_code %s, got %v", shortCode, stats["short_code"])
	}
}

func TestGetStats_NotFound(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", "/api/v1/stats/nonexistent", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}
