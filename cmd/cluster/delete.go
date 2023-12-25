package cluster

import (
	"fmt"

	"github.com/cnrancher/autok3s/cmd/common"
	"github.com/cnrancher/autok3s/pkg/providers"
	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// DeleteCommand delete command.
func deleteCommand() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a K3s cluster",
	}
	deleteCmd.Flags().StringVarP(&providerName, "provider", "p", providerName, "Provider is a module which provides an interface for managing cloud resources")
	deleteCmd.Flags().BoolVarP(&force, "force", "f", force, "Force delete cluster")
	pStr := common.FlagHackLookup("--provider")

	if pStr != "" {
		if reg, err := providers.GetProvider(pStr); err != nil {
			logrus.Fatalln(err)
		} else {
			provider = reg
		}

		deleteCmd.Flags().AddFlagSet(utils.ConvertFlags(deleteCmd, provider.GetCredentialFlags()))
		deleteCmd.Flags().AddFlagSet(utils.ConvertFlags(deleteCmd, provider.GetDeleteFlags()))
		deleteCmd.Example = provider.GetUsageExample("delete")
		deleteCmd.Use = fmt.Sprintf("delete -p %s", pStr)
	}

	deleteCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
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

	deleteCmd.Run = func(cmd *cobra.Command, args []string) {
		provider.GenerateClusterName()
		if err := provider.DeleteK3sCluster(force); err != nil {
			logrus.Fatalln(err)
		}
	}

	return deleteCmd
}
