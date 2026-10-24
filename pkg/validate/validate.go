package validate

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/oche/diagram/pkg/response"
)

var v = validator.New(validator.WithRequiredStructEnabled())

// DecodeAndValidate decodes a JSON request body into dst and validates struct tags.
// Returns true if decoding and validation succeeded, false otherwise (error already written to w).
func DecodeAndValidate(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return false
	}

	if err := v.Struct(dst); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			response.Error(w, http.StatusBadRequest, "validation failed")
			return false
		}

		messages := make([]string, 0, len(validationErrors))
		for _, fe := range validationErrors {
			messages = append(messages, formatFieldError(fe))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error":   "validation failed",
			"details": messages,
		})
		return false
	}

	return true
}

func formatFieldError(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + fe.Param() + " characters"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	case "oneof":
		return field + " must be one of: " + fe.Param()
	default:
		return field + " failed " + fe.Tag() + " validation"
	}
}
