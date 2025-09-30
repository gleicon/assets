package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var kubeconfig string

// k8sCmd representa o comando principal do k8s
var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Gerenciar recursos do Kubernetes",
	Long:  "Comandos para interagir com clusters Kubernetes locais",
}

// k8sListCmd lista aplicações rodando no cluster
var k8sListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista aplicações (pods) rodando no cluster Kubernetes",
	Long:  "Lista todos os pods em execução no cluster Kubernetes local",
	Run: func(cmd *cobra.Command, args []string) {
		// Conectar ao cluster Kubernetes
		clientset, err := getKubernetesClient()
		if err != nil {
			fmt.Printf("Erro ao conectar com Kubernetes: %v\n", err)
			return
		}

		// Listar pods em todos os namespaces
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Erro ao listar pods: %v\n", err)
			return
		}

		if len(pods.Items) == 0 {
			fmt.Println("Nenhum pod encontrado no cluster")
			return
		}

		// Mostrar cabeçalho
		fmt.Printf("%-30s %-20s %-15s %-10s\n", "NOME", "NAMESPACE", "STATUS", "RESTARTS")
		fmt.Println("--------------------------------------------------------------------------------")

		// Listar cada pod
		for _, pod := range pods.Items {
			restarts := int32(0)
			status := string(pod.Status.Phase)
			
			// Contar total de restarts
			for _, containerStatus := range pod.Status.ContainerStatuses {
				restarts += containerStatus.RestartCount
			}

			// Se o pod não está rodando, mostrar mais detalhes do status
			if pod.Status.Phase != "Running" {
				for _, condition := range pod.Status.Conditions {
					if condition.Type == "Ready" && condition.Status == "False" {
						status = condition.Reason
						break
					}
				}
			}

			fmt.Printf("%-30s %-20s %-15s %-10d\n", 
				pod.Name, 
				pod.Namespace, 
				status, 
				restarts)
		}

		fmt.Printf("\nTotal: %d pods\n", len(pods.Items))
	},
}

// getKubernetesClient cria um cliente para o Kubernetes
func getKubernetesClient() (*kubernetes.Clientset, error) {
	var configPath string
	
	// Se kubeconfig foi especificado via flag, usar ele
	if kubeconfig != "" {
		configPath = kubeconfig
	} else {
		// Tentar encontrar o arquivo kubeconfig padrão
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Verificar se o arquivo existe
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("arquivo kubeconfig não encontrado em %s. Execute 'kubectl config view' para verificar a configuração", configPath)
	}

	// Carregar configuração do kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar kubeconfig: %v", err)
	}

	// Criar cliente do Kubernetes
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente Kubernetes: %v", err)
	}

	return clientset, nil
}

func init() {
	// Adicionar comando k8s ao root
	rootCmd.AddCommand(k8sCmd)
	
	// Adicionar subcomando list ao k8s
	k8sCmd.AddCommand(k8sListCmd)
	
	// Flag para especificar caminho do kubeconfig
	k8sCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Caminho para o arquivo kubeconfig (padrão: ~/.kube/config)")
}