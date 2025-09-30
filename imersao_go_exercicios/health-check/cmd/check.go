/*
Copyright © 2025 Gleicon Moraes <gleicon@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"health-checker/internal/checker"
	"health-checker/internal/config"

	"github.com/spf13/cobra"
)

var (
	configFile string
	verbose    bool
	realTime   bool
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Executa verificações de saúde nos serviços configurados",
	Long: `Executa verificações de saúde concorrentes em múltiplos serviços HTTP/HTTPS.

O comando 'check' lê a configuração de um arquivo JSON e verifica o status
de todos os serviços definidos simultaneamente, implementando timeouts
configuráveis e lógica de retry.

Exemplo de uso:
  health-checker check --config services.json
  health-checker check --config services.json --verbose
  health-checker check --config services.json --realtime`,
	Run: runHealthCheck,
}

func runHealthCheck(cmd *cobra.Command, args []string) {
	if configFile == "" {
		fmt.Fprintf(os.Stderr, "Erro: arquivo de configuração é obrigatório. Use --config <arquivo>\n")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao carregar configuração: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("Carregando configuração de: %s\n", configFile)
		fmt.Printf("Verificando %d serviços...\n\n", len(cfg.Services))
	}

	healthChecker := checker.NewHealthChecker(cfg)
	ctx := context.Background()

	var results []checker.HealthStatus

	if realTime {
		progressChan := make(chan checker.HealthStatus, len(cfg.Services))
		// uma goroutine por check
		go func() {
			results = healthChecker.CheckServicesWithProgress(ctx, progressChan)
		}()

		completed := 0
		for status := range progressChan {
			completed++
			if verbose {
				fmt.Printf("[%d/%d] %s: %s (%s)\n",
					completed, len(cfg.Services),
					status.ServiceName,
					status.Status,
					status.ResponseTime.Round(time.Millisecond))
			}
		}
	} else {
		if verbose {
			fmt.Println("Executando verificações...")
		}
		results = healthChecker.CheckAllServices(ctx)
	}

	outputResults(results, verbose)
}

func outputResults(results []checker.HealthStatus, verbose bool) {
	if verbose {
		fmt.Println("\n=== STATUS ===")
	}

	healthyCount := 0
	for _, result := range results {
		if result.Status == "healthy" {
			healthyCount++
		}
	}

	report := map[string]interface{}{
		"timestamp":          time.Now(),
		"total_services":     len(results),
		"healthy_services":   healthyCount,
		"unhealthy_services": len(results) - healthyCount,
		"services":           results,
	}

	jsonOutput, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao gerar JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))

	if verbose {
		fmt.Printf("\nResumo: %d/%d serviços saudáveis\n", healthyCount, len(results))
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringVarP(&configFile, "config", "c", "", "Arquivo de configuração JSON (obrigatório)")
	checkCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Output verboso com informações detalhadas")
	checkCmd.Flags().BoolVarP(&realTime, "realtime", "r", false, "Mostra resultados em tempo real conforme completam")

	checkCmd.MarkFlagRequired("config")
}
