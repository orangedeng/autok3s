package functional

import (
	"fmt"

	"github.com/cnrancher/autok3s/pkg/version"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "Display autok3s version",
		Example: `  autok3s version`,
	}

	short = false
)

func init() {
	versionCmd.Flags().BoolVarP(&short, "short", "s", short, "Print just the version number")
}

// versionCommand returns version information.
func versionCommand() *cobra.Command {
	version := version.GetInfo()

	versionCmd.Run = func(cmd *cobra.Command, args []string) {
		if short {
			fmt.Printf("Version: %s\n", version.Short())
		} else {
			fmt.Printf("Version: %s\n", version.String())
		}
	}

	return versionCmd
}
