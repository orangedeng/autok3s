package kube

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/orangedeng/autok3s/pkg/types"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Status reports whether the cluster answers its /readyz endpoint.
func Status(c *kubernetes.Clientset) string {
	_, err := c.RESTClient().Get().Timeout(15 * time.Second).RequestURI("/readyz").DoRaw(context.TODO())
	if err != nil {
		return types.ClusterStatusStopped
	}
	return types.ClusterStatusRunning
}

// Version returns the git version reported by the cluster's discovery API.
func Version(c *kubernetes.Clientset) string {
	v, err := c.DiscoveryClient.ServerVersion()
	if err != nil {
		return types.ClusterStatusUnknown
	}
	return v.GitVersion
}

// DescribeNodes enriches the provider instance list with the Kubernetes view
// of each node: hostname, kubelet and container runtime versions, roles and
// readiness. An instance is matched by internal IP, external IP or instance
// ID equal to the node name; instances without a match are left untouched.
// The client is kubernetes.Interface so tests can pass a fake clientset.
func DescribeNodes(client kubernetes.Interface, instanceNodes []types.ClusterNode) ([]types.ClusterNode, error) {
	// list cluster nodes.
	timeout := int64(5 * time.Second)
	nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeout})
	if err != nil || nodeList == nil {
		return nil, err
	}
	for _, node := range nodeList.Items {
		var internalIP, externalIP, hostName string
		addressList := node.Status.Addresses
		for _, address := range addressList {
			switch address.Type {
			case v1.NodeInternalIP:
				internalIP = address.Address
			case v1.NodeExternalIP:
				externalIP = address.Address
			case v1.NodeHostName:
				hostName = address.Address
			default:
				continue
			}
		}
		for index, n := range instanceNodes {
			isCurrentInstance := false
			for _, address := range n.InternalIP {
				if address == internalIP {
					isCurrentInstance = true
					break
				}
			}
			if !isCurrentInstance {
				for _, address := range n.ExternalIP {
					if address == externalIP {
						isCurrentInstance = true
						break
					}
				}
			}
			if !isCurrentInstance {
				if n.InstanceID == node.Name {
					isCurrentInstance = true
				}
			}
			if isCurrentInstance {
				n.HostName = hostName
				n.Version = node.Status.NodeInfo.KubeletVersion
				n.ContainerRuntimeVersion = node.Status.NodeInfo.ContainerRuntimeVersion
				// get roles.
				labels := node.Labels
				roles := make([]string, 0)
				for role := range labels {
					if strings.HasPrefix(role, "node-role.kubernetes.io") {
						roleArray := strings.Split(role, "/")
						if len(roleArray) > 1 {
							roles = append(roles, roleArray[1])
						}
					}
				}
				if len(roles) == 0 {
					roles = append(roles, "<none>")
				}
				sort.Strings(roles)
				n.Roles = strings.Join(roles, ",")
				// get status.
				conditions := node.Status.Conditions
				for _, c := range conditions {
					if c.Type == v1.NodeReady {
						if c.Status == v1.ConditionTrue {
							n.Status = "Ready"
						} else {
							n.Status = "NotReady"
						}
						break
					}
				}
				instanceNodes[index] = n
				break
			}
		}
	}
	return instanceNodes, nil
}
