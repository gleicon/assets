package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ServiceConfig representa a configuração de um serviço a ser monitorado
type ServiceConfig struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Method   string `json:"method,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
	Retries  int    `json:"retries,omitempty"`
	Interval string `json:"interval,omitempty"`
}

// HealthCheckConfig representa a configuração geral do health checker
type HealthCheckConfig struct {
	Services       []ServiceConfig `json:"services"`
	DefaultTimeout string          `json:"default_timeout,omitempty"`
	DefaultRetries int             `json:"default_retries,omitempty"`
	OutputFormat   string          `json:"output_format,omitempty"`
}

// LoadConfig carrega a configuração do arquivo especificado
func LoadConfig(filename string) (*HealthCheckConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}

	var config HealthCheckConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("erro ao parsear configuração JSON: %w", err)
	}

	// Define valores padrão se não especificados
	if config.DefaultTimeout == "" {
		config.DefaultTimeout = "10s"
	}
	if config.DefaultRetries == 0 {
		config.DefaultRetries = 3
	}
	if config.OutputFormat == "" {
		config.OutputFormat = "json"
	}

	return &config, nil
}

// SaveConfig salva a configuração no arquivo especificado
func SaveConfig(config *HealthCheckConfig, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar configuração: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("erro ao escrever arquivo de configuração: %w", err)
	}

	return nil
}

// GetTimeout retorna o timeout para um serviço específico ou o padrão
func (s *ServiceConfig) GetTimeout(defaultTimeout string) time.Duration {
	timeout := s.Timeout
	if timeout == "" {
		timeout = defaultTimeout
	}
	
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		return 10 * time.Second // fallback padrão
	}
	
	return duration
}

// GetRetries retorna o número de tentativas para um serviço específico ou o padrão
func (s *ServiceConfig) GetRetries(defaultRetries int) int {
	if s.Retries == 0 {
		return defaultRetries
	}
	return s.Retries
}

// GetMethod retorna o método HTTP para um serviço específico ou GET como padrão
func (s *ServiceConfig) GetMethod() string {
	if s.Method == "" {
		return "GET"
	}
	return s.Method
}