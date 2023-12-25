package functional

import "github.com/spf13/cobra"

var Commands = []*cobra.Command{
	completionCommand(),
	dashboardCommand(),
	explorerCommand(),
	kubectlCommand(),
	serveCommand(),
	SSHCommand(),
	telemetryCommand(),
	versionCommand(),
}
