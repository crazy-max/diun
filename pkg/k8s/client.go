package k8s

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client represents an active kubernetes object
type Client struct {
	ctx context.Context
	API *kubernetes.Clientset
}

// Options holds kubernetes client object options
type Options struct {
	Endpoint    string
	Token       string
	TokenFile   string
	TLSCAFile   string
	TLSInsecure *bool
}

// New initializes a new Kubernetes client
func New(opts Options) (*Client, error) {
	var err error
	var cl *kubernetes.Clientset

	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "":
		log.Debug().Msgf("Creating in-cluster Kubernetes provider client %s", opts.Endpoint)
		cl, err = newInClusterClient(opts)
	case os.Getenv("KUBECONFIG") != "":
		log.Debug().Msgf("Creating cluster-external Kubernetes provider client from KUBECONFIG %s", os.Getenv("KUBECONFIG"))
		cl, err = newExternalClusterClientFromFile(opts, os.Getenv("KUBECONFIG"))
	default:
		log.Debug().Msgf("Creating cluster-external Kubernetes provider client %s", opts.Endpoint)
		cl, err = newExternalClusterClient(opts)
	}

	return &Client{
		ctx: context.Background(),
		API: cl,
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

	configFromFlags.TLSClientConfig.Insecure = true
	return kubernetes.NewForConfig(configFromFlags)
}

func newExternalClusterClient(opts Options) (*kubernetes.Clientset, error) {
	var err error

	if opts.Endpoint == "" {
		return nil, errors.New("Endpoint missing for external cluster client")
	}

	opts.Token, err = utl.GetSecret(opts.Token, opts.TokenFile)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot retrieve bearer token")
	}

	config := &rest.Config{
		Host:        opts.Endpoint,
		BearerToken: opts.Token,
	}

	if opts.TLSCAFile != "" {
		caData, err := ioutil.ReadFile(opts.TLSCAFile)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to read CA file")
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
