// Package kube isolates every interaction with a cluster's API server:
// building a client from the autok3s kubeconfig, probing status and version,
// listing nodes and persisting downloaded kubeconfigs. It depends only on
// client-go and pkg/common, never on providers or the cluster lifecycle.
package kube

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

// Client builds a kube clientset for the given context of the kubeconfig file.
func Client(name, kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := buildConfigFromFlags(name, kubeconfig)
	if err != nil {
		return nil, err
	}
	config.Timeout = 15 * time.Second
	c, err := kubernetes.NewForConfig(config)
	return c, err
}
