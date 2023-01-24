# microfab

'microfab' provides a single container image that allows you to quickly start Hyperledger Fabric when you are developing solutions. You can use it to rapidly iterate over changes to chaincode, and client applications. Configured with your selection of channels and orgs you want, it can be started and stopped in seconds.  

To learn how to use Microfab as part of the development workflow, follow the smart contract part of the [Hyperledger Fabric Sample's Full Stack Tutorial](https://github.com/hyperledger/fabric-samples/blob/main/full-stack-asset-transfer-guide/docs/SmartContractDev/00-Introduction.md)

Check the [reference](./docs/DevelopingContracts.md) in this repo for details in other langauges.

[![asciicast](https://asciinema.org/a/519913.svg)](https://asciinema.org/a/519913)


## Tutorial

Check the [Quick Start Tutorial](./docs/Tutorial.md) - nothing to deployed smart contract in under 5minutes;
## Why microfab?

There are other 'form factors' of Fabric some are aimed at production/k8s deployments others more development focussed.

- test-network with Fabric Samples - a docker-compose approach to starting fabric great for running the samples and as the 'reference standard'
- minifab - also a good way of standing up a overall fabric network
- test-network-nano - based around the separate binaries, useful when developing Fabric itself.

Depending on your circumstances, familiarity and requirements different tools may be better. After running with Microfab in the Full Stack AssetTransfer workshop - Microfab is particularly good for the earlier contract and SDK development phases.

## What Fabric version does Microfab use?

The main branch supports the current LTS release of Fabric; the main docker image and binaries published are for the current Fabric LTS.

There is (will be) an beta branch that will target the development stream of Fabric - that may or may not work of course.


## Reference

- [Configuring Microfab](./docs/ConfiguringMicrofab.md)
- [Getting Started Tutorial](./docs/Tutorial.md)
- [Connecting Clients](./docs/ConnectingClients.md)
### What Microfab can't do

- Run in production, please just don't do it. It's development and test only
- It doesn't yet support RAFT  

## License

Apache-2.0


