## Tutorial: Running MIcrofab and deploying a Smart Contract

_**Time**: 4 minutes_

_**Required setup**: Make sure you've docker installed as well as curl and a nodejs runtime (min v14.x)_

---

- Create a blank directory and change to it. Also a couple of terminal windows will be useful here
- We're going to need the `peer` command from Fabric itself

```
curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh | bash -s -- binary
export PATH=$PATH:$(pwd)/bin
export FABRIC_CFG_PATH=$(pwd)/config
```

- We need to get a SmartContract to test; let's get a pre-packaged one from the test suite in Microfab's own repo
```
curl -sSL https://github.com/hyperledger-labs/microfab/raw/main/integration/data/asset-transfer-basic-typescript.tgz -o asset-transfer-basic-typescript.tgz
```

- Start Microfab with it's default configuration; (in a separate terminal run `docker logs -f microfab` so you can see what it's doing)
```
curl -sSL https://github.com/hyperledger-labs/microfab/releases/download/v0.0.19/microfab-amd64 -o microfab
chmod +x ./microfab
curl -s https://raw.githubusercontent.com/hyperledger-labs/microfab/main/examples/two-orgs.json -o config.json
./microfab start --configFile ./config.json
```

- We need to get the configuration of microfab and the address identities that it created; using the Hyperledger Labs *weft* tool is the quickest

```
microfab connect 
```

- This writes out a certificates and keys in a structure to use with the PeerCLI. Set the current shell enviroment variables for org1

```
source _mfcfg/org1.env
```

- We can then Install, Approve and Commit the chaincode definition

```
peer lifecycle chaincode install $(pwd)/asset-transfer-basic-typescript.tgz
export PACKAGE_ID=$(peer lifecycle chaincode queryinstalled --output json | jq -r '.installed_chaincodes[0].package_id')

peer lifecycle chaincode approveformyorg --orderer orderer-api.127-0-0-1.nip.io:8080 \
                                        --channelID channel1  \
                                        --name assettx  \
                                        -v 0  \
                                        --package-id $PACKAGE_ID \
                                        --sequence 1  
                                        
peer lifecycle chaincode commit --orderer orderer-api.127-0-0-1.nip.io:8080 \
                        --channelID channel1 \
                        --name assettx \
                        -v 0 \
                        --sequence 1 \
                        --waitForEvent
```

- Finally we can invoke a transaction, in this case a query on the Metadata of the contract.
```
peer chaincode query -C channel1 -n assettx -c '{"Args":["org.hyperledger.fabric:GetMetadata"]}'   
```

- To submit a tranasction that needs to commit updates to the ledger, we need to use  `peer chaincode invoke`
```
peer chaincode invoke -C channel1 -n assettx  \
                     -c '{"Args":["org.hyperledger.fabric:GetMetadata"]}' \
                       --orderer orderer-api.127-0-0-1.nip.io:8080 
```

Note for `invoke` the orderer needs to be specified. The output is also presented as escaped json.

```
peer chaincode invoke -C channel1 -n assettx  \
                     -c '{"Args":["org.hyperledger.fabric:GetMetadata"]}' \
                       --orderer orderer-api.127-0-0-1.nip.io:8080  2>&1 \
                       git stat|  sed -e 's/^.*payload://' | sed -e 's/..$//'  -e 's/^.//' -e 's/\\"/"/g' | jq
```

## Microfab CLI

The CLI is a small binary wrapper that will create the docker image (pulling the image if needed), and write out the identitiy information. 

The (original) way was to run the docker commands manually, see below for the equivalents

```
Microfab Launch Control

Usage:
  microfab [command]

microfab
  connect     Writes out connection details for use by the Peer CLI and SDKs
  ping        Pings the microfab image to see if it's running
  start       Starts the microfab image running
  stop        Stops the microfab image running

Additional Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
  -h, --help      help for microfab
  -v, --version   version for microfab

```

### Start
```
Starts the microfab image running

Usage:
  microfab start [flags]

Flags:
      --config string       Microfab config (default "{\"endorsing_organizations\":[{\"name\":\"org1\"}],\"channels\":[{\"name\":\"mychannel\",\"endorsing_organizations\":[\"org1\"]},{\"name\":\"appchannel\",\"endorsing_organizations\":[\"org1\"]}],\"capability_level\":\"V2_5\"}")
      --configFile string   Microfab config file
  -f, --force               Force restart if microfab already running
  -h, --help                help for start
  -l, --logs                Display the logs (docker logs -f microfab)
```

### Connect

```
Writes out connection details for use by the Peer CLI and SDKs

Usage:
  microfab connect [flags]

Flags:
  -f, --force        Force overwriting details directory
  -h, --help         help for connect
      --msp string   msp output directory (default "_mfcfg")
```

## Docker Command Equivalents

```
docker run -d --rm -p 8080:8080 --name microfab ghcr.io/hyperledger-labs/microfab:latest
```

```
curl -s http://console.127-0-0-1.nip.io:8080/ak/api/v1/components | npx @hyperledger-labs/weft microfab -w _wallets -p _gateways -m _msp -f
```
