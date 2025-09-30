package cmd

import (
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var (
	dockerSocket string
)

var rootCmd = &cobra.Command{
	Use:   "infra-tool",
	Short: "Ferramenta de infraestrutura para gerenciar containers Docker",
	Long:  "Uma aplicação CLI em Go para gerenciar containers Docker usando a biblioteca Moby",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao executar comando: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add global Docker socket flag
	rootCmd.PersistentFlags().StringVar(&dockerSocket, "docker-socket", getDefaultDockerSocket(), "Docker socket path")
}

// getDefaultDockerSocket returns the default Docker socket path
// Can be overridden by DOCKER_HOST environment variable
func getDefaultDockerSocket() string {
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		return host
	}
	return "unix:///var/run/docker.sock"
}

// GetDockerSocket returns the configured Docker socket
func GetDockerSocket() string {
	if dockerSocket == "" {
		dockerSocket = getDefaultDockerSocket()
	}
	return dockerSocket
}

// NewDockerClient creates a new Docker client with the configured socket
func NewDockerClient() (*client.Client, error) {
	socket := GetDockerSocket()
	
	// Create client with custom socket
	cli, err := client.NewClientWithOpts(
		client.WithHost(socket),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com Docker no socket %s: %v", socket, err)
	}
	
	return cli, nil
}