package web

import (
	"encoding/json"
	"errors"
	"family-catering/pkg/apperrors"
	"net/http"
	"time"

	log "family-catering/pkg/logger"
)

type JSONResponse struct {
	Success     bool          `json:"success" example:"true"`
	Status      string        `json:"status"`
	Data        interface{}   `json:"data,omitempty" extensions:"x-omitempty"`
	ProcessTime time.Duration `json:"process_time"`
}

type Error struct {
	Message string `json:"message"`
}

type ErrJSONResponse struct {
	Success     bool          `json:"success" example:"false"`
	Status      string        `json:"status"`
	ProcessTime time.Duration `json:"process_time"`
	Error       Error         `json:"error,omitempty"`
}

func WriteSuccessJSON(w http.ResponseWriter, payload interface{}, startRequestTime time.Time) {
	w.Header().Set("Content-Type", "application/json")
	body := JSONResponse{
		Status:      "success",
		Success:     true,
		Data:        payload,
		ProcessTime: time.Since(startRequestTime),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

func WriteFailJSON(w http.ResponseWriter, statusCode int, message string, startRequestTime time.Time) {
	w.Header().Set("Content-Type", "application/json")
	body := ErrJSONResponse{
		Success:     false,
		Status:      "fail",
		Error:       Error{Message: message},
		ProcessTime: time.Since(startRequestTime),
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&body)
}

func WriteErrJSON(w http.ResponseWriter, statusCode int, message string, startRequestTime time.Time) {
	w.Header().Set("Content-Type", "application/json")
	body := ErrJSONResponse{
		Success:     false,
		Status:      "error",
		Error:       Error{Message: message},
		ProcessTime: time.Since(startRequestTime),
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&body)
}

func WriteHTTPError(w http.ResponseWriter, err error, start time.Time) {
	var (
		statusCode int
		statusText string
		apiErr     apperrors.APIError
	)

	if errors.As(err, &apiErr) {
		statusCode, statusText = apiErr.APIError()
		log.ErrorWithCause(err, apperrors.Cause(err), statusText)
		WriteFailJSON(w, statusCode, statusText, start)
		return

	}

	if !errors.Is(err, nil) {
		log.Error(err, http.StatusText(http.StatusInternalServerError))
		WriteErrJSON(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), start)
		return
	}
}

func WriteTooManyRequest(w http.ResponseWriter, r *http.Request) {
	start := RequestStartTimeFromContext(r.Context())
	log.Info("too many request")
	WriteErrJSON(w, http.StatusTooManyRequests, "too many request please try again later", start)
}
