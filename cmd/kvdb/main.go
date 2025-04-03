package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/neekrasov/kvdb/internal/application"
	"github.com/neekrasov/kvdb/internal/config"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitHash   = "unset"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kvdb",
		Short: "Key-value database server",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kvdb version %s\nbuild time: %s\nhash: %s\n",
				version, buildTime, gitHash)
		},
	})

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the database server",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			startServer(configPath)
		},
	}

	runCmd.Flags().StringP("config", "c", "config.yml", "Path to config file")
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func startServer(cfgPath string) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()

	cfg, err := config.GetConfig(cfgPath)
	if err != nil {
		log.Fatalf("failed to get config: %s", err)
	}

	if err := application.New(&cfg).Start(ctx); err != nil {
		log.Fatalf("application error: %s", err)
	}
}
