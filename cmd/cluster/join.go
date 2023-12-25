package cluster

import (
	"fmt"

	"github.com/cnrancher/autok3s/cmd/common"
	"github.com/cnrancher/autok3s/pkg/providers"
	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// JoinCommand join command.
func joinCommand() *cobra.Command {
	joinCmd := &cobra.Command{
		Use:   "join",
		Short: "Join one or more K3s node(s) to an existing cluster",
	}
	joinCmd.Flags().StringVarP(&providerName, "provider", "p", providerName, "Provider is a module which provides an interface for managing cloud resources")
	// load dynamic provider flags.
	pStr := common.FlagHackLookup("--provider")
	if pStr != "" {
		if reg, err := providers.GetProvider(pStr); err != nil {
			logrus.Fatalln(err)
		} else {
			provider = reg
		}

		joinCmd.Flags().AddFlagSet(utils.ConvertFlags(joinCmd, provider.GetCredentialFlags()))
		joinCmd.Flags().AddFlagSet(utils.ConvertFlags(joinCmd, provider.GetJoinFlags()))
		joinCmd.Example = provider.GetUsageExample("join")
		joinCmd.Use = fmt.Sprintf("join -p %s", pStr)
	}

	joinCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if providerName == "" {
			logrus.Fatalln("required flag(s) \"[provider]\" not set")
		}
		common.BindEnvFlags(cmd)
		err := provider.MergeClusterOptions()
		if err != nil {
			return err
		}

		if err = common.MakeSureCredentialFlag(cmd.Flags(), provider); err != nil {
			return err
		}
		utils.ValidateRequiredFlags(cmd.Flags())
		return nil
	}

	joinCmd.Run = func(cmd *cobra.Command, args []string) {
		// generate cluster name. i.e. input: "--name k3s1 --region cn-hangzhou" output: "k3s1.cn-hangzhou".
		provider.GenerateClusterName()
		if err := provider.JoinCheck(); err != nil {
			logrus.Fatalln(err)
		}
		// join k3s node to the cluster which named with generated cluster name.
		if err := provider.JoinK3sNode(); err != nil {
			logrus.Fatalln(err)
		}
	}

	return joinCmd
}
