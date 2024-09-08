package config

import (
	"log"
	"os"
	"strconv"
)

func getEnvAsBool(env string) bool {
	envValue := os.Getenv(env)

	value, err := strconv.ParseBool(envValue)

	if err != nil {
		log.Printf("[ERROR] Error parsing env variable `%s` with value `%s` to bool: %v", env, envValue, err)
	}

	return value
}

func getEnvAsInt(env string) int {
	envValue := os.Getenv(env)

	value, err := strconv.Atoi(envValue)

	if err != nil {
		log.Printf("[ERROR] Error parsing env variable `%s` with value `%s` to int: %v", env, envValue, err)
	}

	return value
}
