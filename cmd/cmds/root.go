package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/COSAE-FR/ripradius/pkg/utils"
	"github.com/COSAE-FR/ripradius/svc/daemon"
	"github.com/COSAE-FR/riputils/svc"
	"github.com/spf13/cobra"
	"os"
)

var (
	// Used for flags.
	cfgFile string
	asJSON  bool

	rootCmd = &cobra.Command{
		Use:   "radiusctl",
		Short: fmt.Sprintf("Control the %s daemon", utils.Name),
		Long:  fmt.Sprintf("Control and show stats from the %s daemon.", utils.Name),
	}
)

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	defaultConfigFile := svc.OSDefaultConfigurationFile(utils.Name)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultConfigFile, "config file")
	rootCmd.PersistentFlags().BoolVarP(&asJSON, "json", "j", false, "output JSON")
}

func printJSON(data interface{}) {
	output, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		output, err = json.MarshalIndent(map[string]interface{}{"error": err}, "", "  ")
		if err != nil {
			fmt.Printf("Error outputting JSON: %s\n", err)
		}
	}
	fmt.Printf("%s\n", output)
}

func printError(err error) {
	if asJSON {
		printJSON(map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		fmt.Printf("Error: %s\n", err)
	}
}

func getDaemonConfig() (*daemon.Configuration, error) {
	return daemon.NewConfiguration(cfgFile)
}
