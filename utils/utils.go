package utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
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

func LoadEnvs() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}

func GetEnvAsBool(env string) bool {
	envValue := os.Getenv(env)

	value, err := strconv.ParseBool(envValue)

	if err != nil {
		log.Printf("[ERROR] Error parsing env variable `%s` with value `%s` to bool: %v", env, envValue, err)
	}

	return value
}

func GetEnvAsInt(env string) int {
	envValue := os.Getenv(env)

	value, err := strconv.Atoi(envValue)

	if err != nil {
		log.Printf("[ERROR] Error parsing env variable `%s` with value `%s` to int: %v", env, envValue, err)
	}

	return value
}
