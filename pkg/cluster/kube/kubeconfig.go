package kube

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/utils"

	"github.com/sirupsen/logrus"
)

// SaveConfig rewrites a kubeconfig downloaded from a K3s master so that it
// points to the public server IP and the cluster's context name, then merges
// it into the autok3s kubeconfig file.
func SaveConfig(cfg, ip, context string) error {
	replacer := strings.NewReplacer(
		"127.0.0.1", ip,
		"localhost", ip,
		"default", context,
	)

	result := replacer.Replace(cfg)

	tempPath := filepath.Join(common.CfgPath, ".kube")
	if err := utils.EnsureFolderExist(tempPath); err != nil {
		return fmt.Errorf("[cluster] generate kubecfg temp folder error, msg: %s", err)
	}

	temp, err := os.CreateTemp(tempPath, common.KubeCfgTempName)
	if err != nil {
		return fmt.Errorf("[cluster] generate kubecfg temp file error, msg: %s", err)
	}
	defer func() {
		_ = temp.Close()
		if err := os.Remove(temp.Name()); err != nil {
			logrus.Errorf("[cluster] remove kubecfg temp file error, msg: %s", err)
		}
	}()

	absPath, _ := filepath.Abs(temp.Name())
	if err = os.WriteFile(absPath, []byte(result), 0600); err != nil {
		return fmt.Errorf("[cluster] write content to kubecfg temp file error, msg: %v", err)
	}

	return common.FileManager.SaveCfg(context, temp.Name())
}
