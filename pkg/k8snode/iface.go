package k8snode

import (
	corev1 "k8s.io/api/core/v1"
)

type Node interface {
	Status(map[string]string) (*corev1.NodeList, error)
	Terminate(corev1.Node) error
	Event(corev1.Node) error
}

type Provider interface {
	TerminateInstance(string) error
}
