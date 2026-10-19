package cmd

import (
	"path/filepath"
	"strconv"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/settings"
	"github.com/orangedeng/autok3s/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"
)

var (
	dashboardCmd = &cobra.Command{
		Use:     "helm-dashboard",
		Short:   "Enable helm-dashboard for K3s cluster",
		Example: "autok3s helm-dashboard",
	}

	dashboardPort = 0
)

func init() {
	dashboardCmd.Flags().IntVarP(&dashboardPort, "port", "", dashboardPort, "Set http port for helm-dashboard")
}

// DashboardCommand will start a helm-dashboard server for specified K3s cluster
func DashboardCommand() *cobra.Command {
	dashboardCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		cfg, err := clientcmd.LoadFromFile(filepath.Join(common.CfgPath, common.KubeCfgFile))
		if err != nil {
			return err
		}
		if len(cfg.Contexts) == 0 {
			logrus.Fatalln("cannot enable helm dashboard without K3s cluster")
		}
		return nil
	}
	dashboardCmd.Run = func(_ *cobra.Command, _ []string) {
		if err := common.CheckCommandExist(common.HelmDashboardCommand); err != nil {
			logrus.Fatalln(err)
		}
		if dashboardPort == 0 {
			port, err := utils.GetFreePort()
			if err != nil {
				logrus.Fatalf("failed to get free port for helm-dashboard: %v", err)
			}
			dashboardPort = port
		}
		if err := settings.HelmDashboardPort.Set(strconv.Itoa(dashboardPort)); err != nil {
			logrus.Fatalln(err)
		}
		err := common.StartHelmDashboard(dashboardCmd.Context(), strconv.Itoa(dashboardPort))
		if err != nil {
			logrus.Fatalln(err)
		}
		logrus.Infof("Helm dashboard started with 127.0.0.1:%d", dashboardPort)
	}

	return dashboardCmd
}
