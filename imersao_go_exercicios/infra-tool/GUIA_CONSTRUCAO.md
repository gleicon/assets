# Guia Passo a Passo: Construindo uma CLI com Cobra

Este guia mostra como construir o `infra-tool` do zero usando o `cobra-cli` e as melhores práticas.

## Pré-requisitos

- Go 1.19+ instalado
- Docker instalado e rodando
- Git instalado

## Passo 1: Configuração Inicial

### 1.1 Instalar o Cobra CLI
```bash
go install github.com/spf13/cobra-cli@latest
```

### 1.2 Criar o projeto
```bash
mkdir infra-tool
cd infra-tool
go mod init infra-tool
```

### 1.3 Inicializar com Cobra CLI
```bash
cobra-cli init
```

Isso criará a estrutura básica:
```
infra-tool/
├── cmd/
│   └── root.go
├── main.go
└── go.mod
```

## Passo 2: Configurar o Comando Root

### 2.1 Editar cmd/root.go
```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
```

## Passo 3: Adicionar Comando Container

### 3.1 Gerar comando container
```bash
cobra-cli add container
```

### 3.2 Editar cmd/container.go
```go
package cmd

import (
	"github.com/spf13/cobra"
)

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Gerencia containers Docker",
	Long:  "Comandos para listar, executar, fazer deploy e verificar saúde de containers",
}

func init() {
	rootCmd.AddCommand(containerCmd)
}
```

## Passo 4: Adicionar Subcomandos

### 4.1 Gerar subcomandos
```bash
cobra-cli add list -p 'containerCmd'
cobra-cli add logs -p 'containerCmd'
cobra-cli add exec -p 'containerCmd'
cobra-cli add health -p 'containerCmd'
cobra-cli add deploy -p 'containerCmd'
```

Isso criará os arquivos:
- `cmd/list.go`
- `cmd/logs.go`
- `cmd/exec.go`
- `cmd/health.go`
- `cmd/deploy.go`

## Passo 5: Adicionar Dependências Docker

### 5.1 Instalar bibliotecas
```bash
go get github.com/docker/docker@latest
go get github.com/docker/go-connections@latest
go mod tidy
```

## Passo 6: Implementar Comando List

### 6.1 Editar cmd/list.go
```go
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista containers em execução",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Erro ao conectar com Docker: %v\n", err)
			return
		}
		defer cli.Close()

		containers, err := cli.ContainerList(ctx, container.ListOptions{})
		if err != nil {
			fmt.Printf("Erro ao listar containers: %v\n", err)
			return
		}

		if len(containers) == 0 {
			fmt.Println("Nenhum container em execução")
			return
		}

		fmt.Printf("%-20s %-30s %-15s %-10s\n", "ID", "NOME", "IMAGEM", "STATUS")
		fmt.Println(strings.Repeat("-", 80))
		
		for _, c := range containers {
			name := strings.TrimPrefix(c.Names[0], "/")
			fmt.Printf("%-20s %-30s %-15s %-10s\n", 
				c.ID[:12], name, c.Image, c.State)
		}
	},
}

func init() {
	containerCmd.AddCommand(listCmd)
}
```

## Passo 7: Implementar Comando Logs

### 7.1 Editar cmd/logs.go
```go
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <nome>",
	Short: "Exibe logs do container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Erro ao conectar com Docker: %v\n", err)
			return
		}
		defer cli.Close()

		options := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Tail:       "100",
		}

		logs, err := cli.ContainerLogs(ctx, containerName, options)
		if err != nil {
			fmt.Printf("Erro ao obter logs do container %s: %v\n", containerName, err)
			return
		}
		defer logs.Close()

		fmt.Printf("=== Logs do container %s ===\n", containerName)
		io.Copy(os.Stdout, logs)
	},
}

func init() {
	containerCmd.AddCommand(logsCmd)
}
```

## Passo 8: Implementar Comando Exec

### 8.1 Editar cmd/exec.go
```go
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <nome> <comando>",
	Short: "Executa comando no container",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		command := args[1:]
		
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Erro ao conectar com Docker: %v\n", err)
			return
		}
		defer cli.Close()

		execConfig := container.ExecOptions{
			AttachStdout: true,
			AttachStderr: true,
			AttachStdin:  true,
			Tty:          true,
			Cmd:          command,
		}

		execID, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
		if err != nil {
			fmt.Printf("Erro ao criar execução no container %s: %v\n", containerName, err)
			return
		}

		execAttach, err := cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{
			Tty: true,
		})
		if err != nil {
			fmt.Printf("Erro ao anexar execução: %v\n", err)
			return
		}
		defer execAttach.Close()

		fmt.Printf("=== Executando '%s' no container %s ===\n", strings.Join(command, " "), containerName)
		io.Copy(os.Stdout, execAttach.Reader)
	},
}

func init() {
	containerCmd.AddCommand(execCmd)
}
```

