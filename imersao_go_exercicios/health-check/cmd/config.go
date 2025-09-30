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
	"fmt"
	"os"

	"health-checker/internal/config"
	"github.com/spf13/cobra"
)

var outputFile string

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gera um arquivo de configuração de exemplo",
	Long: `Gera um arquivo de configuração de exemplo com serviços predefinidos.

O comando 'config' cria um arquivo JSON com configurações de exemplo
para múltiplos serviços, incluindo timeouts, retries e outros parâmetros.

Exemplo de uso:
  health-checker config --output services.json
  health-checker config -o my-services.json`,
	Run: generateConfig,
}

func generateConfig(cmd *cobra.Command, args []string) {
	if outputFile == "" {
		outputFile = "health-check-config.json"
	}

	exampleConfig := &config.HealthCheckConfig{
		Services: []config.ServiceConfig{
			{
				Name:    "Google",
				URL:     "https://www.google.com",
				Method:  "GET",
				Timeout: "5s",
				Retries: 2,
			},
			{
				Name:    "GitHub API",
				URL:     "https://api.github.com",
				Method:  "GET",
				Timeout: "10s",
				Retries: 3,
			},
			{
				Name:    "JSONPlaceholder",
				URL:     "https://jsonplaceholder.typicode.com/posts/1",
				Timeout: "8s",
				Retries: 2,
			},
			{
				Name:    "HTTPBin Status",
				URL:     "https://httpbin.org/status/200",
				Timeout: "6s",
				Retries: 1,
			},
			{
				Name:    "Example.com",
				URL:     "https://example.com",
				Timeout: "5s",
				Retries: 2,
			},
			{
				Name:    "Duck Duck Go",
				URL:     "https://duckduckgo.com",
				Timeout: "7s",
				Retries: 2,
			},
			{
				Name:    "Stack Overflow",
				URL:     "https://stackoverflow.com",
				Timeout: "10s",
				Retries: 3,
			},
			{
				Name:    "Reddit",
				URL:     "https://www.reddit.com",
				Timeout: "8s",
				Retries: 2,
			},
			{
				Name:    "Wikipedia",
				URL:     "https://www.wikipedia.org",
				Timeout: "6s",
				Retries: 2,
			},
			{
				Name:    "Docker Hub",
				URL:     "https://hub.docker.com",
				Timeout: "9s",
				Retries: 3,
			},
			{
				Name:    "Hacker News",
				URL:     "https://news.ycombinator.com",
				Timeout: "8s",
				Retries: 2,
			},
		},
		DefaultTimeout: "10s",
		DefaultRetries: 3,
		OutputFormat:   "json",
	}

	err := config.SaveConfig(exampleConfig, outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao salvar configuração: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Arquivo de configuração gerado com sucesso: %s\n", outputFile)
	fmt.Printf("Configuração inclui %d serviços de exemplo.\n", len(exampleConfig.Services))
	fmt.Printf("\nPara executar as verificações:\n")
	fmt.Printf("  health-checker check --config %s --verbose\n", outputFile)
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVarP(&outputFile, "output", "o", "health-check-config.json", "Nome do arquivo de configuração a ser criado")
}
