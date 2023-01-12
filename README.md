# microfab

'microfab' provides a single container image that allows you to quickly start Hyperledger Fabric when you are developing solutions. You can use it to rapidly iterate over changes to chaincode, and client applications.

This containerized version of Fabric can be easily configured with the selection of channels and orgs you want, and also can be started and stopped in seconds.  You can interfact with it as you would any Fabric setup. Note that this uses *the* fabric binaries and starts Fabric with couchdb and cas for identities. It's not cut down.

[![asciicast](https://asciinema.org/a/519913.svg)](https://asciinema.org/a/519913)

## Why microfab?

There are other 'form factors' of Fabric some are aimed at production/k8s deployments. Of more development focussed form factors some key ones are.

- test-network with Fabric Samples - a docker-compose approach to starting fabric great for running the samples and as the 'reference standard'
- minifab - also a good way of standing up a overall fabric network
- test-network-nano - based around the separate binaries, useful when developing Fabric itself.

Try several out, and see which you prefer and suits your way of working best. 

## Quick Start

To start Microfab with the default configuration using Docker, run the following command:

    docker run -p 8080:8080 ibmcom/ibp-microfab

To access this information, use the following REST API (from another terminal):

    curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components


## Use for debugging Smart Contracts

To learn how to use Microfab as part of the development workflow, follow the smart contract part of the [Hyperledger Fabric Sample's Full Stack Tutorial](https://github.com/hyperledger/fabric-samples/blob/main/full-stack-asset-transfer-guide/docs/SmartContractDev/00-Introduction.md)

## Documentation

- [Starting Microfab]
- [Configruing Microfab]()
- [Examples](./examples/README.md)

### What Microfab can't do

- Run in production, please just don't do it. It's development and test only
- It supports TLS 
- It doesn't yet support RAFT

## What Fabric version does Microfab use?

The main branch supports the current LTS release of Fabric; the main docker image and binaries published are for the current Fabric LTS.

There is (will be) an beta branch that will target the development stream of Fabric - that may or may not work of course.

## License

Apache-2.0