## Passo 9: Implementar Comando Health

### 9.1 Editar cmd/health.go
```go
package cmd

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health <nome>",
	Short: "Verifica saúde do container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Erro ao conectar com Docker: %v\n", err)
			return
		}
		defer cli.Close()

		inspect, err := cli.ContainerInspect(ctx, containerName)
		if err != nil {
			fmt.Printf("Erro ao inspecionar container %s: %v\n", containerName, err)
			return
		}

		fmt.Printf("=== Status de Saúde do Container %s ===\n", containerName)
		fmt.Printf("Estado: %s\n", inspect.State.Status)
		fmt.Printf("Executando: %t\n", inspect.State.Running)
		fmt.Printf("PID: %d\n", inspect.State.Pid)
		fmt.Printf("Código de Saída: %d\n", inspect.State.ExitCode)
		
		if inspect.State.Health != nil {
			fmt.Printf("Health Check: %s\n", inspect.State.Health.Status)
			if len(inspect.State.Health.Log) > 0 {
				lastCheck := inspect.State.Health.Log[len(inspect.State.Health.Log)-1]
				fmt.Printf("Última Verificação: %s\n", lastCheck.Start)
				fmt.Printf("Resultado: %s\n", lastCheck.Output)
			}
		} else {
			fmt.Println("Health Check: Não configurado")
		}

		fmt.Printf("Tempo de Execução: %s\n", inspect.State.StartedAt)
		if inspect.RestartCount > 0 {
			fmt.Printf("Reinicializações: %d\n", inspect.RestartCount)
		}
	},
}

func init() {
	containerCmd.AddCommand(healthCmd)
}
```

## Passo 10: Implementar Comando Deploy

### 10.1 Editar cmd/deploy.go
```go
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
	"github.com/docker/docker/client"
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
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Erro ao conectar com Docker: %v\n", err)
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
```

## Passo 11: Criar Arquivo de Configuração de Exemplo

### 11.1 Criar example-config.json
```json
{
  "name": "meu-nginx",
  "image": "nginx:latest",
  "ports": {
    "80/tcp": "8080"
  },
  "env": [
    "NGINX_HOST=localhost",
    "NGINX_PORT=80"
  ],
  "volumes": [
    "/host/path:/container/path:ro"
  ],
  "restart": "unless-stopped",
  "networks": ["bridge"]
}
```

## Passo 12: Compilar e Testar

### 12.1 Resolver dependências
```bash
go mod tidy
```

### 12.2 Compilar
```bash
go build -o infra-tool
```

### 12.3 Testar
```bash
# Testar ajuda
./infra-tool --help
./infra-tool container --help

# Testar comandos (com Docker rodando)
./infra-tool container list
./infra-tool container deploy example-config.json
```

## Passo 13: Instalação Global (Opcional)

### 13.1 Instalar globalmente
```bash
go install
```

Agora você pode usar `infra-tool` de qualquer lugar no sistema.

## Estrutura Final

```
infra-tool/
├── cmd/
│   ├── root.go           # Comando raiz
│   ├── container.go      # Comando container pai
│   ├── list.go          # Subcomando list
│   ├── logs.go          # Subcomando logs
│   ├── exec.go          # Subcomando exec
│   ├── health.go        # Subcomando health
│   └── deploy.go        # Subcomando deploy
├── main.go              # Ponto de entrada
├── example-config.json  # Exemplo de configuração
├── go.mod               # Módulo Go
├── go.sum               # Checksums das dependências
└── infra-tool           # Binário compilado
```

## Dicas

1. **Use `cobra-cli add` sempre** para gerar novos comandos
2. **Mantenha um arquivo por comando** para organização
3. **Use `go mod tidy`** sempre após adicionar dependências
4. **Teste cada comando** após implementá-lo
5. **Use contextos** para todas as operações Docker
6. **Sempre feche recursos** (defer cli.Close(), reader.Close())

Este guia segue as melhores práticas do Cobra CLI e produz código limpo e organizável!