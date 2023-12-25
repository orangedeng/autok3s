package cluster

import (
	"github.com/cnrancher/autok3s/pkg/providers"
	"github.com/spf13/cobra"
)

var (
	clusterCMD = &cobra.Command{
		Use:   "cluster",
		Short: "The cluster management.",
		Long:  "The cluster command manages the clusters.",
	}
	providerName  string
	provider      providers.Provider
	force         = false
	clusterName   string
	channel       = ""
	version       = ""
	installScript = ""
	uPackageName  = ""
	uPackagePath  = ""
	jsonOut       = false
)

func Command() *cobra.Command {
	clusterCMD.AddCommand(
		listCommand(),
		createCommand(),
		deleteCommand(),
		describeCommand(),
		joinCommand(),
		upgradeCommand(),
	)
	return clusterCMD
}
