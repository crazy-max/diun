package k8s_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/crazy-max/diun/v4/pkg/k8s"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodList(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}

	os.Setenv("KUBECONFIG", "./.dev/minikube.config")

	kc, err := k8s.New(k8s.Options{
		TLSInsecure: utl.NewTrue(),
	})
	assert.NoError(t, err)
	assert.NotNil(t, kc)

	pods, err := kc.PodList(metav1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, pods)
	assert.True(t, len(pods) > 0)

	for _, pod := range pods {
		for _, ctn := range pod.Spec.Containers {
			fmt.Println(pod.Name, ctn.Image)
		}
	}
}
