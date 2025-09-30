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