# Run Heapster in a Kubernetes cluster with a Metricly Sink that pushes data to Metricly Cloud

### Setup a Kubernetes cluster
[Bring up a Kubernetes cluster](https://github.com/kubernetes/kubernetes), if you haven't already.
Ensure that you are able to interact with the cluster via `kubectl` (this may be `kubectl.sh` if using
the local-up-cluster in the Kubernetes repository).

### Start Metricly Heapster pod and service
Ensure that you have a valid checkout of [Heapster](https://github.com/metricly/heapster) and are in the root directory of the Heapster repository, and then run

```shell
$ kubectl create -f deploy/kube-config/metricly/heapster.yaml
```
See also the [Sink Configuration](sink-configuration.md) Metricly Section for more advanced configurations, e.g. filters.

## Troubleshooting guide

See also the [debugging documentation](debugging.md).

1. If there are no Kubernetes elements shown up in Metricly Inventory after about 5 minutes, heapster might not be running properly with the configured Metricly Sink. Use `kubectl` to verify that the `heapster` pod and service is alive and check the pod logs.
    ```
    $ kubectl get pods --namespace=kube-system
    ...
    heapster-5bb59685bd-75f4x                          1/1       Running    2          22d
    ...
    
    $ kubectl get services --namespace=kube-system heapster
    NAME       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
    heapster   ClusterIP   10.101.65.65   <none>        80/TCP    34d

    $ kubectl --namespace=kube-system logs heapster-5bb59685bd-75f4x
    ...
    I0517 15:08:05.134683       1 metricly.go:49] Start exporting data batch to Metricly ...
    I0517 15:08:05.241484       1 metricly.go:79] Exported 85 out of 85 elements using 5 workers
    ...
    ```
