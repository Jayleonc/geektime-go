package cmd

import (
	"fmt"
	"github.com/jayleonc/geektime-go/webook/cmd/command"
	"github.com/spf13/cobra"
	"os"
)

const (
	cliName = "webook"
)

var (
	rootCmd = &cobra.Command{
		Use: cliName,
	}
)

func init() {
	rootCmd.AddCommand(command.NewWebookCommand())
}

func start() error {
	return rootCmd.Execute()
}

func MustStart() {
	if err := start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
