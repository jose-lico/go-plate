package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var Validate = validator.New()

func ParseJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	if r.Body == http.NoBody {
		return errors.New("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(v)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func GetEnvAsBool(env string) bool {
	envValue := os.Getenv(env)

	if envValue == "" {
		return false
	}

	value, err := strconv.ParseBool(envValue)

	if err != nil {
		zap.L().Warn("Error parsing env variable to bool",
			zap.String("env", env),
			zap.String("value", envValue))
	}

	return value
}

func GetEnvAsInt(env string) int {
	envValue := os.Getenv(env)

	if envValue == "" {
		return 0
	}

	value, err := strconv.Atoi(envValue)

	if err != nil {
		zap.L().Warn("Error parsing env variable to int",
			zap.String("env", env),
			zap.String("value", envValue))
	}

	return value
}
