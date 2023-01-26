package registry

import (
	"context"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

func (c *Client) PullImage(imageName string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Err(err)
		return err
	}
	defer cli.Close()

	response, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		log.Err(err)
		return err
	}

	defer response.Close()

	// the pull request will be aborted prematurely unless the response is read
	if _, err = ioutil.ReadAll(response); err != nil {
		log.Err(err)
		return err
	}
	return nil
}
