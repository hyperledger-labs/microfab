#!/bin/bash

source /working/_utils/utils.sh

CHANNEL_NAME="channel1"
CHAINCODE_PACKAGE=${1}
MAX_RETRY="5"
DELAY="3"
CC_SEQUENCE="auto"

println "executing with the following"
println "- CHANNEL_NAME: ${C_GREEN}${CHANNEL_NAME}${C_RESET}"
println "- CHAINCODE_PACKAGE: ${C_GREEN}${CHAINCODE_PACKAGE}${C_RESET}"

. /working/_utils/ccutils.sh

function checkPrereqs() {
  jq --version > /dev/null 2>&1

  if [[ $? -ne 0 ]]; then
    errorln "jq command not found..."
    errorln
    errorln "Follow the instructions in the Fabric docs to install the prereqs"
    errorln "https://hyperledger-fabric.readthedocs.io/en/latest/prereqs.html"
    exit 1
  fi
}

#check for prerequisites
checkPrereqs

PACKAGE_ID=$(peer lifecycle chaincode calculatepackageid /working/_packaged/${CHAINCODE_PACKAGE})

infoln "Chaincode package ID: ${PACKAGE_ID}"

export CC_NAME=${CHAINCODE_PACKAGE%_*}
export CC_VERSION=$(echo ${CHAINCODE_PACKAGE} | sed 's/.tar.gz//g' | sed 's/.*_//')

infoln "Chaincode Name: ${CC_NAME}"
infoln "Chaincode Version: ${CC_VERSION}"

## Install chaincode on peer0.org1 and peer0.org2
infoln "Installing chaincode on org1peer..."
installChaincode

## Set the sequence
resolveSequence

infoln "Using sequence ${CC_SEQUENCE}"

## query whether the chaincode is installed
queryInstalled

## approve the definition for org1
approveForMyOrg

## check whether the chaincode definition is ready to be committed
## expect org1 to have approved and org2 not to
checkCommitReadiness "\"Org1MSP\": true"

## now that we know for sure both orgs have approved, commit the definition
commitChaincodeDefinition 

## query on both orgs to see that the definition committed successfully
queryCommitted

exit 0
