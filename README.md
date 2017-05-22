# kubernetes-node-terminator

kubernetes-node-terminator periodically polls the api server for nodes which have the label `status=unhealthy`.  If the current number of unhealthy nodes is
less than a configured maximum threshold then the nodes will be terminated with a delay in between each termination.

We have an agent that we run on every kubernetes node which periodically checks that health of the node.   More often than not when the node fails the tests the underlying problem is related to the docker container engine.   Sometimes the issue is what we call a soft hang and can be `fixed` by restarting the docker service assuming you use the `--live-restore` flag with your docker daemon.   Other times docker is hung hard and cannot be restarted which leaves the node in a situation where the running pods are unmanageable.   In these situations the test agent labels the node with `status=unhealthy` and waits to be terminated.

# Requirements

* AWS is the only supported cloud provider and the nodes running the pod need to be allowed to perform the action `"ec2:TerminateInstances"`.

* Events are sent to datadog via a kubernetes service named dd-agent which is found via the `DD_AGENT_SERVICE_HOST` environmental variable.

* Kubernetes authentication is done using service tokens.  Point the variable `KUBERNETES_SERVICE_HOST` to your api server if it's not in the default loation.

* Some process is labeling nodes `status=unhealthy` when they need to be terminated.

  For example:
  `kubectl label node ip-10-20-14-192.ec2.internal status=unhealthy`

# Configuration

Control the configuration by setting the following environmental variables

```
KUBERNETES_SERVICE_HOST = Location of the api server.  You only need to set this if you are not using the default kubernetes service for the api server.
DD_AGENT_SERVICE_HOST = This get's set by kubernetes automatically if you have a service called dd-agent.
DRY_RUN = Terminate a node or just log what node would have been terminated.
INTERVAL_SECONDS = Frequency that we poll for unhealthy nodes.   Default is 60 seconds.
DELAY_BETWEEN_TERMINATIONS = When terminating multiple nodes how many seconds should we delay between each termination.  Default is 300 seconds.
AWS_REGION = What AWS region is the pod running in.
MAX_UNHEALTHY = No action will be taken if the current number of unhealthy nodes is greater than this setting.   The default is 1.
HEALTH_PORT = What port to listen on for http health checks.  The default is 8080.
```

# Deploy
To deploy kubernetes-node-terminator you can modify the [terminator-deployment.yaml](terminator-deployment.yaml) to your needs and then run

```
kubectl apply -f terminator-deployment.yaml
```
