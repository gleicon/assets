package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
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
		cli, err := NewDockerClient()
		if err != nil {
			fmt.Printf("%v\n", err)
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