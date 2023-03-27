package cmds

import (
	"fmt"
	"github.com/COSAE-FR/ripradius/pkg/utils"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of ripradius",
	Run: func(cmd *cobra.Command, args []string) {
		if asJSON {
			printJSON(map[string]string{
				"name":    utils.Name,
				"version": utils.Version,
			})
		} else {
			fmt.Printf("%s %s\n", utils.Name, utils.Version)
		}
	},
}
