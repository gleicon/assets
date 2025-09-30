package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health <nome>",
	Short: "Verifica saúde do container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		ctx := context.Background()
		cli, err := NewDockerClient()
		if err != nil {
			fmt.Printf("%v\n", err)
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