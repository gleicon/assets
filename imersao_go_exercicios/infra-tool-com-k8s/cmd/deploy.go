package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
)

type DeployConfig struct {
	Name     string            `json:"name"`
	Image    string            `json:"image"`
	Ports    map[string]string `json:"ports,omitempty"`
	Env      []string          `json:"env,omitempty"`
	Volumes  []string          `json:"volumes,omitempty"`
	Command  []string          `json:"command,omitempty"`
	Restart  string            `json:"restart,omitempty"`
	Networks []string          `json:"networks,omitempty"`
}

var deployCmd = &cobra.Command{
	Use:   "deploy <config.json>",
	Short: "Faz deploy de container a partir de arquivo de configuração",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := args[0]
		
		file, err := os.Open(configFile)
		if err != nil {
			fmt.Printf("Erro ao abrir arquivo de configuração %s: %v\n", configFile, err)
			return
		}
		defer file.Close()

		var config DeployConfig
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&config); err != nil {
			fmt.Printf("Erro ao decodificar arquivo de configuração: %v\n", err)
			return
		}

		ctx := context.Background()
		cli, err := NewDockerClient()
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		defer cli.Close()

		fmt.Printf("=== Iniciando deploy do container %s ===\n", config.Name)

		// Pull da imagem
		fmt.Printf("Baixando imagem %s...\n", config.Image)
		reader, err := cli.ImagePull(ctx, config.Image, image.PullOptions{})
		if err != nil {
			fmt.Printf("Erro ao baixar imagem: %v\n", err)
			return
		}
		io.Copy(os.Stdout, reader)
		reader.Close()

		// Parar container existente se houver
		fmt.Printf("Verificando se container %s já existe...\n", config.Name)
		containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
		if err == nil {
			for _, c := range containers {
				for _, name := range c.Names {
					if strings.TrimPrefix(name, "/") == config.Name {
						fmt.Printf("Parando container existente %s...\n", config.Name)
						cli.ContainerStop(ctx, c.ID, container.StopOptions{})
						cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{})
						break
					}
				}
			}
		}

		// Configurar portas
		exposedPorts := make(nat.PortSet)
		portBindings := make(nat.PortMap)
		
		for containerPort, hostPort := range config.Ports {
			port, err := nat.NewPort("tcp", strings.Split(containerPort, "/")[0])
			if err != nil {
				fmt.Printf("Erro ao configurar porta %s: %v\n", containerPort, err)
				continue
			}
			exposedPorts[port] = struct{}{}
			portBindings[port] = []nat.PortBinding{
				{HostPort: hostPort},
			}
		}

		// Criar container
		containerConfig := &container.Config{
			Image:        config.Image,
			Env:          config.Env,
			Cmd:          config.Command,
			ExposedPorts: exposedPorts,
		}

		hostConfig := &container.HostConfig{
			PortBindings: portBindings,
			Binds:        config.Volumes,
		}

		if config.Restart != "" {
			hostConfig.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(config.Restart)}
		}

		networkingConfig := &network.NetworkingConfig{}

		resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, config.Name)
		if err != nil {
			fmt.Printf("Erro ao criar container: %v\n", err)
			return
		}

		// Iniciar container
		fmt.Printf("Iniciando container %s...\n", config.Name)
		if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			fmt.Printf("Erro ao iniciar container: %v\n", err)
			return
		}

		fmt.Printf("✓ Container %s iniciado com sucesso! ID: %s\n", config.Name, resp.ID[:12])

		// Conectar às redes especificadas
		for _, networkName := range config.Networks {
			if err := cli.NetworkConnect(ctx, networkName, resp.ID, nil); err != nil {
				fmt.Printf("Aviso: Erro ao conectar à rede %s: %v\n", networkName, err)
			} else {
				fmt.Printf("✓ Container conectado à rede %s\n", networkName)
			}
		}
	},
}

func init() {
	containerCmd.AddCommand(deployCmd)
}