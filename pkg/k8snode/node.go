package k8snode

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
)

type Config struct {
	kclient  *kubernetes.Clientset
	provider Provider
	recorder record.EventRecorder
}

func NewConfig(kclient *kubernetes.Clientset, cloudType string, cloudRegion string, dryRun bool) Node {
	cfg := &Config{kclient: kclient,
		recorder: NewEventRecorder(kclient)}

	switch cloudType {
	case "aws":
		cfg.provider = NewAWSEc2Controller(dryRun, cloudRegion)
	default:
		glog.Fatalf("Cloud provider %s not supported\n", cloudType)
	}
	return cfg
}

func NewEventRecorder(kclient *kubernetes.Clientset) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: kclient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		corev1.EventSource{Component: "node-terminator"})

	return recorder
}

func (c Config) Status(labels map[string]string) (*corev1.NodeList, error) {
	nodeList, err := c.kclient.CoreV1().Nodes().List(c.labelsToListOptions(labels))
	return nodeList, err
}

func (c Config) Terminate(node corev1.Node) error {
	instanceID := node.Labels["instance-id"]
	err := c.provider.TerminateInstance(instanceID)
	c.Event(node)
	if err != nil {
		glog.Error(err)
	}
	return err
}

func (c Config) Event(node corev1.Node) error {
	instanceID := node.Labels["instance-id"]
	ref, err := reference.GetReference(scheme.Scheme, &node)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("Terminating unhealthy node with instance-id %s", instanceID)
	c.recorder.Event(ref, corev1.EventTypeWarning, "Unhealthy Node Termination", msg)
	return err
}

func (c Config) labelsToListOptions(labels map[string]string) metav1.ListOptions {
	keys := make([]string, 0, len(labels))
	for k, v := range labels {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}

	return metav1.ListOptions{
		LabelSelector: strings.Join(keys, ","),
	}
}
