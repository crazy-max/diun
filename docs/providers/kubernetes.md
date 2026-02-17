# Kubernetes provider

## About

The Kubernetes provider allows you to analyze the pods of your Kubernetes
cluster to extract images found and check for updates on the registry.

## Quick start

In this section, we quickly go over a basic deployment using your local
Kubernetes cluster.

Here we use our local Kubernetes provider with a minimum configuration to
analyze annotated pods (watch by default disabled).

Now let's create a simple pod for Diun:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: default
  name: diun
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: diun
rules:
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
      - watch
      - list
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: diun
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: diun
subjects:
  - kind: ServiceAccount
    name: diun
    namespace: default
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: default
  name: diun
spec:
  replicas: 1
  selector:
    matchLabels:
      app: diun
  template:
    metadata:
      labels:
        app: diun
      annotations:
        diun.enable: "true"
    spec:
      serviceAccountName: diun
      containers:
        - name: diun
          image: crazymax/diun:latest
          imagePullPolicy: Always
          args: ["serve"]
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
              value: "0 */6 * * *"
            - name: DIUN_WATCH_JITTER
              value: "30s"
            - name: DIUN_PROVIDERS_KUBERNETES
              value: "true"
          volumeMounts:
            - mountPath: "/data"
              name: "data"
      restartPolicy: Always
      volumes:
        # Set up a data directory for diun
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
  namespace: default
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
    spec:
      containers:
        - name: nginx
          image: nginx
          ports:
            - containerPort: 80
```

As an example we use [nginx](https://hub.docker.com/_/nginx/) Docker image. A
few [annotations](#kubernetes-annotations) are added to configure the image
analysis of this pod for Diun. We can now start these 2 pods:

```
kubectl apply -f diun.yml
kubectl apply -f nginx.yml
```

Now take a look at the logs:

```
$ kubectl logs -f -l app=diun --all-containers
Wed, 17 Jun 2020 10:49:58 CEST INF Starting Diun version=4.0.0-beta.3
Wed, 17 Jun 2020 10:49:58 CEST WRN No notifier available
Wed, 17 Jun 2020 10:49:58 CEST INF Cron triggered
Wed, 17 Jun 2020 10:49:59 CEST INF Found 1 image(s) to analyze provider=kubernetes
Wed, 17 Jun 2020 10:50:00 CEST INF New image found image=docker.io/library/nginx:latest provider=kubernetes
Wed, 17 Jun 2020 10:50:02 CEST INF New image found image=docker.io/library/nginx:1.9 provider=kubernetes
Wed, 17 Jun 2020 10:50:02 CEST INF New image found image=docker.io/library/nginx:1.9.5 provider=kubernetes
Wed, 17 Jun 2020 10:50:02 CEST INF New image found image=docker.io/library/nginx:1.9.7 provider=kubernetes
Wed, 17 Jun 2020 10:50:02 CEST INF New image found image=docker.io/library/nginx:1.9.9 provider=kubernetes
Wed, 17 Jun 2020 10:50:02 CEST INF New image found image=docker.io/library/nginx:1.9.4 provider=kubernetes
Wed, 17 Jun 2020 10:50:02 CEST INF New image found image=docker.io/library/nginx:1.9.6 provider=kubernetes
Wed, 17 Jun 2020 10:50:02 CEST INF New image found image=docker.io/library/nginx:1.9.8 provider=kubernetes
Wed, 17 Jun 2020 10:50:03 CEST INF New image found image=docker.io/library/nginx:stable provider=kubernetes
Wed, 17 Jun 2020 10:50:03 CEST INF New image found image=docker.io/library/nginx:stable-alpine provider=kubernetes
Wed, 17 Jun 2020 10:50:03 CEST INF New image found image=docker.io/library/nginx:perl provider=kubernetes
...
```

### Alternative: Run as CronJob

You don't need to continuously run Diun in the cluster. Instead, you can disable Diuns built-in scheduler and simply use a Kubernetes `CronJob`:

```yaml
# Still add ServiceAccount, ClusterRole and ClusterRoleBinding from the example above!
---
apiVersion: batch/v1
kind: CronJob
metadata:
  namespace: default
  name: diun
spec:
  schedule: "0 */6 * * *"
  jobTemplate:
    metadata:
      labels:
        app: diun
    spec:
      template:
        metadata:
          labels:
            app: diun
          annotations:
            diun.enable: "true"
        spec:
          serviceAccountName: diun
          containers:
            - name: diun
              image: crazymax/diun:latest
              imagePullPolicy: Always
              args: ["serve"]
              env:
                - name: TZ
                  value: "Europe/Paris"
                - name: LOG_LEVEL
                  value: "info"
                - name: LOG_JSON
                  value: "false"
                - name: DIUN_WATCH_WORKERS
                  value: "20"
                - name: DIUN_PROVIDERS_KUBERNETES
                  value: "true"
                - name: DIUN_WATCH_SCHEDULE
                  value: "" # NOTE: This is empty to disalbe built-in scheduling
              volumeMounts:
                - mountPath: "/data"
                  name: "data"
          restartPolicy: Always
          volumes:
            # Set up a data directory for diun
            # For production usage, you should consider using PV/PVC instead(or simply using storage like NAS)
            # For more details, please see https://kubernetes.io/docs/concepts/storage/volumes/
            - name: "data"
              hostPath:
                path: "/data"
                type: Directory
