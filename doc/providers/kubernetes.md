# Kubernetes provider

* [About](#about)
* [Quick start](#quick-start)
* [Provider configuration](#provider-configuration)
  * [Configuration file](#configuration-file)
  * [Environment variables](#environment-variables)
* [Kubernetes annotations](#kubernetes-annotations)

## About

The Kubernetes provider allows you to analyze the pods of your Kubernetes cluster to extract images found and check for updates on the registry.

## Quick start

In this section we quickly go over a basic deployment using your local Kubernetes cluster.

Here we use our local Kubernetes provider with a minimum configuration to analyze annotated pods (watch by default disabled).

Now let's create a simple pod for Diun:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: diun
spec:
  replicas: 1
  selector:
    matchLabels:
      app: diun
  template:
    metadata:
      labels:
        app: diun
    spec:
      containers:
      - name: diun
        image: crazymax/diun:latest
        imagePullPolicy: Always
        env:
          - name: TZ
            value: "Europe/Paris"
          - name: LOG_LEVEL
            value: "info"
          - name: LOG_JSON
            value: "false"
          - name: DIUN_WATCH_WORKERS
            value: "20"
          - name: DIUN_WATCH_SCHEDULE
            value: "*/30 * * * *"
          - name: DIUN_PROVIDERS_KUBERNETES
            value: "true"
        volumeMounts:
          - mountPath: "/data"
            name: "data"
      restartPolicy: Always
      volumes:
        # Set up a data directory for gitea
        # For production usage, you should consider using PV/PVC instead(or simply using storage like NAS)
        # For more details, please see https://kubernetes.io/docs/concepts/storage/volumes/
      - name: "data"
        hostPath:
          path: "/data"
          type: Directory
```

And another one with a simple Nginx pod:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      run: nginx
  replicas: 2
  template:
    metadata:
      labels:
        run: nginx
      annotations:
        diun.enable: "true"
        diun.watch_repo: "true"
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
```

As an example we use [nginx](https://hub.docker.com/_/nginx/) Docker image. A few [annotations](#kubernetes-annotations) are added to configure the image analysis of this pod for Diun. We can now start these 2 pods:

```
kubectl apply -f diun.yml
kubectl apply -f nginx.yml
```

Now take a look at the logs:

```
$ kubectl logs -f -l app=diun --all-containers
# TODO: add logs example
```

## Provider configuration

### Configuration file

#### `endpoint`

The Kubernetes server endpoint as URL.

```yaml
providers:
  kubernetes:
    endpoint: "http://localhost:8080"
```

Kubernetes server endpoint as URL, which is only used when the behavior based on environment variables described below does not apply.

When deployed into Kubernetes, Diun reads the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to create the endpoint.

The access token is looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token` and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`. They are both provided automatically as mounts in the pod where Diun is deployed.

When the environment variables are not found, Diun tries to connect to the Kubernetes API server with an external-cluster client. In which case, the endpoint is required. Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes cluster using the granted authentication and authorization of the associated kubeconfig.

#### `token`

```yaml
providers:
  kubernetes:
    token: "atoken"
```

Bearer token used for the Kubernetes client configuration.

#### `tokenFile`

Use content of secret file as bearer token if `token` not defined.

```yaml
providers:
  kubernetes:
    tokenFile: "/run/secrets/token"
```

#### `certAuthFilePath`

Path to the certificate authority file. Used for the Kubernetes client configuration.

```yaml
providers:
  kubernetes:
    certAuthFilePath: "/a/ca.crt"
```

#### `tlsInsecure`

Controls whether client does not verify the server's certificate chain and hostname (default `false`).

```yaml
providers:
  kubernetes:
    tlsInsecure: false
```

#### `namespaces`

Array of namespaces to watch (default all namespaces).

```yaml
providers:
  kubernetes:
    namespaces:
      - default
      - production
```

#### `watchByDefault`

Enable watch by default. If false, pods that don't have `diun.enable: "true"` annotation will be ignored (default `false`).

```yaml
providers:
  kubernetes:
    watchByDefault: false
```

### Environment variables

* `DIUN_PROVIDERS_KUBERNETES`
* `DIUN_PROVIDERS_KUBERNETES_ENDPOINT`
* `DIUN_PROVIDERS_KUBERNETES_TOKEN`
* `DIUN_PROVIDERS_KUBERNETES_TOKENFILE`
* `DIUN_PROVIDERS_KUBERNETES_CERTAUTHFILEPATH`
* `DIUN_PROVIDERS_KUBERNETES_TLSINSECURE`
* `DIUN_PROVIDERS_KUBERNETES_NAMESPACES` (comma separated)
* `DIUN_PROVIDERS_KUBERNETES_WATCHBYDEFAULT`

## Kubernetes annotations

You can configure more finely the way to analyze the image of your pods through Kubernetes annotations:

* `diun.enable`: Set to true to enable image analysis of this pod.
* `diun.regopts_id`: Registry options ID from [`regopts`](../configuration.md#regopts) to use.
* `diun.watch_repo`: Watch all tags of this pod image (default `false`).
* `diun.max_tags`: Maximum number of tags to watch if `diun.watch_repo` enabled. 0 means all of them (default `0`).
* `diun.include_tags`: Semi-colon separated list of regular expressions to include tags. Can be useful if you enable `diun.watch_repo`.
* `diun.exclude_tags`: Semi-colon separated list of regular expressions to exclude tags. Can be useful if you enable `diun.watch_repo`.
