//go:generate go run pkg/settings/script/main.go ./pkg/settings/install.sh
package main

import (
	"os"

	"github.com/orangedeng/autok3s/cmd"
	"github.com/orangedeng/autok3s/cmd/addon"
	"github.com/orangedeng/autok3s/cmd/airgap"
	"github.com/orangedeng/autok3s/cmd/sshkey"
	"github.com/orangedeng/autok3s/pkg/common"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	gitVersion   = "dev"
	gitCommit    string
	gitTreeState string
	buildDate    string
)

func main() {
	rootCmd := cmd.Command()
	rootCmd.AddCommand(cmd.CompletionCommand(), cmd.VersionCommand(gitVersion, gitCommit, gitTreeState, buildDate),
		cmd.ListCommand(), cmd.CreateCommand(), cmd.JoinCommand(), cmd.KubectlCommand(), cmd.DeleteCommand(),
		cmd.SSHCommand(), cmd.DescribeCommand(), cmd.ServeCommand(), cmd.UpgradeCommand(),
		airgap.Command(), sshkey.Command(), addon.Command())

	rootCmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		common.InitLogger(logrus.StandardLogger())
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
