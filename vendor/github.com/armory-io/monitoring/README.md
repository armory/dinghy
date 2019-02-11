# Monitoring
Tools for monitoring our internal applications.

## What is it?
It is a combination of a service that runs inside of the Kubernetes cluster and a Go package.

### Service
DataDog agents run inside the cluster via a daemon set. Additionally, there is a k8s service that is used for the pods to communicate with the agents.

### Library/Package
By default if you make a new DataDog monitor using this package, it will use the DataDog agents within the cluster.

## Repo Layout

| Path          | Description |
| ------------- |:-------------:|
| `./`          | Monitor interface |
| `./etc/`      | K8s DataDog agents and service |
| `./datadog/`  | DataDog implementation of the monitoring interface |
| `./mock/`     | Mock implementation of the monitoring interface. |
