package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/LilyEssence/cosmic-potions/internal/brewing"
	"github.com/LilyEssence/cosmic-potions/internal/store"
)

// ── Response Wrapper ────────────────────────────────────────────────
//
// GO CONCEPT: Unexported Types
// apiResponse and apiError start with lowercase — they're unexported (private
// to this package). External code can't construct or reference them directly.
// But JSON serialization still works because encoding/json uses reflection to
// read struct tags, and reflection can see unexported types within the package
// that created them.
//
// Why unexport these? They're implementation details of how we format responses.
// No other package needs to create an apiResponse directly — they use the
// writeJSON and writeError helper functions instead.

// apiResponse wraps all API responses in a consistent envelope.
// Success: {"data": ...}  Error: {"error": {"code": 404, "message": "..."}}
type apiResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *apiError   `json:"error,omitempty"`
}

// apiError carries machine-readable code and human-readable message.
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ── JSON Helpers ────────────────────────────────────────────────────

// writeJSON sends a success response with the given status code and data.
//
// GO CONCEPT: json.NewEncoder (Streaming Encoder)
// json.NewEncoder(w) creates an encoder that writes directly to w (our
// http.ResponseWriter). This is more efficient than json.Marshal() + w.Write()
// because it avoids allocating an intermediate []byte buffer — the JSON is
// written directly to the response stream.
//
// The trade-off: if encoding fails partway through, some bytes may have already
// been sent to the client. For a game API, this is acceptable. For financial
// APIs, you'd marshal to a buffer first, check for errors, then write.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(apiResponse{Data: data}); err != nil {
		// At this point headers are already sent — we can only log.
		log.Printf("JSON encode error: %v", err)
	}
}

// writeError sends an error response with the given status code and message.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(apiResponse{
		Error: &apiError{Code: status, Message: message},
	}); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

// readJSON decodes a JSON request body into dst.
//
// GO CONCEPT: json.NewDecoder (Streaming Decoder)
// Like the encoder, the decoder reads directly from r.Body (an io.Reader)
// without buffering the entire body into memory first. DisallowUnknownFields()
// makes the decoder reject JSON with keys that don't match any struct field —
// this catches typos in API requests early (e.g., "ingedientIds" instead of
// "ingredientIds").
//
// GO CONCEPT: interface{} Parameter
// The `dst interface{}` parameter accepts any type — similar to TypeScript's
// `any`. The json decoder uses reflection to figure out what fields dst has
// and fills them in. The caller passes a pointer: readJSON(r, &myStruct).
func readJSON(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// ── Error Mapping ───────────────────────────────────────────────────
//
// GO CONCEPT: errors.Is for Sentinel Matching
// errors.Is checks if an error (or any error in its chain) matches a target
// sentinel value. This works even when errors are wrapped with fmt.Errorf:
//
//   original := brewing.ErrTooFewIngredients
//   wrapped  := fmt.Errorf("validation: %w", original)
//   errors.Is(wrapped, brewing.ErrTooFewIngredients)  // true
//
// The %w verb in fmt.Errorf wraps the error, creating a chain. errors.Is
// unwraps the chain to find matches. This is why we use sentinel errors
// instead of string comparison — wrapping preserves identity.

// mapError translates known domain errors into appropriate HTTP responses.
// Unknown errors become 500 Internal Server Error with a generic message
// (never leak internal details to the client).
func mapError(w http.ResponseWriter, err error, entityName string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, entityName+" not found")
	case errors.Is(err, brewing.ErrTooFewIngredients):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, brewing.ErrTooManyIngredients):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		log.Printf("unexpected error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
