# kubectl-ips

A kubectl plugin that lists IP addresses from Kubernetes pods with flexible filtering options.

This plugin serves as both a useful tool for cluster administrators and developers who need to quickly identify pod IP addresses, and as an example of how to build a kubectl plugin using the same tools and helpers available to kubectl itself.

## Features

* Lists all pod IP addresses (including multiple IPs per pod)
* Supports namespace filtering
* Label selector support for pod filtering
* Multiple output formats: table (default), wide, JSON, YAML, and name-only
* Shows pod names and namespaces by default, with option to show only IPs
* Kubernetes-style table output with status, age, and node information
* Sorted output for consistency
* Handles both IPv4 and IPv6 addresses

## Installation

### From Source

```shell
go install github.com/andreygrechin/kubectl-ips/cmd/kubectl-ips@latest
```

### Manual Build

```shell
git clone https://github.com/andreygrechin/kubectl-ips.git
cd kubectl-ips
make build
```

Then copy the `bin/kubectl-ips` file to a directory in your `PATH`.

## Usage

### Basic Commands

List pod IPs in the current namespace:

```shell
kubectl ips
```

List pod IPs in all namespaces:

```shell
kubectl ips --all-namespaces
kubectl ips -A
```

List pod IPs in a specific namespace:

```shell
kubectl ips --namespace=kube-system
kubectl ips -n kube-system
```

### Output Formats

Output in JSON format:

```shell
kubectl ips -o json
```

Output in YAML format:

```shell
kubectl ips -o yaml
```

Show only pod names:

```shell
kubectl ips -o name
```

Hide table headers:

```shell
kubectl ips --no-headers
```

### Advanced Options

Show only IP addresses without pod names (legacy):

```shell
kubectl ips --show-ips-only
```

Filter pods by label selector:

```shell
kubectl ips --selector=app=nginx
kubectl ips -l app=nginx,env=production
```

Combine options:

```shell
kubectl ips -A --selector=app=nginx --no-headers
```

## Output Examples

Default table format:

```text
NAME                                 IP           STATUS    AGE
nginx-deployment-5d59d67564-8g7nm    10.244.0.5   Running   2d
nginx-deployment-5d59d67564-ktht2    10.244.1.3   Running   2d
nginx-deployment-5d59d67564-wnx8l    10.244.2.7   Running   2d
```

Wide format with additional information:

```text
NAME                                 IP           STATUS    READY   RESTARTS   NODE          AGE
nginx-deployment-5d59d67564-8g7nm    10.244.0.5   Running   1/1     0          worker-node-1 2d
nginx-deployment-5d59d67564-ktht2    10.244.1.3   Running   1/1     0          worker-node-2 2d
nginx-deployment-5d59d67564-wnx8l    10.244.2.7   Running   1/1     0          worker-node-3 2d
```

With `--all-namespaces`:

```text
NAMESPACE   NAME                                 IP           STATUS    AGE
default     nginx-deployment-5d59d67564-8g7nm    10.244.0.5   Running   2d
default     nginx-deployment-5d59d67564-ktht2    10.244.1.3   Running   2d
kube-system coredns-5d78c9869d-6mx58             10.244.0.2   Running   5d
```

With `--show-ips-only` flag (legacy format):

```text
10.244.0.5
10.244.1.3
10.244.2.7
```

## Command Line Options

### Filtering Options

* `--all-namespaces, -A`: List pods from all namespaces
* `--namespace, -n`: Specify a particular namespace
* `--selector, -l`: Filter pods using label selectors

### Output Options

* `--output, -o`: Output format (table, wide, json, yaml, name)
* `--no-headers`: Don't print column headers
* `--show-labels`: Show labels as the last column
* `--show-ips-only`: Display only IP addresses without pod names (legacy)

### Standard Options

* Standard kubectl flags like `--kubeconfig`, `--context`, etc.

## Implementation Details

The plugin uses the [client-go library](https://github.com/kubernetes/client-go) to connect to the Kubernetes cluster and list pods. It makes use of the genericclioptions in [k8s.io/cli-runtime](https://github.com/kubernetes/cli-runtime) to generate a set of configuration flags which are in turn used to connect to the Kubernetes API server and retrieve pod information.

Key features of the implementation:

* Creates a custom kubectl command that follows standard patterns
* Uses the Kubernetes client-go library to interact with the cluster
* Lists and filters resources across namespaces
* Extracts IP addresses from each pod's status, supporting both single and multiple IP addresses per pod (dual-stack networking)
* By default shows pod names and namespaces for better context, with option to show only IP addresses
* Results are sorted for consistent output

## Requirements

* `kubectl` installed and configured
* Go 1.24+ (for building from source)
* Access to a Kubernetes cluster
* Appropriate RBAC permissions to list pods

## Development

Run tests:

```shell
make test
```

Format code:

```shell
make format
```

Run linter:

```shell
make lint
```

## Cleanup

You can "uninstall" this plugin from kubectl by simply removing it from your PATH. In case of using "go install", you can also run:

```shell
go clean -i github.com/andreygrechin/kubectl-ips/cmd/kubectl-ips
```
