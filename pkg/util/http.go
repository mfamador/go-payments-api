package util

import (
	"encoding/json"
	"fmt"
	"github.com/mfamador/go-payments-api/pkg/logger"
	"net/http"
	"strconv"
)

type HttpService struct {
	BaseUrl string
}

func (s *HttpService) UrlFor(path string) string {
	return fmt.Sprintf("%s%s", s.BaseUrl, path)
}

type EmptyResponse struct{}

func HandleHttpError(w http.ResponseWriter, r *http.Request, status int, err error) {
	RenderJSON(w, r, status, &EmptyResponse{})
	logger.Error(err)
}

func RenderJSON(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(data)
	if err != nil {
		logger.Error(err)
	}
}

func RenderNoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func IntFromStringOrDefault(actual string, defaultValue int) int {
	if actual == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(actual)
	if err != nil {
		return defaultValue
	}
	return v
}
