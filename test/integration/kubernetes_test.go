//go:build integration

package integration

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	provider "github.com/crazy-max/diun/v4/internal/provider/kubernetes"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

func kubernetesClient(t *testing.T) *kubernetes.Clientset {
	t.Helper()
	// K3s writes a loopback endpoint; clients use its Compose service name.
	kubeconfig, err := clientcmd.LoadFromFile("/kubeconfig/config.yaml")
	require.NoError(t, err)
	for _, cluster := range kubeconfig.Clusters {
		cluster.Server = "https://kubernetes:6443"
	}
	filename := filepath.Join(t.TempDir(), "kubeconfig")
	require.NoError(t, clientcmd.WriteToFile(*kubeconfig, filename))
	t.Setenv("KUBECONFIG", filename)
	cfg, err := clientcmd.BuildConfigFromFlags("", filename)
	require.NoError(t, err)
	cfg.Timeout = 10 * time.Second
	client, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)
	return client
}

func TestKubernetesProvider(t *testing.T) {
	client := kubernetesClient(t)
	var namespaces []string
	for range 2 {
		ns, err := client.CoreV1().Namespaces().Create(t.Context(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "diun-integration-"},
		}, metav1.CreateOptions{})
		require.NoError(t, err)
		namespaces = append(namespaces, ns.Name)
		t.Cleanup(func() {
			require.NoError(t, client.CoreV1().Namespaces().Delete(context.Background(), ns.Name, metav1.DeleteOptions{}))
		})
	}
	data, err := os.ReadFile("/src/test/fixtures/kubernetes1/pods.yml")
	require.NoError(t, err)
	var pods corev1.PodList
	require.NoError(t, yaml.Unmarshal(data, &pods))
	for _, ns := range namespaces {
		_, err := client.CoreV1().ServiceAccounts(ns).Create(t.Context(), &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "diun"},
		}, metav1.CreateOptions{})
		require.NoError(t, err)
		for _, pod := range pods.Items {
			_, err := client.CoreV1().Pods(ns).Create(t.Context(), pod.DeepCopy(), metav1.CreateOptions{})
			require.NoError(t, err)
		}
	}

	cases := []struct {
		name           string
		namespaces     []string
		watchByDefault bool
		want           []string
	}{
		{
			name:       "annotations",
			namespaces: namespaces[:1],
			want:       []string{"enabled/busybox", "enabled/alpine"},
		},
		{
			name:           "watch by default",
			namespaces:     namespaces[:1],
			watchByDefault: true,
			want:           []string{"enabled/busybox", "enabled/alpine", "unlabeled/busybox"},
		},
		{
			name:       "exclude namespace",
			namespaces: []string{namespaces[0], namespaces[1], "!" + namespaces[1]},
			want:       []string{"enabled/busybox", "enabled/alpine"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&model.PrdKubernetes{}).GetDefaults()
			cfg.Namespaces = tc.namespaces
			cfg.WatchByDefault = new(tc.watchByDefault)
			jobs := provider.New(cfg, (&model.Defaults{}).GetDefaults()).ListJob()
			var got []string
			for _, job := range jobs {
				require.Equal(t, "kubernetes", job.Provider)
				require.Equal(t, namespaces[0], job.Image.Metadata["pod_namespace"])
				got = append(got, job.Image.Metadata["pod_name"]+"/"+job.Image.Metadata["ctn_name"])
			}
			require.ElementsMatch(t, tc.want, got)
		})
	}
	t.Run("application", func(t *testing.T) {
		t.Setenv("DIUN_PROVIDERS_KUBERNETES_NAMESPACES", namespaces[0])
		runApplicationFixture(t, "kubernetes1", "docker.io/library/busybox", "1.36.1")
	})
}

