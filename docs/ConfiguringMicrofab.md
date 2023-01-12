
## Configuring microfab

Microfab can be configured by specifying the `MICROFAB_CONFIG` environment variable. For example, to start Microfab with different organizations using Docker, run the following commands:

    export MICROFAB_CONFIG='{
        "endorsing_organizations":[
            {
                "name": "SampleOrg"
            }
        ],
        "channels":[
            {
                "name": "mychannel",
                "endorsing_organizations":[
                    "SampleOrg"
                ]
            }
        ]
    }'

    docker run -p 8080:8080 -e MICROFAB_CONFIG ibmcom/ibp-microfab

The configuration is a JSON object with the following keys:

- `domain`

  The domain name to use. The domain name must be resolvable both outside and inside the container, and it must resolve to an IP address of that container (or the system hosting the container).

  Default value: `"127-0-0-1.nip.io"`

- `port`

  The port to use. The port must be accessible both outside and inside the container.

  Default value: `8080`

- `directory`

  The directory to store data in within the container.

  Default value: `"/home/microfab/data"`

- `ordering_organization`

  The ordering organization.

  Default value:

      {
        "name": "Orderer" // The name of the organization.
      }

- `endorsing_organizations`

  The list of endorsing organizations.

  Default value:

      [
        {
          "name": "Org1" // The name of the organization.
        }
      ]

- `channels`

  The list of channels.

  Default value:

      [
        {
          "name": "channel1", // The name of the channel.
          "endorsing_organizations": [ // The list of endorsing organizations that are members of the channel.
            "Org1"
          ],
          "capability_level": "V2_0" // Optional: the application capability level of the channel.
        }
      ]

- `capability_level`

  The application capability level of all channels. Can be overriden on a per-channel basis.

  Default value: `"V2_0"`

- `couchdb`

  Whether or not to use CouchDB as the world state database.

  Default value: `true`

- `certificate_authorities`

  Whether or not to create certificate authorities for all endorsing organizations.

  Default value: `true`

- `timeout`

  The time to wait for all components to start.

  Default value: `"30s"`

- `tls`

  The TLS configuration.

  Default value:

      {
        "enabled": false, // Set to true to enable TLS.
        "certificate": null, // Optional: the TLS certificate to be used.
        "private_key": null, // Optional: the TLS private key to be used.
        "ca": null // Optional: the TLS CA certificate to be used.
      }

### Examples

Configuration example for enabling TLS:

    export MICROFAB_CONFIG='{
        "port": 8443,
        "tls": {
          "enabled": true
        }
    }'

    docker run -p 8443:8443 -e MICROFAB_CONFIG ibmcom/ibp-microfab
