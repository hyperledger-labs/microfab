/*
 * SPDX-License-Identifier: Apache-2.0
 */

package fablet

import (
	"encoding/json"
	"os"
	"path"
)

// Organization represents an organization in the configuration.
type Organization struct {
	Name string `json:"name"`
}

// Channel represents a channel in the configuration.
type Channel struct {
	Name                   string   `json:"name"`
	EndorsingOrganizations []string `json:"endorsing_organizations"`
}

// Config represents the configuration.
type Config struct {
	Domain                 string         `json:"domain"`
	Port                   int            `json:"port"`
	Directory              string         `json:"directory"`
	OrderingOrganization   Organization   `json:"ordering_organization"`
	EndorsingOrganizations []Organization `json:"endorsing_organizations"`
	Channels               []Channel      `json:"channels"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() (*Config, error) {
	home, ok := os.LookupEnv("FABLET_HOME")
	if !ok {
		var err error
		home, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	config := &Config{
		Domain:    "127-0-0-1.nip.io",
		Port:      8080,
		Directory: path.Join(home, "data"),
		OrderingOrganization: Organization{
			Name: "Orderer",
		},
		EndorsingOrganizations: []Organization{
			{
				Name: "Org1",
			},
		},
		Channels: []Channel{
			{
				Name: "channel1",
				EndorsingOrganizations: []string{
					"Org1",
				},
			},
		},
	}
	if env, ok := os.LookupEnv("FABLET_CONFIG"); ok {
		err := json.Unmarshal([]byte(env), config)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}
