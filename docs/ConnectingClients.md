# Connecting Clients

## Fabric Gateway Clients

As an example look at the the [`trader-typescript` example](https://github.com/hyperledger/fabric-samples/tree/main/full-stack-asset-transfer-guide/applications/trader-typescript). Specifically the [`config.ts` file](https://github.com/hyperledger/fabric-samples/blob/main/full-stack-asset-transfer-guide/applications/trader-typescript/src/config.ts).  

Though is is typescript the information that is needed is the same irrespective of language. Any typical application will need to know the following information. In the example this is using envronment variables

|                |                                                                  |
|----------------|------------------------------------------------------------------|
| ENDPOINT       | Endpoint address of the gateway service                          |
| MSP_ID         | User's organization Member Services Provider ID                  |
| CERTIFICATE    | User's certificate file                                          |
| PRIVATE_KEY    | User's private key file                                          |
| CHANNEL_NAME   | Channel to which the chaincode is deployed                       |
| CHAINCODE_NAME | Channel to which the chaincode is deployed                       |
| TLS_CERT       | TLS CA root certificate (only if using TLS and private CA)       |
| HOST_ALIAS     | TLS hostname override (only if TLS cert does not match endpoint) |

In the [tutorial](./Tutorial.md) Microfab was started as follows

```
docker run -d --rm -p 8080:8080 --name microfab ghcr.io/hyperledger-labs/microfab:latest
curl -s http://console.127-0-0-1.nip.io:8080/ak/api/v1/components | npx @hyperledger-labs/weft microfab -w _wallets -p _gateways -m _msp -f
```

To setup the environment for the Peer Command (nonTLS):
```
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_MSPCONFIGPATH=$(pwd)/_msp/Org1/org1admin/msp
export CORE_PEER_ADDRESS=org1peer-api.127-0-0-1.nip.io:8080
```

To setup the environment for the example gateway application (nonTLS)

```
export ENDPOINT=org1peer-api.127-0-0-1.nip.io:8080
export MSP_ID=Org1MSP
export CERTIFICATE=$(pwd)/_msp/org1/org1admin/msp/admincerts/cert.pem
export PRIVATE_KEY=$(pwd)/_msp/org1/org1admin/msp/keystore/cert_sk
```

