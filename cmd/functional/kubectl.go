package functional

import (
	"github.com/cnrancher/autok3s/pkg/cli/kubectl"

	"github.com/spf13/cobra"
)

// kubectlCommand kubectl command.
func kubectlCommand() *cobra.Command {
	return kubectl.EmbedCommand()
}
