package ui

import (
	"crypto/tls"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	localUI        = "./static"
	repoUI         = "pkg/server/ui/static"
	embedUI        = "static"
	uiDefaultIndex = "http://localhost:5173"
)

var (
	insecureClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

// Serve serve ui component.
func Serve() http.Handler {
	if uiMode() == "dev" && !localUIAvailable() {
		return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			_ = serveIndex(writer)
		})
	}
	return serveAsset()
}

// ServeNotFound server ui component, not found will be return index.html.
func ServeNotFound(next http.Handler) http.Handler {
	if uiMode() == "dev" && !localUIAvailable() {
		return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			_ = serveIndex(writer)
		})
	}
	return serveAssetNotFound(next)
}

func ServeJavascript(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// This is used to enforce application/javascript MIME on Windows (https://github.com/orangedeng/autok3s/issues/426)
		// refer to: https://github.com/golang/go/issues/32350
		if strings.HasSuffix(r.URL.Path, ".js") {
			rw.Header().Set("Content-Type", "application/javascript")
		}
		next.ServeHTTP(rw, r)
	})
}

// assetsFS returns the filesystem backing the ui assets. Prod builds use
// the embedded assets; dev builds fall back to the local ./static directory
// (populated by building the ui in ui/) when it exists.
func assetsFS() fs.FS {
	if uiMode() == "dev" && localUIAvailable() {
		return os.DirFS(localUIDir())
	}
	sub, err := fs.Sub(assets, embedUI)
	if err != nil {
		return assets
	}
	return sub
}

func serveAsset() http.Handler {
	handler := fsFunc(func(name string) (fs.File, error) {
		assetPath := name
		file, err := assetsFS().Open(assetPath)
		if err != nil {
			return nil, err
		}
		return file, err
	})

	return http.FileServer(http.FS(handler))
}

func serveAssetNotFound(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		p := strings.TrimPrefix(req.URL.Path, "/ui/")
		_, err := assetsFS().Open(p)
		if err != nil && os.IsNotExist(err) {
			f, err := assetsFS().Open("index.html")
			if err != nil {
				http.NotFound(rw, req)
				return
			}
			_, _ = io.Copy(rw, f)
			_ = f.Close()
			return
		}
		next.ServeHTTP(rw, req)
	})
}

func serveIndex(resp io.Writer) error {
	uiIndex := os.Getenv("AUTOK3S_UI_INDEX")
	if uiIndex == "" {
		uiIndex = uiDefaultIndex
	}
	r, err := insecureClient.Get(uiIndex)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	_, err = io.Copy(resp, r.Body)
	return err
}

// localUIAvailable reports whether locally built ui assets exist on disk,
// so that non-prod builds can serve them without a vite dev server. It
// checks both the working directory (./static) and the repository layout
// (pkg/server/ui/static, populated by `make autok3s build` / `npm run
// build` in ui/) to support running from either location.
func localUIAvailable() bool {
	_, err := os.Stat(path.Join(localUI, "index.html"))
	if err == nil {
		return true
	}
	_, err = os.Stat(path.Join(repoUI, "index.html"))
	return err == nil
}

// localUIDir returns the directory containing the locally built ui assets.
func localUIDir() string {
	if _, err := os.Stat(path.Join(localUI, "index.html")); err == nil {
		return localUI
	}
	return repoUI
}

func uiMode() string {
	mode := os.Getenv("AUTOK3S_UI_MODE")
	if mode == "" {
		return DefaultMode
	}
	return mode
}
