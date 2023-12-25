//go:generate go run pkg/settings/script/main.go ./pkg/settings/install.sh
package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/cnrancher/autok3s/cmd"
	"github.com/cnrancher/autok3s/cmd/functional"
	"github.com/cnrancher/autok3s/pkg/cli/kubectl"
	"github.com/cnrancher/autok3s/pkg/common"
	"github.com/cnrancher/autok3s/pkg/metrics"
	"github.com/cnrancher/autok3s/pkg/version"

	"github.com/docker/docker/pkg/reexec"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	reexec.Register("kubectl", kubectl.Main)
}

func main() {

	args := os.Args[0]
	os.Args[0] = filepath.Base(os.Args[0])
	if reexec.Init() {
		return
	}
	os.Args[0] = args

	rootCmd := cmd.Command()

	rootCmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		common.InitLogger(logrus.StandardLogger())
		common.MetricsPrompt(c)
		common.SetupPrometheusMetrics(version.GitVersion())
		go metrics.Report()
		if functional.IsServe(c) {
			metrics.ReportEach(c.Context(), 1*time.Hour)
		}
	}
	rootCmd.PersistentPostRun = func(c *cobra.Command, args []string) {
		metrics.Report()
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
