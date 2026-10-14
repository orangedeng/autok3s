package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/sirupsen/logrus"
	k8sproxy "k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sProxyHandler struct{}

// transportCache caches http.RoundTrippers keyed by the sha256 of the
// kubeconfig content. Creating a new transport per request leaks idle
// connections (and therefore local ephemeral ports) because each transport
// keeps its own connection pool which is never closed.
var (
	transportCache = sync.Map{}
	cacheMutex     sync.Mutex
)

func getTransportForConfig(kubeconfig string) (http.RoundTripper, error) {
	sum := sha256.Sum256([]byte(kubeconfig))
	key := hex.EncodeToString(sum[:])

	if cached, ok := transportCache.Load(key); ok {
		return cached.(http.RoundTripper), nil
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if cached, ok := transportCache.Load(key); ok {
		return cached.(http.RoundTripper), nil
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, err
	}
	transport, err := rest.TransportFor(cfg)
	if err != nil {
		return nil, err
	}
	transportCache.Store(key, transport)
	return transport, nil
}

func NewK8sProxy() http.Handler {
	return &K8sProxyHandler{}
}

type errorResponder struct{}

func (r *errorResponder) Error(w http.ResponseWriter, _ *http.Request, err error) {
	logrus.Errorf("Error while proxying request: %v", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (kh *K8sProxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	config := req.Header.Get("config")
	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(utils.StringSupportBase64(config)))

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}
	host := cfg.Host
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}
	u, err := url.Parse(host)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}

	u.Path = strings.TrimPrefix(req.URL.Path, "/k8s/proxy")
	u.RawQuery = req.URL.RawQuery
	req.URL.Host = req.Host
	responder := &errorResponder{}
	transport, err := getTransportForConfig(utils.StringSupportBase64(config))
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}
	handler := k8sproxy.NewUpgradeAwareHandler(u, transport, true, false, responder)
	handler.ServeHTTP(rw, req)
}
