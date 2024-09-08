package config

type APIConfig struct {
	Env string

	Host string
	Port int
}

func NewAPIConfig() *APIConfig {
	return &APIConfig{}
}
