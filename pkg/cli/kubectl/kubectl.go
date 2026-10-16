package kubectl

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/hosts/dialer"
	autok3stypes "github.com/orangedeng/autok3s/pkg/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	xterm "golang.org/x/term"
	"k8s.io/client-go/tools/clientcmd"
)

func Main() {
	if runtime.GOOS == "windows" {
		os.Args = os.Args[1:]
	}
	kubenv := os.Getenv("KUBECONFIG")
	for i, arg := range os.Args {
		if strings.HasPrefix(arg, "--kubeconfig=") {
			kubenv = strings.Split(arg, "=")[1]
		} else if strings.HasPrefix(arg, "--kubeconfig") && i+1 < len(os.Args) {
			kubenv = os.Args[i+1]
		}
	}
	if kubenv == "" {
		// get default kubeconfig
		config := filepath.Join(common.CfgPath, common.KubeCfgFile)
		if _, serr := os.Stat(config); serr == nil {
			os.Setenv("KUBECONFIG", config)
		}
		if err := checkReadConfigPermissions(config); err != nil {
			logrus.Warn(err)
		}
	}

	if err := EmbedCommand().Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// EmbedCommand returns the `autok3s kubectl` command. It prefers a kubectl
// binary found in PATH, so the kubeconfig managed by autok3s keeps working
// as usual. When no local kubectl exists, it falls back to executing
// `k3s kubectl` on the first master node of the target cluster over SSH,
// since every node of a K3s cluster ships the kubectl subcommand.
func EmbedCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "kubectl",
		Short:              "Kubectl control interface",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(_ *cobra.Command, args []string) error {
			return runKubectl(args)
		},
	}
}

func runKubectl(args []string) error {
	if bin, err := exec.LookPath("kubectl"); err == nil {
		c := exec.Command(bin, args...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}
	return runK3sKubectlOverSSH(args)
}

// runK3sKubectlOverSSH executes `k3s kubectl` on the first master node of
// the cluster matching the requested context.
func runK3sKubectlOverSSH(args []string) error {
	contextName, args := splitContextFlag(args)
	if contextName == "" {
		var err error
		contextName, err = resolveContext()
		if err != nil {
			return err
		}
	}

	state, err := common.DefaultDB.GetClusterByID(contextName)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("kubectl not found in PATH and context %q is not a cluster managed by autok3s, "+
			"install kubectl to use it with a local kubeconfig", contextName)
	}
	if state.Provider == "k3d" {
		return fmt.Errorf("kubectl not found in PATH, the k3d provider requires a local kubectl binary, " +
			"please install kubectl first")
	}

	masters := make([]autok3stypes.Node, 0)
	if err := json.Unmarshal(state.MasterNodes, &masters); err != nil {
		return err
	}
	if len(masters) == 0 {
		return fmt.Errorf("cluster %q has no master node to run k3s kubectl on", contextName)
	}

	command := strings.Join(append([]string{"k3s", "kubectl"}, args...), " ")
	if masters[0].SSHUser != "root" {
		command = fmt.Sprintf("sudo %s", command)
	}

	d, err := dialer.NewSSHDialer(&masters[0], true, common.NewLogger(nil))
	if err != nil {
		return err
	}
	defer d.Close()

	return runSSHCommand(d.GetClient(), command)
}

// runSSHCommand runs the command over the SSH client, requesting a PTY when
// local stdin is a terminal so interactive subcommands keep working.
func runSSHCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if xterm.IsTerminal(int(os.Stdin.Fd())) {
		term := os.Getenv("TERM")
		if term == "" {
			term = "xterm"
		}
		width, height := 80, 24
		if w, h, err := xterm.GetSize(int(os.Stdin.Fd())); err == nil {
			width, height = w, h
		}
		if err := session.RequestPty(term, height, width, ssh.TerminalModes{ssh.ECHO: 1}); err != nil {
			return err
		}
	}

	return session.Run(command)
}

// splitContextFlag extracts `--context` from the args: the context is used to
// pick the target cluster locally, because the kubeconfig on the master node
// only holds the context of its own cluster.
func splitContextFlag(args []string) (string, []string) {
	remaining := make([]string, 0, len(args))
	contextName := ""
	for i := 0; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "--context="):
			contextName = strings.TrimPrefix(args[i], "--context=")
		case args[i] == "--context" && i+1 < len(args):
			contextName = args[i+1]
			i++
		default:
			remaining = append(remaining, args[i])
		}
	}
	return contextName, remaining
}

// resolveContext returns the context to operate on: the current context of
// the kubeconfig managed by autok3s, or the only managed cluster when there
// is exactly one.
func resolveContext() (string, error) {
	cfgFile := filepath.Join(common.CfgPath, common.KubeCfgFile)
	if cfg, err := clientcmd.LoadFromFile(cfgFile); err == nil && cfg.CurrentContext != "" {
		return cfg.CurrentContext, nil
	}

	clusters, err := common.DefaultDB.ListCluster("")
	if err != nil {
		return "", err
	}
	if len(clusters) == 1 {
		return clusters[0].ContextName, nil
	}
	return "", fmt.Errorf("kubectl not found in PATH and no context could be resolved automatically, " +
		"specify --context <name> or install kubectl")
}

func checkReadConfigPermissions(configFile string) error {
	file, err := os.OpenFile(configFile, os.O_RDONLY, 0600)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("Unable to read %s, please start server "+
				"with --write-kubeconfig-mode to modify kube config permissions", configFile)
		}
	}
	file.Close()
	return nil
}
