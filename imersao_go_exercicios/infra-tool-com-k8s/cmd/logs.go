package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <nome>",
	Short: "Exibe logs do container",
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