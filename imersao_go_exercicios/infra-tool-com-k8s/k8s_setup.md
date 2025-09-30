# Setup do Kubernetes para Iniciantes

Este guia ajuda você a configurar um cluster Kubernetes local para usar com a nossa ferramenta de infraestrutura.

## Pré-requisitos

- Docker instalado e funcionando
- Go 1.24+ instalado

## Opção 1: Kind (Kubernetes in Docker) - Recomendado

O Kind é a maneira mais simples de rodar Kubernetes localmente usando containers Docker.

### Instalação do Kind

**No macOS:**
```bash
brew install kind
```

**No Linux:**
```bash
# Baixar o binário
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

### Criando um cluster

```bash
# Criar um cluster simples
kind create cluster --name meu-cluster

# Verificar se o cluster foi criado
kubectl cluster-info --context kind-meu-cluster
```

### Testando o cluster

```bash
# Listar nós
kubectl get nodes

# Criar um pod de teste
kubectl run nginx-teste --image=nginx --port=80

# Verificar o pod
kubectl get pods
```

## Opção 2: Minikube

### Instalação do Minikube

**No macOS:**
```bash
brew install minikube
```

**No Linux:**
```bash
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```

### Iniciando o Minikube

```bash
# Iniciar cluster
minikube start

# Verificar status
minikube status
```

## Instalação do kubectl

O kubectl é necessário para gerenciar o cluster Kubernetes.

**No macOS:**
```bash
brew install kubectl
```

**No Linux:**
```bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

## Configuração

Após instalar e criar o cluster, você terá um arquivo de configuração em `~/.kube/config`. Este arquivo é usado automaticamente pela nossa ferramenta.

### Verificando a configuração

```bash
# Ver contextos disponíveis
kubectl config get-contexts

# Ver configuração atual
kubectl config view
```

## Usando nossa ferramenta

Após configurar o Kubernetes, você pode usar nossa ferramenta:

```bash
# Compilar a ferramenta
go build -o infra-tool

# Listar pods no cluster
./infra-tool k8s list

# Usar kubeconfig personalizado (opcional)
./infra-tool k8s list --kubeconfig /caminho/para/config
```

## Aplicações de exemplo para testar

### Deploy de uma aplicação simples

```bash
# Criar um deployment do nginx
kubectl create deployment nginx-app --image=nginx

# Expor o deployment
kubectl expose deployment nginx-app --port=80 --type=NodePort

# Ver os recursos criados
kubectl get all
```

Agora você pode usar `./infra-tool k8s list` para ver os pods da aplicação!

## Limpeza

### Kind
```bash
kind delete cluster --name meu-cluster
```

### Minikube
```bash
minikube delete
```

## Dicas importantes

1. **Contexto**: Certifique-se de estar no contexto correto do cluster
2. **Pods vs Deployments**: Pods são instâncias individuais, Deployments gerenciam múltiplos pods
3. **Namespaces**: Use namespaces para organizar recursos (`kubectl get pods -n nome-namespace`)
4. **Logs**: Para debugar, use `kubectl logs nome-do-pod`

## Troubleshooting

### Erro "kubeconfig não encontrado"
- Verifique se o arquivo `~/.kube/config` existe
- Execute `kubectl config view` para ver a configuração

### Erro "connection refused"
- Verifique se o cluster está rodando (`kubectl cluster-info`)
- Para Kind: `kind get clusters`
- Para Minikube: `minikube status`

### Pods não aparecem
- Verifique se há pods rodando: `kubectl get pods --all-namespaces`
- Crie um pod de teste: `kubectl run teste --image=nginx`

---

**Dica:** Use sempre `kubectl get pods` primeiro para verificar se há aplicações rodando antes de testar nossa ferramenta!