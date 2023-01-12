## Starting microfab

To start Microfab with the default configuration using Docker, run the following command:

    docker run -p 8080:8080 ibmcom/ibp-microfab

Microfab provides a REST API. This REST API provides all the information you need to connect to the Hyperledger Fabric runtime using any of the Hyperledger Fabric SDKs.

To access this information, use the following REST API:

    curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components

Connection profiles are returned with a type of `gateway`:

    curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components | jq '.[] | select(.type == "gateway")'

Identities (certificate and private key pairs) are returned with a type of `identity`:

    curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components | jq '.[] | select(.type == "identity")'
