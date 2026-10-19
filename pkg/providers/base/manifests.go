package base

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"encoding/base64"
	"path/filepath"
	"text/template"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/hosts/dialer"
)

// ReleaseManifests release manifests.
func (p *ProviderBase) ReleaseManifests() error {
	// remove ui manifest to release ELB.
	masterIP := p.IP
	for _, n := range p.Status.MasterNodes {
		if n.InternalIPAddress[0] == masterIP {
			dialer, err := dialer.NewSSHDialer(&n, true, p.Logger)
			if err != nil {
				return err
			}
			defer dialer.Close()
			_, _ = dialer.ExecuteCommands(
				fmt.Sprintf("kubectl delete -f %s/ui.yaml", common.K3sManifestsDir),
				fmt.Sprintf("rm %s/ui.yaml", common.K3sManifestsDir),
			)

			break
		}
	}
	return nil
}

// GetCustomManifests get custom manifests.
func (p *ProviderBase) GetCustomManifests() ([]string, error) {
	// check is folder or file.
	info, err := os.Stat(p.Manifests)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		cmd, err := prepareManifestFile(p.Manifests, info.Name())
		return []string{cmd}, err
	}
	// upload all files under directory, not include recursive folders.
	deployCmd := make([]string, 0)
	files, err := os.ReadDir(p.Manifests)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		cmd, err := prepareManifestFile(filepath.Join(p.Manifests, f.Name()), f.Name())
		if err != nil {
			return nil, err
		}
		deployCmd = append(deployCmd, cmd)
	}
	return deployCmd, nil
}

func prepareManifestFile(path, name string) (string, error) {
	manifestContent, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(uploadManifestCmd, base64.StdEncoding.EncodeToString(manifestContent), common.K3sManifestsDir, name), nil
}

func (p *ProviderBase) parseDefaultTemplates() template.FuncMap {
	templateValues := p.MasterNodes
	return template.FuncMap{
		"providerTemplate": func(metaName string) string {
			node := templateValues[0]
			nodeType := reflect.TypeOf(node)
			fieldValue := ""
			found := true
			for i := 0; i < nodeType.NumField(); i++ {
				field := nodeType.Field(i)
				if v, ok := field.Tag.Lookup("json"); ok {
					fieldName := strings.Split(v, ",")[0]
					if strings.EqualFold(fieldName, metaName) {
						found = true
						if field.Type.Kind() == reflect.Slice {
							value, ok := reflect.ValueOf(node).Field(i).Interface().([]string)
							if ok {
								fieldValue = value[0]
							}
						} else {
							fieldValue = reflect.ValueOf(node).Field(i).String()
						}
						break
					}

				}
			}
			if !found {
				p.Logger.Warnf("[%s] there's no metaName %s defined", p.Provider, metaName)
			}
			return fieldValue
		},
	}
}

func (p *ProviderBase) addonInstallation(plugin string) (string, error) {
	// check addon plugin
	addon, err := common.DefaultDB.GetAddon(plugin)
	if err != nil {
		p.Logger.Errorf("[%s] failed to get addon by name %s, got error: %v", p.Provider, plugin, err)
		return "", err
	}
	manifest := addon.Manifest
	defaultValues := addon.Values
	// check --set values
	setValues := map[string]string{}
	for key, value := range p.Values {
		if strings.HasPrefix(key, fmt.Sprintf("%s.", plugin)) {
			setValues[strings.TrimPrefix(key, fmt.Sprintf("%s.", plugin))] = value
		} else {
			setValues[key] = value
		}
	}
	values, err := common.GenerateValues(setValues, defaultValues)
	if err != nil {
		p.Logger.Errorf("[%s] failed to generate values for addon %s with values %v: %v", p.Provider, plugin, p.Values, err)
		return "", err
	}
	p.Logger.Debugf("assemble manifest with value %++v", values)
	assembleManifest, err := common.AssembleManifest(values, string(manifest), p.parseDefaultTemplates())
	if err != nil {
		p.Logger.Errorf("[%s] failed to assemble manifest for addon %s with values %v: %v", p.Provider, plugin, setValues, err)
		return "", err
	}
	return fmt.Sprintf(deployPluginCmd,
		base64.StdEncoding.EncodeToString(assembleManifest), common.K3sManifestsDir, plugin), nil
}
