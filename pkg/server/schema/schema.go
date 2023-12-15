package schema

import (
	"net/http"

	"github.com/cnrancher/autok3s/pkg/common"
	"github.com/cnrancher/autok3s/pkg/server/store/addon"
	"github.com/cnrancher/autok3s/pkg/server/store/cluster"
	"github.com/cnrancher/autok3s/pkg/server/store/credential"
	"github.com/cnrancher/autok3s/pkg/server/store/explorer"
	"github.com/cnrancher/autok3s/pkg/server/store/kubectl"
	"github.com/cnrancher/autok3s/pkg/server/store/pkg"
	"github.com/cnrancher/autok3s/pkg/server/store/provider"
	"github.com/cnrancher/autok3s/pkg/server/store/settings"
	"github.com/cnrancher/autok3s/pkg/server/store/sshkey"
	"github.com/cnrancher/autok3s/pkg/server/store/template"
	"github.com/cnrancher/autok3s/pkg/server/store/websocket"
	wkube "github.com/cnrancher/autok3s/pkg/server/store/websocket/kubectl"
	"github.com/cnrancher/autok3s/pkg/server/store/websocket/ssh"
	backendtypes "github.com/cnrancher/autok3s/pkg/types"
	autok3stypes "github.com/cnrancher/autok3s/pkg/types/apis"

	"github.com/rancher/apiserver/pkg/types"
	wranglertypes "github.com/rancher/wrangler/pkg/schemas"
)

var (
	initFuncs = []schemaInit{
		initProvider,
		initCluster,
		initCredential,
		initMutual,
		initKubeconfig,
		initLogs,
		initTemplates,
		initExplorer,
		initSettings,
		initPackage,
		initSSHKey,
		initAddon,
	}
)

func InitSchema(s *types.APISchemas) *types.APISchemas {
	return chain(s, initFuncs...)
}

type schemaInit func(s *types.APISchemas) *types.APISchemas

func chain(result *types.APISchemas, chainFuncs ...schemaInit) *types.APISchemas {
	for _, f := range chainFuncs {
		result = f(result)
	}
	return result
}

func initProvider(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(autok3stypes.Provider{}, func(schema *types.APISchema) {
		schema.Store = &provider.Store{}
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{http.MethodGet}
	})
	return s
}

func initCluster(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(backendtypes.SSH{}, nil)
	s.MustImportAndCustomize(backendtypes.Node{}, nil)
	s.MustImportAndCustomize(autok3stypes.KubeconfigOutput{}, nil)
	s.MustImportAndCustomize(autok3stypes.EnableExplorerOutput{}, nil)
	s.MustImportAndCustomize(autok3stypes.UpgradeInput{}, nil)
	s.MustImportAndCustomize(autok3stypes.Cluster{}, func(schema *types.APISchema) {
		schema.Store = &cluster.Store{}
		// common.DefaultDB.Register()
		schema.CollectionMethods = []string{http.MethodGet, http.MethodPost}
		schema.ResourceMethods = []string{http.MethodGet, http.MethodDelete}
		schema.ResourceActions["join"] = wranglertypes.Action{
			Input: "cluster",
		}
		schema.ResourceActions["enableExplorer"] = wranglertypes.Action{
			Output: "enableExplorerOutput",
		}
		schema.ResourceActions["disableExplorer"] = wranglertypes.Action{}
		schema.ResourceActions["downloadKubeconfig"] = wranglertypes.Action{
			Output: "kubeconfigOutput",
		}
		schema.ResourceActions["upgrade"] = wranglertypes.Action{
			Input: "upgradeInput",
		}
		schema.Formatter = cluster.Formatter
		schema.ActionHandlers = cluster.HandleCluster()
		schema.ByIDHandler = cluster.LinkCluster
	})
	return s
}

func initCredential(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(autok3stypes.Credential{}, func(schema *types.APISchema) {
		schema.Store = &credential.Store{}
		schema.CollectionMethods = []string{http.MethodGet, http.MethodPost}
		schema.ResourceMethods = []string{http.MethodGet, http.MethodPut, http.MethodDelete}
		schema.Formatter = credential.Formatter
	})
	return s
}

func initMutual(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(autok3stypes.Mutual{}, func(schema *types.APISchema) {
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{}
		schema.ListHandler = ssh.Handler
	})
	return s
}

func initKubeconfig(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(autok3stypes.Config{}, func(schema *types.APISchema) {
		schema.Store = &kubectl.Store{}
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{http.MethodGet}
		schema.ByIDHandler = wkube.KubeHandler
	})
	return s
}

func initLogs(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(autok3stypes.Logs{}, func(schema *types.APISchema) {
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{}
		schema.ListHandler = websocket.LogHandler
		schema.CodeName = "log"
		schema.CodeNamePlural = "logs"
		schema.PluralName = "logs"
	})
	return s
}

func initTemplates(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(autok3stypes.ClusterTemplate{}, func(schema *types.APISchema) {
		schema.Store = &template.Store{}
		schema.CollectionMethods = []string{http.MethodGet, http.MethodPost}
		schema.ResourceMethods = []string{http.MethodGet, http.MethodDelete, http.MethodPut}
	})
	return s
}

func initExplorer(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(common.Explorer{}, func(schema *types.APISchema) {
		schema.Store = &explorer.Store{}
		formatter := explorer.NewFormatter()
		schema.Formatter = formatter.Formatter
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{http.MethodGet}
	})
	return s
}

func initSettings(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(common.Setting{}, func(schema *types.APISchema) {
		schema.Store = &settings.Store{}
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{http.MethodPut, http.MethodGet}
	})
	return s
}

func initPackage(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(common.Package{}, func(schema *types.APISchema) {
		schema.Store = &pkg.Store{}
		schema.CollectionMethods = []string{http.MethodGet, http.MethodPost}
		schema.ResourceMethods = []string{http.MethodGet, http.MethodDelete, http.MethodPut}
		schema.CollectionActions["import"] = wranglertypes.Action{
			Output: "package",
		}
		schema.CollectionActions["updateInstallScript"] = wranglertypes.Action{}
		schema.ResourceActions["cancel"] = wranglertypes.Action{}
		schema.ResourceActions["download"] = wranglertypes.Action{}
		schema.Formatter = pkg.Format
		schema.CollectionFormatter = pkg.CollectionFormat
		schema.ActionHandlers = pkg.ActionHandlers
		schema.LinkHandlers = pkg.LinkHandlers
	})
	return s
}

func initSSHKey(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(common.SshKey{}, func(schema *types.APISchema) {
		schema.CollectionMethods = []string{http.MethodGet, http.MethodPost}
		schema.ResourceMethods = []string{http.MethodGet, http.MethodDelete}
		schema.Store = &sshkey.Store{}
		schema.ActionHandlers = sshkey.ActionHandlers()
		schema.Formatter = sshkey.Format
		schema.ResourceActions["export"] = wranglertypes.Action{
			Output: "sshKey",
		}
	})
	return s
}

func initAddon(s *types.APISchemas) *types.APISchemas {
	s.MustImportAndCustomize(common.Addon{}, func(schema *types.APISchema) {
		schema.Store = &addon.Store{}
		schema.CollectionMethods = []string{http.MethodPost, http.MethodGet}
		schema.ResourceMethods = []string{http.MethodPut, http.MethodGet, http.MethodDelete}
	})
	return s
}
