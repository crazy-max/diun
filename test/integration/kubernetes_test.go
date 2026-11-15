//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	provider "github.com/crazy-max/diun/v4/internal/provider/kubernetes"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

func TestKubernetesProvider(t *testing.T) {
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
