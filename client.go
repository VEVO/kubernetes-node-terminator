package main

import (
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

type kubernetesClient interface {
	getNodes(v1.ListOptions) (*v1.NodeList, error)
}

type kubernetesClientConfig struct {
	clientset *kubernetes.Clientset
}

func newClient() kubernetesClient {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &kubernetesClientConfig{clientset: clientset}
}

func (c kubernetesClientConfig) getNodes(listOptions v1.ListOptions) (*v1.NodeList, error) {
	nodeList, err := c.clientset.Core().Nodes().List(listOptions)
	return nodeList, err
}
