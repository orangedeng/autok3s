package cluster

import (
	"fmt"

	"github.com/cnrancher/autok3s/cmd/common"
	"github.com/cnrancher/autok3s/pkg/providers"
	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CreateCommand create command.
func createCommand() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a K3s cluster",
	}
	createCmd.Flags().StringVarP(&providerName, "provider", "p", providerName, "Provider is a module which provides an interface for managing cloud resources")
	// load dynamic provider flags.
	pStr := common.FlagHackLookup("--provider")
	if pStr != "" {
		if reg, err := providers.GetProvider(pStr); err != nil {
			logrus.Fatalln(err)
		} else {
			provider = reg
		}

		createCmd.Flags().AddFlagSet(utils.ConvertFlags(createCmd, provider.GetCredentialFlags()))
		createCmd.Flags().AddFlagSet(utils.ConvertFlags(createCmd, provider.GetOptionFlags()))
		createCmd.Flags().AddFlagSet(utils.ConvertFlags(createCmd, provider.GetCreateFlags()))
		createCmd.Example = provider.GetUsageExample("create")
		createCmd.Use = fmt.Sprintf("create -p %s", pStr)
	}

	createCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if providerName == "" {
			logrus.Fatalln("required flag(s) \"--provider\" not set")
		}
		common.BindEnvFlags(cmd)
		if err := common.MakeSureCredentialFlag(cmd.Flags(), provider); err != nil {
			return err
		}
		utils.ValidateRequiredFlags(cmd.Flags())
		return nil
	}

	createCmd.Run = func(cmd *cobra.Command, args []string) {
		// generate cluster name. i.e. input: "--name k3s1 --region cn-hangzhou" output: "k3s1.cn-hangzhou.<provider>".
		provider.GenerateClusterName()
		if err := provider.BindCredential(); err != nil {
			logrus.Fatalln(err)
		}
		if err := provider.CreateCheck(); err != nil {
			logrus.Fatalln(err)
		}

		// create k3s cluster with generated cluster name.
		if err := provider.CreateK3sCluster(); err != nil {
			logrus.Fatalln(err)
		}
	}

	return createCmd
}
