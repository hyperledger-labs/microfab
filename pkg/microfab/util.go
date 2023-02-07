package microfab

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// PullStatus of the JSON structure returned by docker
type PullStatus struct {
	Status string `json:"status"`
}

// DownloadImage pulls the image if needed
func DownloadImage(microFabImage string) error {
	ctx := context.Background()
	log.Printf("Checking for any image already")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrapf(err, "Unable to create Docker client")
	}
	defer cli.Close()

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return errors.Wrapf(err, "Unable to list images")
	}

	found := false
	for _, image := range images {
		found = slices.Contains(image.RepoTags, microFabImage)
		if found {
			break
		}
	}

	if !found {
		log.Printf("Pulling image %s", microFabImage)
		out, err := cli.ImagePull(ctx, microFabImage, types.ImagePullOptions{})
		if err != nil {
			return errors.Wrapf(err, "Unable to pull images")
		}

		defer out.Close()

		// rather inelegant way of getting status - effectively scanning
		// the stdout of the command
		buf := bufio.NewScanner(out)
		for buf.Scan() {
			var s PullStatus
			json.Unmarshal(buf.Bytes(), &s)
			if strings.HasPrefix(s.Status, "Status: Downloaded") {
				log.Printf(s.Status)
			}

		}
	} else {
		log.Printf("Found image %s", microFabImage)
	}

	return nil
}

// ImageRunning determines if the image is already running
func ImageRunning(containerName string) (bool, error) {
	ctx := context.Background()
	log.Printf("Checking if '%s' already running", containerName)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, errors.Wrapf(err, "Unable to create Docker client")
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return false, errors.Wrapf(err, "Unable to list containers")
	}

	// not sure about why docker adds the prefix /
	cn := fmt.Sprintf("/%s", containerName)

	// check the list of names to see if it is there
	// assume it is running if it's present
	found := false
	for _, container := range containers {
		found = slices.Contains(container.Names, cn)
		if found {
			break
		}
	}

	return found, nil

}
