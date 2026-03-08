package mw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// ── CORS Middleware Tests ────────────────────────────────────────────

func TestCORS_AllowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:5173", "https://lixie.art"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("expected origin header 'http://localhost:5173', got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Allow-Methods header")
	}
	if got := w.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Errorf("expected Max-Age 86400, got %q", got)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:5173"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no origin header for disallowed origin, got %q", got)
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	handler := CORS([]string{"http://localhost:5173"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("GET", "/api/health", nil)
	// No Origin header set
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS headers without Origin, got %q", got)
	}
}

func TestCORS_PreflightOptions(t *testing.T) {
	innerCalled := false
	handler := CORS([]string{"http://localhost:5173"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			innerCalled = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("OPTIONS", "/api/brew", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("expected CORS headers on preflight, got %q", got)
	}
	if innerCalled {
		t.Error("inner handler should NOT be called for preflight")
	}
}

func TestCORS_PreflightDisallowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:5173"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("OPTIONS", "/api/brew", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// OPTIONS still returns 204 (it's a valid HTTP method), but no CORS headers
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS headers for disallowed preflight, got %q", got)
	}
}

// ── Request Logger Tests ────────────────────────────────────────────

func TestRequestLogger_PassesThrough(t *testing.T) {
	innerCalled := false
	handler := RequestLogger(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			innerCalled = true
			w.WriteHeader(http.StatusCreated)
		}),
	)

	req := httptest.NewRequest("POST", "/api/brew", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !innerCalled {
		t.Error("inner handler should be called")
	}
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestRequestLogger_CapturesStatusCode(t *testing.T) {
	handler := RequestLogger(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)

	req := httptest.NewRequest("GET", "/api/planets/missing", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestStatusWriter_DefaultStatus(t *testing.T) {
	w := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

	// Write body without explicit WriteHeader — should default to 200
	sw.Write([]byte("hello"))

	if sw.status != http.StatusOK {
		t.Errorf("expected default status 200, got %d", sw.status)
	}
}

func TestStatusWriter_CapturesExplicitStatus(t *testing.T) {
	w := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

	sw.WriteHeader(http.StatusTeapot)

	if sw.status != http.StatusTeapot {
		t.Errorf("expected 418, got %d", sw.status)
	}
}

func TestStatusWriter_OnlyFirstWriteHeaderCounts(t *testing.T) {
	w := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

	sw.WriteHeader(http.StatusNotFound)
	sw.WriteHeader(http.StatusOK) // second call — should not change status

	if sw.status != http.StatusNotFound {
		t.Errorf("expected 404 (first write), got %d", sw.status)
	}
}
