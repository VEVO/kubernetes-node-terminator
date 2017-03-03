package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/golang/glog"
)

var (
	terminatorLogLevel  = os.Getenv("LOG_LEVEL")
	dryRunStr           = os.Getenv("DRY_RUN")
	dryRun              = false
	datadogSvcAddress   = os.Getenv("DD_AGENT_SERVICE_HOST")
	awsRegion           = os.Getenv("AWS_REGION")
	terminationDelayStr = os.Getenv("DELAY_BETWEEN_TERMINATIONS")
	terminationDelay    = time.Duration(300)
	intervalStr         = os.Getenv("INTERVAL_SECONDS")
	interval            = time.Duration(60)
	maxUnhealthyStr     = os.Getenv("MAX_UNHEALTHY")
	maxUnhealthy        = 1
)

type instance struct {
	instanceID   string
	terminatedAt time.Time
}

type terminatorState struct {
	terminated    []*instance
	k8sClient     kubernetesClient
	awsClient     *awsClient
	datadogClient *datadogClient
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

func (t *terminatorState) terminateInstance(instanceID string) error {

	if t.okToTerminate(instanceID) {
		r, err := t.awsClient.ec2.terminateInstance(instanceID)

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case "DryRunOperation":
					glog.Infof("DryRun: %s", err)
					break
				default:
					return fmt.Errorf("an error occurred while terminating instance %s\n Error: %s\n Response: %s", instanceID, err, r)
				}
			}
		}

		eventText := fmt.Sprintf("Terminating unhealthy instance %s.", instanceID)
		glog.Info(eventText)
		err = t.datadogClient.sendEvent(eventText)

		i := &instance{
			instanceID:   instanceID,
			terminatedAt: time.Now()}

		t.terminated = append(t.terminated, i)
	}

	return nil
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

func main() {
	flag.Parse()

	flag.Lookup("logtostderr").Value.Set("true")

	if terminatorLogLevel != "" {
		flag.Lookup("v").Value.Set(terminatorLogLevel)
	} else {
		flag.Lookup("v").Value.Set("2")
	}

	if dryRunStr != "" {
		dryRun = true
	}

	if awsRegion == "" {
		glog.Fatal("Set the AWS_REGION variable to the name of the desired AWS region")
	}

	if datadogSvcAddress == "" {
		glog.Fatal("Set the DD_AGENT_SERVICE_HOST to the ipaddress of the datadog agent service")
	}

	if terminationDelayStr != "" {
		t, _ := strconv.Atoi(terminationDelayStr)
		terminationDelay = time.Duration(t)
	}

	if intervalStr != "" {
		t, _ := strconv.Atoi(intervalStr)
		interval = time.Duration(t)
	}

	if maxUnhealthyStr != "" {
		t, _ := strconv.Atoi(maxUnhealthyStr)
		maxUnhealthy = t
	}

	glog.Infof("Terminator started. Interval is %d, delay is %d and dry run mode is %t", interval, terminationDelay, dryRun)

	nodesController := kubernetesNodes{}
	labels := make(map[string]string)
	labels["status"] = "unhealthy"

	var instanceID string
	state := &terminatorState{
		awsClient:     newAwsClient(dryRun),
		k8sClient:     newClient(),
		datadogClient: newDatadogClient(datadogSvcAddress)}

	state.datadogClient.title = "Kubernetes Cluster: Node Terminator"
	state.datadogClient.tags = "#kubernetes,docker"

	for {
		glog.Info("Checking for unhealthy instances")

		nodeList, err := nodesController.getNodesByLabel(state.k8sClient, labels)
		if err != nil {
			glog.Fatalf("failed to populate node by label: %s", err)
		}

		if len(nodeList.Items) == 0 {
			glog.Infof("No unhealthy nodes")

		} else if len(nodeList.Items) <= maxUnhealthy {
			for _, i := range nodeList.Items {
				instanceID = i.Labels["instance-id"]
				err = state.terminateInstance(instanceID)
				if err != nil {
					glog.Errorf("An error occurred terminating instance %s\n. Error: %s", instanceID, err)
				}
				time.Sleep(time.Second * terminationDelay)
			}

		} else {
			glog.Infof("No action will be taken while the unhealthy node count (%d) is greater than MAX_UNHEALTHY (%d).", len(nodeList.Items), maxUnhealthy)
		}

		state.expireTerminatedInstances()
		time.Sleep(time.Second * interval)
	}
}
