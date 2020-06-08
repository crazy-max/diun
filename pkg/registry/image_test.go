package registry_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
)

func TestParseImage(t *testing.T) {
	testCases := []struct {
		desc      string
		parseOpts registry.ParseImageOptions
		expected  registry.Image
	}{
		{
			desc: "bintray artifactory-oss",
			parseOpts: registry.ParseImageOptions{
				Name: "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
			},
			expected: registry.Image{
				Domain: "jfrog-docker-reg2.bintray.io",
				Path:   "jfrog/artifactory-oss",
				Tag:    "4.0.0",
			},
		},
		{
			desc: "bintray xray-server",
			parseOpts: registry.ParseImageOptions{
				Name: "docker.bintray.io/jfrog/xray-server:2.8.6",
			},
			expected: registry.Image{
				Domain: "docker.bintray.io",
				Path:   "jfrog/xray-server",
				Tag:    "2.8.6",
			},
		},
		{
			desc: "dockerhub alpine",
			parseOpts: registry.ParseImageOptions{
				Name: "alpine",
			},
			expected: registry.Image{
				Domain: "docker.io",
				Path:   "library/alpine",
				Tag:    "latest",
			},
		},
		{
			desc: "dockerhub crazymax/nextcloud",
			parseOpts: registry.ParseImageOptions{
				Name: "docker.io/crazymax/nextcloud:latest",
			},
			expected: registry.Image{
				Domain: "docker.io",
				Path:   "crazymax/nextcloud",
				Tag:    "latest",
			},
		},
		{
			desc: "gcr busybox",
			parseOpts: registry.ParseImageOptions{
				Name: "gcr.io/google-containers/busybox:latest",
			},
			expected: registry.Image{
				Domain: "gcr.io",
				Path:   "google-containers/busybox",
				Tag:    "latest",
			},
		},
		{
			desc: "github ddns-route53",
			parseOpts: registry.ParseImageOptions{
				Name: "docker.pkg.github.com/crazy-max/ddns-route53/ddns-route53:latest",
			},
			expected: registry.Image{
				Domain: "docker.pkg.github.com",
				Path:   "crazy-max/ddns-route53/ddns-route53",
				Tag:    "latest",
			},
		},
		{
			desc: "gitlab meltano",
			parseOpts: registry.ParseImageOptions{
				Name: "registry.gitlab.com/meltano/meltano",
			},
			expected: registry.Image{
				Domain: "registry.gitlab.com",
				Path:   "meltano/meltano",
				Tag:    "latest",
			},
		},
		{
			desc: "quay hypercube",
			parseOpts: registry.ParseImageOptions{
				Name: "quay.io/coreos/hyperkube",
			},
			expected: registry.Image{
				Domain: "quay.io",
				Path:   "coreos/hyperkube",
				Tag:    "latest",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			img, err := registry.ParseImage(tt.parseOpts)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.expected.Domain, img.Domain)
			assert.Equal(t, tt.expected.Path, img.Path)
			assert.Equal(t, tt.expected.Tag, img.Tag)
		})
	}
}

func TestHubLink(t *testing.T) {
	testCases := []struct {
		desc      string
		parseOpts registry.ParseImageOptions
		expected  string
	}{
		{
			desc: "bintray artifactory-oss",
			parseOpts: registry.ParseImageOptions{
				Name: "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
			},
			expected: "https://bintray.com/jfrog/reg2/jfrog%3Aartifactory-oss",
		},
		{
			desc: "bintray kubexray",
			parseOpts: registry.ParseImageOptions{
				Name: "jfrog-docker-reg2.bintray.io/kubexray:latest",
			},
			expected: "https://bintray.com/jfrog/reg2/kubexray",
		},
		{
			desc: "bintray xray-server",
			parseOpts: registry.ParseImageOptions{
				Name: "docker.bintray.io/jfrog/xray-server:2.8.6",
			},
			expected: "https://bintray.com/jfrog/reg2/jfrog%3Axray-server",
		},
		{
			desc: "dockerhub alpine",
			parseOpts: registry.ParseImageOptions{
				Name: "alpine",
			},
			expected: "https://hub.docker.com/_/alpine",
		},
		{
			desc: "dockerhub crazymax/nextcloud",
			parseOpts: registry.ParseImageOptions{
				Name: "docker.io/crazymax/nextcloud:latest",
			},
			expected: "https://hub.docker.com/r/crazymax/nextcloud",
		},
		{
			desc: "gcr busybox",
			parseOpts: registry.ParseImageOptions{
				Name: "gcr.io/google-containers/busybox:latest",
			},
			expected: "https://gcr.io/google-containers/busybox",
		},
		{
			desc: "github ddns-route53",
			parseOpts: registry.ParseImageOptions{
				Name: "docker.pkg.github.com/crazy-max/ddns-route53/ddns-route53:latest",
			},
			expected: "https://github.com/crazy-max/ddns-route53/packages",
		},
		{
			desc: "gitlab meltano",
			parseOpts: registry.ParseImageOptions{
				Name: "registry.gitlab.com/meltano/meltano",
			},
			expected: "https://gitlab.com/meltano/meltano/container_registry",
		},
		{
			desc: "quay hypercube",
			parseOpts: registry.ParseImageOptions{
				Name: "quay.io/coreos/hyperkube",
			},
			expected: "https://quay.io/repository/coreos/hyperkube",
		},
		{
			desc: "redhat etcd",
			parseOpts: registry.ParseImageOptions{
				Name: "registry.access.redhat.com/rhel7/etcd",
			},
			expected: "https://access.redhat.com/containers/#/registry.access.redhat.com/rhel7/etcd",
		},
		{
			desc: "private",
			parseOpts: registry.ParseImageOptions{
				Name:   "myregistry.example.com/an/image:latest",
				HubTpl: "https://{{ .Domain }}/ui/repos/{{ .Path }}",
			},
			expected: "https://myregistry.example.com/ui/repos/an/image",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			img, err := registry.ParseImage(tt.parseOpts)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.expected, img.HubLink)
		})
	}
}