func TestLocalDigestKubernetes(t *testing.T) {
	requireLocalDigestFixture(t)
	f := newRegistryFixture(t)
	publish := runnablePublisher(t, f)
	old := publish("A")
	image := f.image("latest").Name
	client := kubernetesClient(t)
	ns, err := client.CoreV1().Namespaces().Create(t.Context(), &corev1.Namespace{GenerateName: "diun-digest-"}, metav1.CreateOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.CoreV1().Namespaces().Delete(context.Background(), ns.Name, metav1.DeleteOptions{}))
	})
	deployment, err := client.AppsV1().Deployments(ns.Name).Create(t.Context(), &appsv1.Deployment{
		Name: "digest",
		Spec: appsv1.DeploymentSpec{
			Replicas: new(int32(2)),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "digest"}},
			Strategy: appsv1.DeploymentStrategy{
				Type:          appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{MaxSurge: new(intstr.FromInt32(1)), MaxUnavailable: new(intstr.FromInt32(0))},
			},
			Template: corev1.PodTemplateSpec{
				Labels:      map[string]string{"app": "digest"},
				Annotations: map[string]string{"diun.enable": "true", "diun.regopt": "local", "revision": "A"},
				Spec:        corev1.PodSpec{Containers: []corev1.Container{{Name: "sleeper", Image: image, ImagePullPolicy: corev1.PullAlways}}},
			},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	versions := map[string]runtimeImage{"A": old}
	waitImages := func(want map[string]int) {
		t.Helper()
		var last []corev1.Pod
		defer func() {
			if t.Failed() {
				t.Logf("last Pods: %+v", last)
			}
		}()
		require.Eventually(t, func() bool {
			pods, err := client.CoreV1().Pods(ns.Name).List(t.Context(), metav1.ListOptions{LabelSelector: "app=digest"})
			if err != nil {
				return false
			}
			last = pods.Items
			got := map[string]int{}
			for _, pod := range pods.Items {
				for _, status := range pod.Status.ContainerStatuses {
					if status.State.Running == nil {
						continue
					}
					version := "unknown:" + status.ImageID
					for name, image := range versions {
						if strings.HasSuffix(status.ImageID, image.index.String()) || strings.HasSuffix(status.ImageID, image.manifest.String()) {
							version = name
						}
					}
					got[version]++
				}
			}
			return maps.Equal(got, want)
		}, 4*time.Minute, 500*time.Millisecond, "Kubernetes did not reach expected running versions: %v", want)
		for _, pod := range last {
			for _, status := range pod.Status.ContainerStatuses {
				t.Logf("Pod %s runtime imageID=%s", pod.Name, status.ImageID)
			}
		}
	}
	waitImages(map[string]int{"A": 2})
	cfg := (&model.PrdKubernetes{}).GetDefaults()
	cfg.Namespaces = []string{ns.Name}
	diun := startDigestWatcher(t, f, &model.Providers{Kubernetes: cfg})
	expectDigestMetric(t, "kubernetes", 0, 1)
	versions["B"] = publish("B")
	diun.Run()
	expectDigestMetric(t, "kubernetes", 1, 2)
	deployment, err = client.AppsV1().Deployments(ns.Name).Get(t.Context(), deployment.Name, metav1.GetOptions{})
	require.NoError(t, err)
	deployment.Spec.MinReadySeconds = 300
	deployment.Spec.Template.Annotations["revision"] = "B"
	_, err = client.AppsV1().Deployments(ns.Name).Update(t.Context(), deployment, metav1.UpdateOptions{})
	require.NoError(t, err)
	waitImages(map[string]int{"A": 2, "B": 1})
	expectDigestMetric(t, "kubernetes", 1, 2)
	deployment, err = client.AppsV1().Deployments(ns.Name).Get(t.Context(), deployment.Name, metav1.GetOptions{})
	require.NoError(t, err)
	deployment.Spec.MinReadySeconds = 0
	_, err = client.AppsV1().Deployments(ns.Name).Update(t.Context(), deployment, metav1.UpdateOptions{})
	require.NoError(t, err)
	waitImages(map[string]int{"B": 2})
	expectDigestMetric(t, "kubernetes", 0, 2)
}
