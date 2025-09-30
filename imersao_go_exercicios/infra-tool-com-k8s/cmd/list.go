package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista containers em execução",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cli, err := NewDockerClient()
		if err != nil {
			fmt.Printf("%v\n", err)
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