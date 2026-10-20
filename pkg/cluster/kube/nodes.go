package kube

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NodeView is the apiserver side identity of one node: the object name,
// the autok3s identity labels and the internal address. It is the minimum
// needed to tie a Node object back to a persisted cluster state entry.
type NodeView struct {
	Name       string
	Labels     map[string]string
	InternalIP string
}

// ListNodeViews lists the apiserver view of every node of the cluster.
// The client is kubernetes.Interface so tests can pass a fake clientset.
func ListNodeViews(client kubernetes.Interface) ([]NodeView, error) {
	timeout := int64(5 * time.Second)
	nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeout})
	if err != nil {
		return nil, err
	}
	views := make([]NodeView, 0, len(nodeList.Items))
	for _, node := range nodeList.Items {
		view := NodeView{Name: node.Name, Labels: node.Labels}
		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				view.InternalIP = address.Address
				break
			}
		}
		views = append(views, view)
	}
	return views, nil
}
