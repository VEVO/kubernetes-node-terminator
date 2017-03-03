package main

import (
	"k8s.io/client-go/pkg/api/v1"
)

type kubernetesNode struct {
	name string
}

type kubernetesNodes struct {
	list []kubernetesNode
}

func (k kubernetesNodes) getNodesByLabel(client kubernetesClient, labels map[string]string) (*v1.NodeList, error) {
	listOptions := v1.ListOptions{
		LabelSelector: keysString(labels),
	}
	nodeObject, err := client.getNodes(listOptions)
	return nodeObject, err
}
