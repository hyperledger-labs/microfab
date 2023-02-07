package microfab

import (
	"context"
	"log"

	"github.com/pkg/errors"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stops the microfab image running",
	GroupID: "mf",
	RunE: func(cmd *cobra.Command, args []string) error {
		return Stop("microfab")
	},
}

// Stop stops the container
func Stop(containername string) error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	log.Printf("Attempting to stop the container")

	if err := cli.ContainerStop(ctx, containername, container.StopOptions{}); err != nil {
		return errors.Wrapf(err, "Unable to stop container %s: %s", containername, err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, containername, container.WaitConditionRemoved)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	log.Printf("Container stopped and removed")

	return nil
}