```

The key to this setup is setting `watch.schedule` (`DIUN_WATCH_SCHEDULE` via environment variable) to empty. If you're using the YAML/JSON config, you can explicitly set it to `null` as well. See the [schedule documentation](../config/watch.md#schedule) for more information.

## Configuration

!!! hint
    Environment variable `DIUN_PROVIDERS_KUBERNETES=true` can be used to enable this provider with default values.

### `endpoint`

The Kubernetes server endpoint as URL.

!!! example "File"
    ```yaml
    providers:
      kubernetes:
        endpoint: "http://localhost:8080"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_KUBERNETES_ENDPOINT`

Kubernetes server endpoint as URL, which is only used when the behavior based
on environment variables described below does not apply.

When deployed into Kubernetes, Diun reads the environment variables
`KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to
create the endpoint.

The access token is looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token`
and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
They are both provided automatically as mounts in the pod where Diun is deployed.

When the environment variables are not found, Diun tries to connect to the
Kubernetes API server with an external-cluster client. In which case, the
endpoint is required. Specifically, it may be set to the URL used by
`kubectl proxy` to connect to a Kubernetes cluster using the granted
authentication and authorization of the associated kubeconfig.

### `token`

Bearer token used for the Kubernetes client configuration.

!!! example "File"
    ```yaml
    providers:
      kubernetes:
        token: "atoken"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_KUBERNETES_TOKEN`

### `tokenFile`

Use content of secret file as bearer token if `token` not defined.

!!! example "File"
    ```yaml
    providers:
      kubernetes:
        tokenFile: "/run/secrets/token"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_KUBERNETES_TOKEN`

### `certAuthFilePath`

Path to the certificate authority file. Used for the Kubernetes client
configuration.

!!! example "File"
    ```yaml
    providers:
      kubernetes:
        certAuthFilePath: "/a/ca.crt"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_KUBERNETES_CERTAUTHFILEPATH`

### `tlsInsecure`

Controls whether client does not verify the server's certificate chain and
hostname (default `false`).

!!! example "File"
    ```yaml
    providers:
      kubernetes:
        tlsInsecure: false
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_KUBERNETES_TLSINSECURE`

### `namespaces`

Array of namespaces to watchBy default, it watches all namespaces. You can
limit monitoring to specific namespaces by listing them. This helps reduce
scope and focus on relevant pods only.

!!! example "File"
    ```yaml
    providers:
      kubernetes:
        namespaces:
          - default
          - production
    ```

You can also negate namespaces by prefixing them with `!` if you want to watch
all namespaces except specific ones:

!!! example "File"
    ```yaml
    providers:
      kubernetes:
        namespaces:
          - !kube-system
          - !kube-public
          - !kube-node-lease
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_KUBERNETES_NAMESPACES` (comma separated)

### `watchByDefault`

Enable watch by default. If false, pods that don't have `diun.enable: "true"`
annotation will be ignored (default `false`).

!!! example "File"
    ```yaml
    providers:
      kubernetes:
        watchByDefault: false
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_KUBERNETES_WATCHBYDEFAULT`

## Kubernetes annotations

You can configure more finely the way to analyze the image of your pods through
Kubernetes annotations:

| Name                | Default                        | Description                                                                                                                                             |
|---------------------|--------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| `diun.enable`       |                                | Set to true to enable image analysis of this pod                                                                                                        |
| `diun.regopt`       |                                | [Registry options](../config/regopts.md) name to use                                                                                                    |
| `diun.watch_repo`   | `false`                        | Watch all tags of this pod image ([be careful](../faq.md#docker-hub-rate-limits) with this setting)                                                     |
| `diun.notify_on`    | `new;update`                   | Semicolon separated list of status to be notified: `new`, `update`.                                                                                     |
| `diun.sort_tags`    | `reverse`                      | [Sort tags method](../faq.md#tags-sorting-when-using-watch_repo) if `diun.watch_repo` enabled. One of `default`, `reverse`, `semver`, `lexicographical` |
| `diun.max_tags`     | `0`                            | Maximum number of tags to watch if `diun.watch_repo` enabled. `0` means all of them                                                                     |
| `diun.include_tags` |                                | Semicolon separated list of regular expressions to include tags. Can be useful if you enable `diun.watch_repo`                                          |
| `diun.exclude_tags` |                                | Semicolon separated list of regular expressions to exclude tags. Can be useful if you enable `diun.watch_repo`                                          |
| `diun.hub_link`     | _automatic_                    | Set registry hub link for this image                                                                                                                    |
| `diun.platform`     | _automatic_                    | Platform to use (e.g. `linux/amd64`)                                                                                                                    |
| `diun.metadata.*`   | See [below](#default-metadata) | Additional metadata that can be used in [notification template](../faq.md#notification-template) (e.g. `diun.metadata.foo=bar`)                         |

## Default metadata

| Key                           | Description       |
|-------------------------------|-------------------|
| `diun.metadata.pod_name`      | Pod name          |
| `diun.metadata.pod_status`    | Pod status        |
| `diun.metadata.pod_namespace` | Pod namespace     |
| `diun.metadata.pod_createdat` | Pod creation date |
| `diun.metadata.ctn_name`      | Container name    |
| `diun.metadata.ctn_command`   | Container command |
