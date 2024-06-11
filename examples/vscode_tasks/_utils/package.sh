#!/bin/bash

CHAINCODE_NAME=${1}
CHAINCODE_VERSION=${2}
CHAINCODE_LANGUAGE=${3}
PACKAGE_TYPE=${4}
CHAINCODE_PATH=${5}

NOTPASSED=""

if [ -z "${CHAINCODE_NAME}" ];then
	NOTPASSED="${NOTPASSED} CHAINCODE_NAME"
fi

if [ -z "${CHAINCODE_VERSION}" ];then
	NOTPASSED="${NOTPASSED} CHAINCODE_VERSION"
fi

if [ -z "${CHAINCODE_LANGUAGE}" ];then
	NOTPASSED="${NOTPASSED} CHAINCODE_LANGUAGE"
fi

if [ -z "${PACKAGE_TYPE}" ];then
	NOTPASSED="${NOTPASSED} PACKAGE_TYPE"
fi

if [ -z "${CHAINCODE_PATH}" ];then
	NOTPASSED="${NOTPASSED} CHAINCODE_PATH"
fi

if [ ! -z "${NOTPASSED}" ]; then
	echo "${NOTPASSED} not passed"
  echo ""
	echo "Usage: ./package.sh <CHAINCODE_NAME> <CHAINCODE_VERSION> <CHAINCODE_LANGUAGE> <PACKAGE_TYPE> <CHAINCODE_PATH>"
  echo ""
  echo "CHAINCODE_NAME     - chaincode name for the packaged chaincode"
  echo "CHAINCODE_VERSION  - chaincode version for the packaged chaincode"
  echo "CHAINCODE_LANGUAGE - chaincode language (golang, node or java)"
  echo "PACKAGE_TYPE       - chaincode packaging type (cds or tar)"
  echo "CHAINCODE_PATH     - path to the chaincode source (fully qualfied path)"
  echo ""
  echo "Examples:"
  echo ""
  echo "Pacakging for cds file:"
  echo "./package.sh marbles02 1.1.2 node cds /Users/dev/fabric/chaincode/chaincode/marbles02/javascript"
  echo ""
  echo "Pacakging for tar file:"
  echo "./package.sh marbles02 3.0.0 golang tar /Users/dev/support/chaincode/marbles02/go"
  echo ""
	exit 1
fi

if ! [[ ${CHAINCODE_LANGUAGE} == "node" || ${CHAINCODE_LANGUAGE} == "golang" || ${CHAINCODE_LANGUAGE} == "java" ]]; then
	echo "Chaincode language must be node, golang or java"
	exit 1
fi

if ! [[ ${PACKAGE_TYPE} == "cds" || ${PACKAGE_TYPE} == "tar" ]]; then
	echo "Package type must be cds or tar"
	exit 1
fi

export FABRIC_CFG_PATH=/etc/hyperledger/fabric
export CORE_PEER_MSPCONFIGPATH=/working/_msp/Org1/org1admin/msp

if [[ ("${PACKAGE_TYPE}" == "cds") ]]; then
  echo "Building chaincode package ${CHAINCODE_NAME}@${CHAINCODE_VERSION}.cds"
  peer chaincode package /working/_packaged/${CHAINCODE_NAME}@${CHAINCODE_VERSION}.cds --path ${CHAINCODE_PATH} --lang ${CHAINCODE_LANGUAGE} -n ${CHAINCODE_NAME} -v ${CHAINCODE_VERSION}
else
  echo "Building chaincode package ${CHAINCODE_NAME}_${CHAINCODE_VERSION}.tar"
  peer lifecycle chaincode package /working/_packaged/${CHAINCODE_NAME}_${CHAINCODE_VERSION}.tar.gz --path ${CHAINCODE_PATH} --lang  ${CHAINCODE_LANGUAGE} --label ${CHAINCODE_NAME}_${CHAINCODE_VERSION}
fi

echo "Complete"

