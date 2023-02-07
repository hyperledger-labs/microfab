package microfab

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the microfab image running",
	RunE: func(cmd *cobra.Command, args []string) error {
		return start()
	},
}

func init() {

}

func start() error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrapf(err, "Unable to create Docker client")
	}
	defer cli.Close()
	var env = make([]string, 2, 200)

	log.Printf("Starting microfab container..\n")

	cfg = viper.GetString("MICROFAB_CONFIG")

	if cfg == "" {
		return errors.Errorf("Can't start - config is blank")
	}

	env[0] = "FABRIC_LOGGING_SPEC=info"
	env[1] = fmt.Sprintf("MICROFAB_CONFIG=%s", cfg)
	microFabImage := "ghcr.io/hyperledger-labs/microfab:latest"
	containername := "microfab"

	// pull down the image if needed
	err = DownloadImage(microFabImage)
	if err != nil {
		return err
	}

	// check to see if a container is allready running
	running, err := ImageRunning(containername)
	if err != nil {
		return err
	}

	if running {
		if force {
			if err = Stop(containername); err != nil {
				return err
			}
		} else {
			return errors.Errorf("Unable to start '%s' is already running: use --force", containername)
		}

	}
	// create the configuration for the container, primarily exposing port 8080
	config := &container.Config{
		Image:        microFabImage,
		ExposedPorts: nat.PortSet{"8080": struct{}{}},
		Env:          env,
	}

	hostConfig := &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{nat.Port("8080"): {{HostIP: "127.0.0.1", HostPort: "8080"}}},
		AutoRemove:   true,
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containername)
	if err != nil {

		fmt.Printf("%v  %v\n", resp, err)
		return errors.Wrapf(err, "Unable to create container")
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return errors.Wrapf(err, "Unable to start contianer")
	}

	log.Printf("Container ID %s\n", resp.ID)
	log.Printf("Microfab is up and running\n")

	return nil
}
