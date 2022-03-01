package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/forbole/juno/v2/cmd"

	"github.com/spf13/cobra"

	"github.com/desmos-labs/soteria/cmd/export"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Build the root command
	rootCmd := &cobra.Command{
		Use: "soteria",
	}

	rootCmd.AddCommand(
		export.NewCmdExport(),
	)

	// Build and run the executor
	executor := cmd.PrepareRootCmd("soteria", rootCmd)
	err := executor.Execute()
	if err != nil {
		os.Exit(1)
	}
}
