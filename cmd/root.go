package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/adrg/xdg"
	"github.com/andrewhowdencom/vox/internal/config"
	"github.com/andrewhowdencom/vox/internal/debug"
	"github.com/andrewhowdencom/vox/internal/ports/cli"
	"github.com/andrewhowdencom/vox/internal/ports/web"
	"github.com/andrewhowdencom/vox/internal/telemetry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var logLevel string
var telemetryShutdown func()

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vox",
		Short: "A tool for product managers to understand customer needs.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
			debug.LogDNSInfo()

			// Unmarshal the config into our struct
			var cfg config.Config
			if err := viper.Unmarshal(&cfg); err != nil {
				slog.Error("failed to unmarshal config", slog.Any("error", err))
				os.Exit(1)
			}

			// Initialise telemetry
			var err error
			telemetryShutdown, err = telemetry.Init(&cfg)
			if err != nil {
				slog.Error("failed to initialise telemetry", slog.Any("error", err))
				os.Exit(1)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if telemetryShutdown != nil {
				telemetryShutdown()
			}
		},
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/.vox.yaml)")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	// Add subcommands
	cmd.AddCommand(NewInterviewCmd())
	cmd.AddCommand(web.NewServeCmd())
	cmd.AddCommand(cli.NewDebugCmd())

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set up logging
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in XDG config directory with name ".vox" (without extension).
		viper.AddConfigPath("/etc/vox/")
		viper.AddConfigPath("/etc/")
		viper.AddConfigPath(xdg.ConfigHome)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".vox")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		slog.Info("Using config file", "path", viper.ConfigFileUsed())
	}
}
