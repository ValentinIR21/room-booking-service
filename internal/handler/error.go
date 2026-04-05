package handler

import (
	"avito-talk/internal/api"
	"encoding/json"
	"log"
	"net/http"
)

func httpStatus(code api.ErrorResponseErrorCode) int {
	switch code {
	case api.UNAUTHORIZED:
		return http.StatusUnauthorized
	case api.FORBIDDEN:
		return http.StatusForbidden
	case api.ROOMNOTFOUND, api.SLOTNOTFOUND, api.BOOKINGNOTFOUND, api.NOTFOUND:
		return http.StatusNotFound
	case api.SCHEDULEEXISTS, api.SLOTALREADYBOOKED:
		return http.StatusConflict
	case api.INVALIDREQUEST:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func writeError(w http.ResponseWriter, code api.ErrorResponseErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus(code))
	if err := json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    string(code),
			"message": message,
		},
	}); err != nil {
		log.Printf("writeError: failed to encode response: %v", err)
	}
}

// OapiErrorHandler обрабатывает ошибки парсинга параметров из oapi-codegen
func OapiErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, api.INVALIDREQUEST, err.Error())
}
