package server

import (
	"net/http"

	"github.com/orangedeng/autok3s/pkg/server/proxy"
	"github.com/orangedeng/autok3s/pkg/server/ui"

	"github.com/gorilla/mux"
	responsewriter "github.com/rancher/apiserver/pkg/middleware"
	"github.com/rancher/apiserver/pkg/server"
	"github.com/rancher/apiserver/pkg/store/apiroot"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/sirupsen/logrus"

	// registers the pprof handlers on http.DefaultServeMux, which is only
	// ever served by the dedicated loopback listener in StartPprof.
	_ "net/http/pprof"
)

// StartPprof serves the standard pprof endpoints on a dedicated listener
// (loopback by default) so profiling never rides the main server, which
// may be bound to a public interface.
func StartPprof(addr string) {
	go func() {
		logrus.Infof("serving pprof on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logrus.Warnf("pprof listener stopped: %v", err)
		}
	}()
}

// Start starts daemon.
func Start() http.Handler {
	s := server.DefaultAPIServer()
	initMutual(s.Schemas)
	initProvider(s.Schemas)
	initCluster(s.Schemas)
	initCredential(s.Schemas)
	initKubeconfig(s.Schemas)
	initLogs(s.Schemas)
	initTemplates(s.Schemas)
	initSettings(s.Schemas)
	initPackage(s.Schemas)
	initSSHKey(s.Schemas)
	initAddon(s.Schemas)

	apiroot.Register(s.Schemas, []string{"v1"})
	router := mux.NewRouter()
	router.UseEncodedPath()
	router.StrictSlash(true)

	middleware := responsewriter.Chain{
		responsewriter.Gzip,
		responsewriter.FrameOptions,
		responsewriter.CacheMiddleware("json", "js", "css", "svg", "png", "woff", "woff2"),
		ui.ServeNotFound,
		ui.ServeJavascript,
	}
	router.PathPrefix("/ui/").Handler(middleware.Handler(http.StripPrefix("/ui/", ui.Serve())))

	router.Path("/").HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, "/ui/", http.StatusFound)
	})

	router.PathPrefix("/meta/proxy").Handler(proxy.NewProxy("/proxy/"))
	router.PathPrefix("/k8s/proxy").Handler(proxy.NewK8sProxy())
	router.Path("/{prefix}/{type}").Queries("action", "{action}").Handler(s)
	router.Path("/{prefix}/{type}").Handler(s)
	router.Path("/{prefix}/{type}/{name}").Queries("link", "{link}").Handler(s)
	router.Path("/{prefix}/{type}/{name}").Queries("action", "{action}").Handler(s)
	router.Path("/{prefix}/{type}/{name}").Handler(s)

	router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Handle(&types.APIRequest{
			Request:   r,
			Response:  rw,
			Type:      "apiRoot",
			URLPrefix: "v1",
		})
	})

	return router
}
