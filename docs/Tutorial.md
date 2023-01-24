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
docker run -d --rm -p 8080:8080 --name microfab ghcr.io/hyperledger-labs/microfab:latest
```

- We need to get the configuration of microfab and the address identities that it created; using the Hyperledger Labs *weft* tool is the quickest
```
curl -s http://console.127-0-0-1.nip.io:8080/ak/api/v1/components | npx @hyperledger-labs/weft microfab -w _wallets -p _gateways -m _msp -f
```

- This will show us some environment variables we can use to work with Fabric.  

```
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_MSPCONFIGPATH=$(pwd)/_msp/Org1/org1admin/msp
export CORE_PEER_ADDRESS=org1peer-api.127-0-0-1.nip.io:8080
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