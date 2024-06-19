#!/bin/bash

C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_YELLOW='\033[1;33m'

function installPrereqs() {

  infoln "installing prereqs"

  FILE=../install-fabric.sh     
  if [ ! -f $FILE ]; then
    curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh && chmod +x install-fabric.sh
    cp install-fabric.sh ..
  fi
  
  IMAGE_PARAMETER=""
  if [ "$IMAGETAG" != "default" ]; then
    IMAGE_PARAMETER="-f ${IMAGETAG}"
  fi 

  CA_IMAGE_PARAMETER=""
  if [ "$CA_IMAGETAG" != "default" ]; then
    CA_IMAGE_PARAMETER="-c ${CA_IMAGETAG}"
  fi 

  cd ..
  ./install-fabric.sh ${IMAGE_PARAMETER} ${CA_IMAGE_PARAMETER} docker binary

}

# println echos string
function println() {
  echo -e "$1"
}

# errorln echos i red color
function errorln() {
  println "${C_RED}${1}${C_RESET}"
}

# successln echos in green color
function successln() {
  println "${C_GREEN}${1}${C_RESET}"
}

# infoln echos in blue color
function infoln() {
  println "${C_BLUE}${1}${C_RESET}"
}

# warnln echos in yellow color
function warnln() {
  println "${C_YELLOW}${1}${C_RESET}"
}

# fatalln echos in red color and exits with fail status
function fatalln() {
  errorln "$1"
  exit 1
}

export -f errorln
export -f successln
export -f infoln
export -f warnln
