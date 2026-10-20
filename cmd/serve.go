package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/server"

	"github.com/pkg/browser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Run as daemon and serve HTTP/HTTPS request",
	}

	bindPort     = "8080"
	bindAddress  = "127.0.0.1"
	pprofAddress = "127.0.0.1:6060"
)

func init() {
	serveCmd.Flags().StringVar(&bindPort, "bind-port", bindPort, "HTTP/HTTPS bind port")
	serveCmd.Flags().StringVar(&bindAddress, "bind-address", bindAddress, "HTTP/HTTPS bind address")
	serveCmd.Flags().StringVar(&pprofAddress, "pprof-address", pprofAddress,
		"dedicated address for the pprof endpoints, loopback by default, empty to disable")
}

// ServeCommand serve command.
func ServeCommand() *cobra.Command {
	serveCmd.Run = func(_ *cobra.Command, _ []string) {
		common.IsCLI = false
		router := server.Start()
		addr := fmt.Sprintf("%s:%s", bindAddress, bindPort)

		if !isLoopbackHost(bindAddress) {
			logrus.Warnf(
				"binding to %s exposes the unauthenticated API and the stored cloud credentials to that interface",
				bindAddress)
		}

		if pprofAddress != "" {
			server.StartPprof(pprofAddress)
		}

		srv := &http.Server{Addr: addr, Handler: router}

		// SIGINT/SIGTERM trigger a graceful shutdown so in-flight sqlite
		// writes are never cut in half.
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		serverErr := make(chan error, 1)
		go func() {
			logrus.Infof("run as daemon, listening on %s:%s", bindAddress, bindPort)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				serverErr <- err
				return
			}
			serverErr <- nil
		}()

		if err := browser.OpenURL("http://" + addr); err != nil {
			logrus.Warnf("failed to open browser to addr %s", addr)
		}

		select {
		case <-ctx.Done():
			logrus.Info("shutting down autok3s server")
		case err := <-serverErr:
			if err != nil {
				logrus.Errorf("server exited: %v", err)
			}
		}

		// terminal websockets are long-lived, so bound the drain window and
		// force-close whatever is still open when it expires.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logrus.Warnf("graceful shutdown timed out, dropping remaining connections: %v", err)
		}

		// close the sqlite handle so the on-disk state is consistent for
		// the next CLI or server run.
		if common.DefaultDB != nil {
			if sqlDB, err := common.DefaultDB.DB.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
	}

	return serveCmd
}

// isLoopbackHost reports whether a bind address only accepts local
// connections. An empty host means every interface, so it is not loopback.
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
