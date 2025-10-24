package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newDebugConfigCmd creates a new cobra command for the "debug config" command.
func newDebugConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Prints the current configuration",
		Long:  `Prints the current configuration to the console.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			allSettings := viper.AllSettings()
			fmt.Printf("%v\n", allSettings)
			return nil
		},
	}
}
