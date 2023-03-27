package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/COSAE-FR/ripradius/pkg/api/binding"
	"github.com/COSAE-FR/ripradius/pkg/utils"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the status of the running ripradius",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := getDaemonConfig()
		if err != nil {
			printError(err)
			return
		}
		client := resty.New()
		client.SetBaseURL(fmt.Sprintf("http://%s:%d", cfg.Api.IPAddress, cfg.Api.Port))
		resp, err := client.R().SetHeader("Accept", "application/json").Get("/api/v1/status")
		if err != nil {
			printError(err)
			return
		}
		statusCode := resp.StatusCode()
		switch statusCode {
		case 200:
			status := binding.ServerStatus{}
			if err := json.Unmarshal(resp.Body(), &status); err != nil {
				printError(err)
				return
			}
			if asJSON {
				printJSON(status)
			} else {
				printStatus(status)
			}
		default:
			printError(fmt.Errorf("cannot get server status: %s (%d)", resp.Status(), statusCode))
		}

	},
}

func printStatus(status binding.ServerStatus) {
	fmt.Printf("# %s statistics\n\n## Cache\n\n   - Misses: %d\n   - Hits: %d\n   - Added: %d\n   - Evicted: %d\n   - Entries: %d\n   - Offline: %v\n",
		utils.Name, status.Cache.Misses, status.Cache.Hits, status.Cache.Added, status.Cache.Evicted, status.Cache.Entries, status.Cache.Offline)
}
