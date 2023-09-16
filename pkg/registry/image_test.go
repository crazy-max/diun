package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImage(t *testing.T) {
	testCases := []struct {
		desc      string
		parseOpts ParseImageOptions
		expected  Image
	}{
		{
			desc: "bintray artifactory-oss",
			parseOpts: ParseImageOptions{
				Name: "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
			},
			expected: Image{
				Domain: "jfrog-docker-reg2.bintray.io",
				Path:   "jfrog/artifactory-oss",
				Tag:    "4.0.0",
			},
		},
		{
			desc: "bintray xray-server",
			parseOpts: ParseImageOptions{
				Name: "docker.bintray.io/jfrog/xray-server:2.8.6",
			},
			expected: Image{
				Domain: "docker.bintray.io",
				Path:   "jfrog/xray-server",
				Tag:    "2.8.6",
			},
		},
		{
			desc: "dockerhub alpine",
			parseOpts: ParseImageOptions{
				Name: "alpine",
			},
			expected: Image{
				Domain: "docker.io",
				Path:   "library/alpine",
				Tag:    "latest",
			},
		},
		{
			desc: "dockerhub crazymax/nextcloud",
			parseOpts: ParseImageOptions{
				Name: "docker.io/crazymax/nextcloud:latest",
			},
			expected: Image{
				Domain: "docker.io",
				Path:   "crazymax/nextcloud",
				Tag:    "latest",
			},
		},
		{
			desc: "gcr busybox",
			parseOpts: ParseImageOptions{
				Name: "gcr.io/google-containers/busybox:latest",
			},
			expected: Image{
				Domain: "gcr.io",
				Path:   "google-containers/busybox",
				Tag:    "latest",
			},
		},
		{
			desc: "gcr busybox tag/digest",
			parseOpts: ParseImageOptions{
				Name: "gcr.io/google-containers/busybox:latest" + sha256digest,
			},
			expected: Image{
				Domain: "gcr.io",
				Path:   "google-containers/busybox",
				Tag:    "latest",
				Digest: sha256digest,
			},
		},
		{
			desc: "github ddns-route53",
			parseOpts: ParseImageOptions{
				Name: "docker.pkg.github.com/crazy-max/ddns-route53/ddns-route53:latest",
			},
			expected: Image{
				Domain: "docker.pkg.github.com",
				Path:   "crazy-max/ddns-route53/ddns-route53",
				Tag:    "latest",
			},
		},
		{
			desc: "gitlab meltano",
			parseOpts: ParseImageOptions{
				Name: "registry.gitlab.com/meltano/meltano",
			},
			expected: Image{
				Domain: "registry.gitlab.com",
				Path:   "meltano/meltano",
				Tag:    "latest",
			},
		},
		{
			desc: "quay hypercube",
			parseOpts: ParseImageOptions{
				Name: "quay.io/coreos/hyperkube",
			},
			expected: Image{
				Domain: "quay.io",
				Path:   "coreos/hyperkube",
				Tag:    "latest",
			},
		},
		{
			desc: "ghcr ddns-route53",
			parseOpts: ParseImageOptions{
				Name: "ghcr.io/crazy-max/ddns-route53",
			},
			expected: Image{
				Domain: "ghcr.io",
				Path:   "crazy-max/ddns-route53",
				Tag:    "latest",
			},
		},
		{
			desc: "ghcr radarr",
			parseOpts: ParseImageOptions{
				Name: "ghcr.io/linuxserver/radarr",
			},
			expected: Image{
				Domain: "ghcr.io",
				Path:   "linuxserver/radarr",
				Tag:    "latest",
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			img, err := ParseImage(tt.parseOpts)
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
		parseOpts ParseImageOptions
		expected  string
	}{
		{
			desc: "bintray artifactory-oss",
			parseOpts: ParseImageOptions{
				Name: "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
			},
			expected: "https://bintray.com/jfrog/reg2/jfrog%3Aartifactory-oss",
		},
		{
			desc: "bintray kubexray",
			parseOpts: ParseImageOptions{
				Name: "jfrog-docker-reg2.bintray.io/kubexray:latest",
			},
			expected: "https://bintray.com/jfrog/reg2/kubexray",
		},
		{
			desc: "bintray xray-server",
			parseOpts: ParseImageOptions{
				Name: "docker.bintray.io/jfrog/xray-server:2.8.6",
			},
			expected: "https://bintray.com/jfrog/reg2/jfrog%3Axray-server",
		},
		{
			desc: "dockerhub alpine",
			parseOpts: ParseImageOptions{
				Name: "alpine",
			},
			expected: "https://hub.docker.com/_/alpine",
		},
		{
			desc: "dockerhub crazymax/nextcloud",
			parseOpts: ParseImageOptions{
				Name: "docker.io/crazymax/nextcloud:latest",
			},
			expected: "https://hub.docker.com/r/crazymax/nextcloud",
		},
		{
			desc: "gcr busybox",
			parseOpts: ParseImageOptions{
				Name: "gcr.io/google-containers/busybox:latest",
			},
			expected: "https://gcr.io/google-containers/busybox",
		},
		{
			desc: "github ddns-route53",
			parseOpts: ParseImageOptions{
				Name: "docker.pkg.github.com/crazy-max/ddns-route53/ddns-route53:latest",
			},
			expected: "https://github.com/crazy-max/ddns-route53/packages",
		},
		{
			desc: "gitlab meltano",
			parseOpts: ParseImageOptions{
				Name: "registry.gitlab.com/meltano/meltano",
			},
			expected: "https://gitlab.com/meltano/meltano/container_registry",
		},
		{
			desc: "quay hypercube",
			parseOpts: ParseImageOptions{
				Name: "quay.io/coreos/hyperkube",
			},
			expected: "https://quay.io/repository/coreos/hyperkube",
		},
		{
			desc: "ghcr ddns-route53",
			parseOpts: ParseImageOptions{
				Name: "ghcr.io/crazy-max/ddns-route53",
			},
			expected: "https://github.com/users/crazy-max/packages/container/package/ddns-route53",
		},
		{
			desc: "ghcr radarr",
			parseOpts: ParseImageOptions{
				Name: "ghcr.io/linuxserver/radarr",
			},
			expected: "https://github.com/users/linuxserver/packages/container/package/radarr",
		},
		{
			desc: "redhat etcd",
			parseOpts: ParseImageOptions{
				Name: "registry.access.redhat.com/rhel7/etcd",
			},
			expected: "https://access.redhat.com/containers/#/registry.access.redhat.com/rhel7/etcd",
		},
		{
			desc: "private",
			parseOpts: ParseImageOptions{
				Name:   "myregistry.example.com/an/image:latest",
				HubTpl: "https://{{ .Domain }}/ui/repos/{{ .Path }}",
			},
			expected: "https://myregistry.example.com/ui/repos/an/image",
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			img, err := ParseImage(tt.parseOpts)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.expected, img.HubLink)
		})
	}
}
