#!/bin/bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -xeuo pipefail

# Grab the current directory and make the cfg directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/.. && pwd )"
export CFG=$DIR/_cfg
rm -rf "${CFG}" || mkdir -p "${CFG}"
mkdir -p "${CFG}/data"

: ${MICROFAB_IMAGE:="ghcr.io/hyperledger-labs/microfab:latest"}

if docker inspect microfab &>/dev/null; then
    echo "Removing existing microfab container:"
    docker kill microfab
fi

export MICROFAB_CONFIG='{
        "couchdb":false,
        "endorsing_organizations":[
            {
                "name": "org1"
            }, 
            {
                "name":"org2"
            }

        ],
        "channels":[
            {
                "name": "ch-a",
                "endorsing_organizations":[
                    "org1","org2"
                ]
            }
        ],
        "tls": {
            "enabled":true
        },
        "capability_level":"V2_5"
    }'

# docker run  --name microfab -u $(id -u) -p 8080:8080 --add-host host.docker.internal:host-gateway \
#             --rm -e MICROFAB_CONFIG="${MICROFAB_CONFIG}" \
#             -e FABRIC_LOGGING_SPEC=info \
#             -v "${CFG}/data":/home/microfab/data \
#             ${MICROFAB_IMAGE}
docker run -d --name microfab  -p 8080:8080 --add-host host.docker.internal:host-gateway \
            --rm -e MICROFAB_CONFIG="${MICROFAB_CONFIG}" \
            -e FABRIC_LOGGING_SPEC=info \
            ${MICROFAB_IMAGE}

# Get the configuration and extract the information
sleep 25

curl -sSL --insecure https://console.127-0-0-1.nip.io:8080/ak/api/v1/components
curl -sSL --insecure https://console.127-0-0-1.nip.io:8080/ak/api/v1/components | npx @hyperledger-labs/weft microfab -w $CFG/_wallets -p $CFG/_gateways -m $CFG/_msp -f

# Chaincodes are all ready packaged up in the integration directory

# set for peer 1
export CORE_PEER_TLS_ENABLED=true 
export CORE_PEER_LOCALMSPID=org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE="${CFG}/_msp/tls/org1peer/tlsca-org1peer-cert.pem"
export CORE_PEER_MSPCONFIGPATH="${CFG}/_msp/org1/org1admin/msp"
export CORE_PEER_ADDRESS=org1peer-api.127-0-0-1.nip.io:8080
export ORDERER_CA="$CFG/_msp/tls/orderer/tlsca-orderer-cert.pem"

peer lifecycle chaincode install ${DIR}/integration/data/asset-transfer-basic-go.tgz
export PACKAGE_ID=$(peer lifecycle chaincode queryinstalled --output json | jq -r '.installed_chaincodes[0].package_id')
echo $PACKAGE_ID
peer lifecycle chaincode approveformyorg  --orderer orderer-api.127-0-0-1.nip.io:8080  \
                                        --channelID ch-a  \
                                        --name basic-go  \
                                        -v 0  \
                                        --package-id $PACKAGE_ID \
                                        --sequence 1  \
                                        --tls  \
                                        --cafile $ORDERER_CA

peer lifecycle chaincode checkcommitreadiness --channelID ch-a --name basic-go -v 0 --sequence 1


# set for peer 2

export CORE_PEER_TLS_ENABLED=true 
export CORE_PEER_LOCALMSPID=org2MSP
export CORE_PEER_TLS_ROOTCERT_FILE="${CFG}/_msp/tls/org2peer/tlsca-org2peer-cert.pem"
export CORE_PEER_MSPCONFIGPATH="${CFG}/_msp/org2/org2admin/msp"
export CORE_PEER_ADDRESS=org2peer-api.127-0-0-1.nip.io:8080
export ORDERER_CA="$CFG/_msp/tls/orderer/tlsca-orderer-cert.pem"

peer lifecycle chaincode install ${DIR}/integration/data/asset-transfer-basic-go.tgz
export PACKAGE_ID=$(peer lifecycle chaincode queryinstalled --output json | jq -r '.installed_chaincodes[0].package_id')
echo $PACKAGE_ID
peer lifecycle chaincode approveformyorg  --orderer orderer-api.127-0-0-1.nip.io:8080  \
                                        --channelID ch-a  \
                                        --name basic-go  \
                                        -v 0  \
                                        --package-id $PACKAGE_ID \
                                        --sequence 1  \
                                        --tls  \
                                        --cafile $ORDERER_CA

peer lifecycle chaincode checkcommitreadiness --channelID ch-a --name basic-go -v 0 --sequence 1

# commit as either org
peer lifecycle chaincode commit --orderer orderer-api.127-0-0-1.nip.io:8080   \
                        --peerAddresses org1peer-api.127-0-0-1.nip.io:8080 --tlsRootCertFiles "${CFG}/_msp/tls/org1peer/tlsca-org1peer-cert.pem" \
                        --peerAddresses org2peer-api.127-0-0-1.nip.io:8080 --tlsRootCertFiles "${CFG}/_msp/tls/org2peer/tlsca-org2peer-cert.pem" \
                        --channelID ch-a --name basic-go -v 0 \
                        --sequence 1 \
                        --tls --cafile $ORDERER_CA --waitForEvent
 
peer chaincode query -C ch-a -n basic-go -c '{"Args":["org.hyperledger.fabric:GetMetadata"]}'                        