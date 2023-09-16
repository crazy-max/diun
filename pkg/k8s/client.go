package k8s

import (
	"context"
	"os"

	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client represents an active kubernetes object
type Client struct {
	ctx        context.Context
	namespaces []string
	API        *kubernetes.Clientset
}

// Options holds kubernetes client object options
type Options struct {
	Endpoint         string
	Token            string
	TokenFile        string
	CertAuthFilePath string
	TLSInsecure      *bool
	Namespaces       []string
}

// New initializes a new Kubernetes client
func New(opts Options) (*Client, error) {
	var err error
	var api *kubernetes.Clientset

	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "":
		log.Debug().Msgf("Creating in-cluster Kubernetes provider client %s", opts.Endpoint)
		api, err = newInClusterClient(opts)
	case os.Getenv("KUBECONFIG") != "":
		log.Debug().Msgf("Creating cluster-external Kubernetes provider client from KUBECONFIG %s", os.Getenv("KUBECONFIG"))
		api, err = newExternalClusterClientFromFile(opts, os.Getenv("KUBECONFIG"))
	default:
		log.Debug().Msgf("Creating cluster-external Kubernetes provider client %s", opts.Endpoint)
		api, err = newExternalClusterClient(opts)
	}

	if len(opts.Namespaces) == 0 {
		opts.Namespaces = []string{metav1.NamespaceAll}
	}

	return &Client{
		ctx:        context.Background(),
		namespaces: opts.Namespaces,
		API:        api,
	}, err
}

func newInClusterClient(opts Options) (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create in-cluster configuration")
	}

	if opts.Endpoint != "" {
		config.Host = opts.Endpoint
	}
	if opts.TLSInsecure != nil {
		config.TLSClientConfig.Insecure = *opts.TLSInsecure
	}

	return kubernetes.NewForConfig(config)
}

func newExternalClusterClientFromFile(opts Options, file string) (*kubernetes.Clientset, error) {
	configFromFlags, err := clientcmd.BuildConfigFromFlags("", file)
	if err != nil {
		return nil, err
	}
	if opts.TLSInsecure != nil {
		configFromFlags.TLSClientConfig.Insecure = *opts.TLSInsecure
	}

	return kubernetes.NewForConfig(configFromFlags)
}

func newExternalClusterClient(opts Options) (*kubernetes.Clientset, error) {
	var err error

	if opts.Endpoint == "" {
		return nil, errors.New("endpoint missing for external cluster client")
	}

	opts.Token, err = utl.GetSecret(opts.Token, opts.TokenFile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot retrieve bearer token")
	}

	config := &rest.Config{
		Host:        opts.Endpoint,
		BearerToken: opts.Token,
	}

	if opts.CertAuthFilePath != "" {
		caData, err := os.ReadFile(opts.CertAuthFilePath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read CA file")
		}
		config.TLSClientConfig = rest.TLSClientConfig{
			CAData: caData,
		}
	}
	if opts.TLSInsecure != nil {
		config.TLSClientConfig.Insecure = *opts.TLSInsecure
	}

	return kubernetes.NewForConfig(config)
}
