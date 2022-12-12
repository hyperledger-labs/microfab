#!/bin/bash

set -xeu -o pipefail
DIR=-"$(cd "$(dirname "${BASH_SOURCE[0]}" )"/.. && pwd )"

# Register and Enroll a New user

# expect path is correct setup

            `export FABRIC_CA_CLIENT_HOME=${dir}/test-network/organizations/peerOrganizations/org1.example.com/ && ` +
            `fabric-ca-client register --caname ca-org1 --id.name owner --id.secret ownerpw --id.type client --tls.certfiles "${dir}/test-network/organizations/fabric-ca/org1/tls-cert.pem" && ` +
            `fabric-ca-client enroll -u https://owner:ownerpw@localhost:7054 --caname ca-org1 -M "${dir}/test-network/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp" --tls.certfiles "${dir}/test-network/organizations/fabric-ca/org1/tls-cert.pem" && ` +
            `cp "${dir}/test-network/organizations/peerOrganizations/org1.example.com/msp/config.yaml" "${dir}/test-network/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp/config.yaml" && ` +
            `export FABRIC_CA_CLIENT_HOME=${dir}/test-network/organizations/peerOrganizations/org2.example.com/ && ` +
            `fabric-ca-client register --caname ca-org2 --id.name buyer --id.secret buyerpw --id.type client --tls.certfiles "${dir}/test-network/organizations/fabric-ca/org2/tls-cert.pem" && ` +
            `fabric-ca-client enroll -u https://buyer:buyerpw@localhost:8054 --caname ca-org2 -M "${dir}/test-network/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp" --tls.certfiles "${dir}/test-network/organizations/fabric-ca/org2/tls-cert.pem" && ` +
            `cp "${dir}/test-network/organizations/peerOrganizations/org2.example.com/msp/config.yaml" "${dir}/test-network/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp/config.yaml"`;
    execSync(cmd);
