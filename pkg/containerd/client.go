package containerd

import (
	"context"
	"sort"
	"strings"
	"time"

	containersapi "github.com/containerd/containerd/api/services/containers/v1"
	tasksapi "github.com/containerd/containerd/api/services/tasks/v1"
	versionapi "github.com/containerd/containerd/api/services/version/v1"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultTimeout  = 10 * time.Second
	namespaceHeader = "containerd-namespace"
)

// Client represents an active containerd object.
type Client struct {
	ctx          context.Context
	conn         *grpc.ClientConn
	containerAPI containersapi.ContainersClient
	taskAPI      tasksapi.TasksClient
}

// Options holds containerd client object options.
type Options struct {
	Endpoint string
	Timeout  time.Duration
}

// New initializes a new containerd API client with default values.
func New(opts Options) (*Client, error) {
	endpoint := normalizeEndpoint(opts.Endpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	conn, err := grpc.NewClient(
		dialAddress(endpoint),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(contextDialer),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create containerd gRPC client")
	}

	ctx := context.Background()
	pingCtx, cancel := context.WithTimeoutCause(ctx, timeout, errors.WithStack(context.DeadlineExceeded))
	defer cancel()

	if _, err := versionapi.NewVersionClient(conn).Version(pingCtx, &emptypb.Empty{}, grpc.WaitForReady(true)); err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "failed to connect to containerd")
	}

	return &Client{
		ctx:          ctx,
		conn:         conn,
		containerAPI: containersapi.NewContainersClient(conn),
		taskAPI:      tasksapi.NewTasksClient(conn),
	}, nil
}

// Close closes the containerd client.
func (c *Client) Close() error {
	return c.conn.Close()
}

// ContainerList returns containerd containers for a namespace.
func (c *Client) ContainerList(namespace string) ([]*containersapi.Container, error) {
	resp, err := c.containerAPI.List(withNamespace(c.ctx, namespace), &containersapi.ListContainersRequest{})
	if err != nil {
		return nil, err
	}

	ctns := resp.GetContainers()
	sort.Slice(ctns, func(i, j int) bool {
		if ctns[i].Image == ctns[j].Image {
			return ctns[i].ID < ctns[j].ID
		}
		return ctns[i].Image < ctns[j].Image
	})

	return ctns, nil
}

// TaskList returns containerd tasks for a namespace.
func (c *Client) TaskList(namespace string) ([]*tasktypes.Process, error) {
	resp, err := c.taskAPI.List(withNamespace(c.ctx, namespace), &tasksapi.ListTasksRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetTasks(), nil
}

func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "unix://")
	endpoint = strings.TrimPrefix(endpoint, "npipe://")
	return endpoint
}

func withNamespace(ctx context.Context, namespace string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, namespaceHeader, namespace)
}
