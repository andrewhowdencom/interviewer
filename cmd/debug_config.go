package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// debugConfigCmd represents the debug config command
var debugConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show the configuration.",
	Long:  `Show the configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		allSettings := viper.AllSettings()
		b, err := json.MarshalIndent(allSettings, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	debugCmd.AddCommand(debugConfigCmd)
}
