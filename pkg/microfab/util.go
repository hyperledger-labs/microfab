package microfab

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
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

	images, err := cli.ImageList(ctx, image.ListOptions{})
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
		out, err := cli.ImagePull(ctx, microFabImage, image.PullOptions{})
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

	containers, err := cli.ContainerList(ctx, container.ListOptions{})
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

// GetConfig resolves the microfab configuration
func GetConfig() (string, error) {
	cfg = viper.GetString("MICROFAB_CONFIG")
	if cfg == "" {
		cf := path.Clean(cfgFile)
		exist, err := Exists(cf)
		if err != nil {
			return "", err
		}

		if exist {
			cfgData, err := os.ReadFile(cf)
			if err != nil {
				return "", err
			}
			cfg = string(cfgData)
		} else {
			return "", errors.Errorf("Unable to locate config from file, envvar or cli option")
		}

	}
	return cfg, nil
}

// Exists returns whether the given file or directory exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
