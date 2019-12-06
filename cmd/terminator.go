package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/VEVO/kubernetes-node-terminator/pkg/k8snode"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	cloudProvider       = os.Getenv("CLOUD_PROVIDER")
	cloudRegion         = os.Getenv("CLOUD_REGION")
	dryRun              = false
	dryRunStr           = os.Getenv("DRY_RUN")
	terminationDelayStr = os.Getenv("DELAY_BETWEEN_TERMINATIONS")
	terminationDelay    = time.Duration(300)
	intervalStr         = os.Getenv("INTERVAL_SECONDS")
	interval            = time.Duration(60)
	maxUnhealthyStr     = os.Getenv("MAX_UNHEALTHY")
	maxUnhealthy        = 1
	healthPortStr       = os.Getenv("HEALTH_PORT")
	healthPort          = "8080"
)

type instance struct {
	instanceID   string
	terminatedAt time.Time
}

type terminatorState struct {
	terminated []*instance
	heartBeat  time.Time
}

func (t *terminatorState) okToTerminate(instanceID string) bool {
	var alreadyTerminated bool

	glog.V(4).Infof("Checking for already terminated instances")
	for _, e := range t.terminated {
		glog.V(4).Infof("Checking if instance %s was already terminated.", e.instanceID)

		if instanceID == e.instanceID {
			alreadyTerminated = true
			glog.V(4).Infof("Instance %s was already terminated.", e.instanceID)
			break
		}
	}

	return !alreadyTerminated
}

func (t *terminatorState) expireTerminatedInstances() {
	var expirationDuration = 1 * time.Hour
	var now = time.Now()

	currentState := &terminatorState{}

	for _, e := range t.terminated {
		glog.V(4).Infof("Expiration candidate is %v", e)
		timeDiff := now.Sub(e.terminatedAt)

		if timeDiff < expirationDuration {
			glog.V(4).Infof("Candidate %v not expired", e)
			currentState.terminated = append(currentState.terminated, e)
		}

	}
	t.terminated = currentState.terminated
}

func (t *terminatorState) healthServer(healthPort string, interval time.Duration) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// Current timestamp minus 2 intervals to allow a little buffer
		now := time.Now().Add(-(interval * 2) * time.Second)

		if now.After(t.heartBeat) {
			glog.V(4).Infof("Health failing - hearbeat time %s", t.heartBeat)
			w.WriteHeader(500)
			w.Write([]byte("error"))
		} else {
			glog.V(4).Infof("Health passing - hearbeat time %s", t.heartBeat)
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		}
	})

	go http.ListenAndServe(":"+healthPort, nil)
}

func newClient() (*kubernetes.Clientset, error) {
	var clientset *kubernetes.Clientset

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return clientset, err
	}
	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return clientset, err
	}
	return clientset, err
}

func main() {
	var err error
	flag.Parse()

	flag.Lookup("logtostderr").Value.Set("true")

	switch {
	case dryRunStr != "":
		dryRun, err = strconv.ParseBool(dryRunStr)
		if err != nil {
			glog.Fatalf("Error parsing DRY_RUN value: %s", err)
		}
	case cloudProvider == "":
		glog.Fatal("Set the CLOUD_PROVIDER variable")
	case cloudRegion == "":
		glog.Fatal("Set the CLOUD_REGION variable")
	case intervalStr != "":
		t, _ := strconv.Atoi(intervalStr)
		interval = time.Duration(t)
	case healthPortStr != "":
		healthPort = healthPortStr
	case maxUnhealthyStr != "":
		t, _ := strconv.Atoi(maxUnhealthyStr)
		maxUnhealthy = t
	}

	glog.Info("Starting node-terminator")
	client, err := newClient()
	if err != nil {
		glog.Fatal(err.Error())
	}

	state := &terminatorState{
		heartBeat: time.Now()}

	state.healthServer("8080", interval)

	config := k8snode.NewConfig(client, cloudProvider, cloudRegion, dryRun)

	labels := make(map[string]string)
	labels["status"] = "unhealthy"

	var instanceID string
	for {
		glog.Info("Checking for unhealthy instances")

		nodeList, err := config.Status(labels)
		if err != nil {
			glog.Fatalf("failed to populate node by label: %s", err)
		}

		if len(nodeList.Items) == 0 {
			glog.Infof("No unhealthy nodes")

		} else if len(nodeList.Items) <= maxUnhealthy {
			for _, i := range nodeList.Items {
				instanceID = i.Labels["instance-id"]

				if state.okToTerminate(instanceID) {
					err := config.Terminate(i)
					if err != nil {
						glog.Errorf("An error occurred terminating instance %s\n. Error: %s", instanceID, err)
					}
					i := &instance{
						instanceID:   instanceID,
						terminatedAt: time.Now()}

					state.terminated = append(state.terminated, i)
					state.heartBeat = time.Now()
					time.Sleep(time.Second * terminationDelay)
				}
			}

		} else {
			glog.Infof("No action will be taken while the unhealthy node count (%d) is greater than MAX_UNHEALTHY (%d).", len(nodeList.Items), maxUnhealthy)
		}

		state.expireTerminatedInstances()
		state.heartBeat = time.Now()
		time.Sleep(time.Second * interval)
	}
}
