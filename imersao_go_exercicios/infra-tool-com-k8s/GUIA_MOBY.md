# Guia Passo a Passo - Biblioteca Moby/Docker para Go

## 1. Configuração Inicial

### Importar as bibliotecas necessárias:
```go
import (
    "context"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)
```

### Criar cliente Docker:
```go
ctx := context.Background()
cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
if err != nil {
    // tratar erro
}
defer cli.Close()
```

## 2. Operações com Containers

### Listar Containers:
```go
containers, err := cli.ContainerList(ctx, container.ListOptions{})
// Para incluir containers parados: ListOptions{All: true}

for _, c := range containers {
    fmt.Printf("ID: %s, Nome: %s, Imagem: %s\n", 
        c.ID[:12], c.Names[0], c.Image)
}
```

### Obter Logs de Container:
```go
options := container.LogsOptions{
    ShowStdout: true,
    ShowStderr: true,
    Follow:     true,    // para stream contínuo
    Tail:       "100",   // últimas 100 linhas
}

logs, err := cli.ContainerLogs(ctx, containerID, options)
if err != nil {
    // tratar erro
}
defer logs.Close()

// Copiar logs para stdout
io.Copy(os.Stdout, logs)
```

### Executar Comandos em Container:
```go
// Criar execução
execConfig := types.ExecConfig{
    AttachStdout: true,
    AttachStderr: true,
    AttachStdin:  true,
    Tty:          true,
    Cmd:          []string{"ls", "-la"},
}

execID, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
if err != nil {
    // tratar erro
}

// Anexar à execução
execAttach, err := cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{
    Tty: true,
})
if err != nil {
    // tratar erro
}
defer execAttach.Close()

// Ler resultado
io.Copy(os.Stdout, execAttach.Reader)
```

### Inspecionar Container (Health Check):
```go
inspect, err := cli.ContainerInspect(ctx, containerID)
if err != nil {
    // tratar erro
}

// Acessar informações
fmt.Printf("Estado: %s\n", inspect.State.Status)
fmt.Printf("Executando: %t\n", inspect.State.Running)
fmt.Printf("PID: %d\n", inspect.State.Pid)

// Health check (se disponível)
if inspect.State.Health != nil {
    fmt.Printf("Health Status: %s\n", inspect.State.Health.Status)
}
```

## 3. Criação e Deploy de Containers

### Pull de Imagem:
```go
reader, err := cli.ImagePull(ctx, "nginx:latest", types.ImagePullOptions{})
if err != nil {
    // tratar erro
}
io.Copy(os.Stdout, reader) // mostra progresso
reader.Close()
```

### Criar Container:
```go
// Configuração do container
containerConfig := &container.Config{
    Image: "nginx:latest",
    Env:   []string{"VAR=value"},
    Cmd:   []string{"nginx", "-g", "daemon off;"},
    ExposedPorts: map[string]struct{}{
        "80/tcp": {},
    },
}

// Configuração do host
hostConfig := &container.HostConfig{
    PortBindings: map[string][]types.PortBinding{
        "80/tcp": {{HostPort: "8080"}},
    },
    Binds: []string{"/host/path:/container/path"},
    RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
}

// Criar
resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "meu-container")
if err != nil {
    // tratar erro
}

// Iniciar
err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
```

### Parar e Remover Container:
```go
// Parar
err := cli.ContainerStop(ctx, containerID, container.StopOptions{})

// Remover
err = cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})
```

## 4. Padrões Importantes

### Context:
- Sempre use `context.Background()` ou contexto personalizado
- Permite cancelamento e timeout de operações

### Error Handling:
- Sempre verifique erros das operações Docker
- Erros comuns: container não encontrado, permissões, Docker daemon indisponível

### Resource Management:
- Use `defer cli.Close()` para fechar cliente
- Feche readers/writers quando terminar (logs, exec attach)

### ID vs Nome:
- Use ID do container para operações precisas
- Nomes podem ser mais legíveis mas podem conflitar

## 5. Exemplo Completo

```go
package main

import (
    "context"
    "fmt"
    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types/container"
)

func main() {
    ctx := context.Background()
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        panic(err)
    }
    defer cli.Close()

    // Listar containers
    containers, err := cli.ContainerList(ctx, container.ListOptions{})
    if err != nil {
        panic(err)
    }

    for _, c := range containers {
        fmt.Printf("Container: %s (%s)\n", c.Names[0], c.ID[:12])
    }
}
```

Este guia cobre as operações essenciais da biblioteca Moby/Docker. Para operações mais avançadas, consulte a documentação oficial do Docker SDK for Go.