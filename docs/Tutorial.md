## Tutorial

AIM: 

- Create a blank directory
- We're going to need the `peer` cli from Fabric itself

```
curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh | bash -s -- binary
export PATH=$PATH:$(pwd)/bin
export FABRIC_CFG_PATH=$(pwd)/config
```

- Start Microfab with it's default config and get it write it's data directly to the local file system.
```
mkdir data
docker run --rm -p 8080:8080 --name microfab ghcr.io/hyperledger-labs/microfab:latest
```

- We need to get a SmartContract to test; get a pre-packaged one from the test suite in Microfab's own repo
```
curl -sSL https://github.com/hyperledger-labs/microfab/raw/main/integration/data/asset-transfer-basic-typescript.tgz -o asset-transfer-basic-typescript.tgz
```

- Configure the local `peer` cli to get the information.

```


- 