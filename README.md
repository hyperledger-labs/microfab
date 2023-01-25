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

The idea is to have branches per release of Fabric.

- `fabric-2.5` is Microfab using the 2.5 LTS for example. (and this is the default branch)
- `fabric-2.4` uses thes (non-LTS) Fabric 2.4
- `beta-3.0` will start to use Fabric 3.0 when it starts to become available

## Reference

- [Configuring Microfab](./docs/ConfiguringMicrofab.md)
- [Getting Started Tutorial](./docs/Tutorial.md)
- [Connecting Clients](./docs/ConnectingClients.md)

### What Microfab can't do

- Run in production, please just don't do it. It's development and test only
- It doesn't yet support RAFT  

### Unable to connect errors

If you experience connection rejected type errors please check if your DNS is correctly able to resolve the `nip.io` addresses

```
ping server.127-0-0-1.nip.io
PING server.127-0-0-1.nip.io (127.0.0.1) 56(84) bytes of data.
64 bytes from localhost (127.0.0.1): icmp_seq=1 ttl=64 time=0.488 ms
64 bytes from localhost (127.0.0.1): icmp_seq=2 ttl=64 time=0.035 ms
```

Some DNS servcies reject the DNS rewriting that this service uses and causes problems as you might image with microfab's proxy. 
A feature we'd to add is a detailed list of all the ports that internal services are using, so you expose them directly. Help appreciated!

## License

Apache-2.0


