package main

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"net/http"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxByte := 1 << 20 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxByte))
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(data); err != nil {
		return err
	}
	return nil
}

func writeJSONError(w http.ResponseWriter, status int, message string) error {
	data := map[string]string{"error": message}
	return writeJSON(w, status, data)

}

func jsonResponse(w http.ResponseWriter, status int, data any, msg ...string) error {
	type response struct {
		Data    any    `json:"data,omitempty"`
		Message string `json:"message,omitempty"`
	}

	var message string
	if len(msg) > 0 {
		message = msg[0]
	}

	var responseData any
	if data != nil {
		responseData = data
	}

	return writeJSON(w, status, response{Data: responseData, Message: message})
}

func errorResponse(w http.ResponseWriter, status int, message string) error {
	type errorResp struct {
		Error string `json:"error"`
	}
	return writeJSON(w, status, errorResp{Error: message})
}

func formatValidationErrors(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorMessages := make([]string, 0)

		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()

			switch tag {
			case "required":
				errorMessages = append(errorMessages, field+" is required")
			case "email":
				errorMessages = append(errorMessages, field+" must be a valid email address")
			case "min":
				errorMessages = append(errorMessages, field+" must be at least "+e.Param()+" characters long")
			default:
				errorMessages = append(errorMessages, field+" is invalid: "+tag)
			}
		}

		if len(errorMessages) > 0 {
			return errorMessages[0] // Return the first error for simplicity
		}
	}

	return err.Error()
}
