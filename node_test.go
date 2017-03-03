package main

import "testing"

func TestKubernetesNodes_GetNodesByLabel(t *testing.T) {
	client := newFakeClient()
	nodesController := kubernetesNodes{}
	labels := make(map[string]string)
	labels["status"] = "unhealthy"

	nodeList, err := nodesController.getNodesByLabel(client, labels)
	if err != nil {
		t.Errorf("failed to populate node by label: %s", err)
	}

	if len(nodeList.Items) <= 0 {
		t.Error("failed to get node by label")
	}

	for _, node := range nodeList.Items {
		if node.Labels["status"] != "unhealthy" {
			t.Errorf("expected unhealthy but got %s", node.Labels["status"])
		}
	}
}
