package checker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"health-checker/internal/config"
)

// HealthStatus representa o status de saúde de um serviço
type HealthStatus struct {
	ServiceName  string        `json:"service_name"`
	URL          string        `json:"url"`
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code,omitempty"`
	Error        string        `json:"error,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
	Retries      int           `json:"retries_used"`
}

// HealthChecker gerencia as verificações de saúde
type HealthChecker struct {
	client *http.Client
	config *config.HealthCheckConfig
}

// NewHealthChecker cria uma nova instância do health checker
func NewHealthChecker(cfg *config.HealthCheckConfig) *HealthChecker {
	return &HealthChecker{
		client: &http.Client{},
		config: cfg,
	}
}

// CheckService verifica a saúde de um único serviço com retry
func (hc *HealthChecker) CheckService(ctx context.Context, service config.ServiceConfig) HealthStatus {
	status := HealthStatus{
		ServiceName: service.Name,
		URL:         service.URL,
		Timestamp:   time.Now(),
	}

	timeout := service.GetTimeout(hc.config.DefaultTimeout)
	maxRetries := service.GetRetries(hc.config.DefaultRetries)
	method := service.GetMethod()

	// Context com timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		start := time.Now()
		
		req, err := http.NewRequestWithContext(ctxWithTimeout, method, service.URL, nil)
		if err != nil {
			lastErr = err
			status.Retries = attempt
			continue
		}

		resp, err := hc.client.Do(req)
		status.ResponseTime = time.Since(start)
		
		if err != nil {
			lastErr = err
			status.Retries = attempt
			
			// Se não é o último retry, espera um pouco antes de tentar novamente
			if attempt < maxRetries {
				select {
				case <-time.After(time.Duration(attempt+1) * time.Second):
				case <-ctxWithTimeout.Done():
					lastErr = ctxWithTimeout.Err()
					break
				}
			}
			continue
		}

		defer resp.Body.Close()
		
		status.StatusCode = resp.StatusCode
		status.Retries = attempt
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			status.Status = "healthy"
			return status
		} else {
			lastErr = fmt.Errorf("status code não saudável: %d", resp.StatusCode)
			status.Status = "unhealthy"
			
			// Se não é o último retry, tenta novamente
			if attempt < maxRetries {
				select {
				case <-time.After(time.Duration(attempt+1) * time.Second):
				case <-ctxWithTimeout.Done():
					lastErr = ctxWithTimeout.Err()
					break
				}
			}
		}
	}

	// Se chegou aqui, todas as tentativas falharam
	status.Status = "unhealthy"
	if lastErr != nil {
		status.Error = lastErr.Error()
	}
	
	return status
}

// CheckAllServices verifica todos os serviços configurados de forma concorrente
func (hc *HealthChecker) CheckAllServices(ctx context.Context) []HealthStatus {
	services := hc.config.Services
	results := make([]HealthStatus, len(services))
	
	var wg sync.WaitGroup
	
	for i, service := range services {
		wg.Add(1)
		go func(index int, svc config.ServiceConfig) {
			defer wg.Done()
			results[index] = hc.CheckService(ctx, svc)
		}(i, service)
	}
	
	wg.Wait()
	return results
}

// CheckServicesWithProgress verifica todos os serviços e reporta progresso
func (hc *HealthChecker) CheckServicesWithProgress(ctx context.Context, progressChan chan<- HealthStatus) []HealthStatus {
	services := hc.config.Services
	results := make([]HealthStatus, len(services))
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	for i, service := range services {
		wg.Add(1)
		go func(index int, svc config.ServiceConfig) {
			defer wg.Done()
			
			status := hc.CheckService(ctx, svc)
			
			mu.Lock()
			results[index] = status
			mu.Unlock()
			
			// Envia progresso se o canal foi fornecido
			if progressChan != nil {
				select {
				case progressChan <- status:
				case <-ctx.Done():
					return
				}
			}
		}(i, service)
	}
	
	wg.Wait()
	
	if progressChan != nil {
		close(progressChan)
	}
	
	return results
}