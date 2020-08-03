#!/bin/bash

#
# SPDX-License-Identifier: Apache-2.0
#

#######################################
# Exit script with an error
# Globals:
#   None
# Arguments:
#   message: Optional error message
# Returns:
#   None
#######################################
function error_exit {
    echo "${1:-"Unknown Error"}" 1>&2
    exit 1
}

#######################################
# Check whether the chaincode type is supported
# Globals:
#   None
# Arguments:
#   CHAINCODE_METADATA_DIR: Location of metadata.json
#   CHAINCODE_TYPE: The required chaincode type
# Returns:
#   None
#######################################
function check_chaincode_type {
    local CHAINCODE_METADATA_DIR="$1"
    local CHAINCODE_TYPE="$2"

    if [ ! "$(jq -r .type "$CHAINCODE_METADATA_DIR/metadata.json" | tr '[:upper:]' '[:lower:]')" = "$CHAINCODE_TYPE" ]; then
        error_exit "Unsupported chaincode type: $CHAINCODE_TYPE"
    fi
}

#######################################
# Copy the two types of metadata the peer consumes from chaincode source, i.e.
#   1. state database index definitions for CouchDB
#   2. external chaincode server connection information
# Globals:
#   None
# Arguments:
#   FROM_DIR: Location of the metadata
#   TO_DIR: Destination directory
# Returns:
#   None
#######################################
function copy_build_metadata_artifacts {
    local FROM_DIR="$1"
    local TO_DIR="$2"

    # CouchDB index definitions
    if [ -d "$FROM_DIR/metadata" ]; then
        cp -a "$FROM_DIR/metadata" "$TO_DIR/metadata"
    fi

    # Server connection information
    if [ -f "$FROM_DIR/connection.json" ]; then
        cp "$FROM_DIR/connection.json" "$TO_DIR/connection.json"
    fi
}

#######################################
# Copy the two types of metadata the peer consumes to the required release layout, i.e.
#   1. state database index definitions for CouchDB
#   2. external chaincode server connection information
# Globals:
#   None
# Arguments:
#   FROM_DIR: Location of the metadata
#   TO_DIR: Destination directory
# Returns:
#   None
#######################################
function copy_release_metadata_artifacts {
    local FROM_DIR="$1"
    local TO_DIR="$2"

    # CouchDB index definitions
    if [ -d "$FROM_DIR/metadata" ]; then
        cp -a "$FROM_DIR/metadata/"* "$TO_DIR/"
    fi

    # Server connection information
    if [ -f "$FROM_DIR/connection.json" ]; then
        mkdir -p "$TO_DIR/chaincode/server"
        cp "$FROM_DIR/connection.json" "$TO_DIR/chaincode/server"
    fi
}

#######################################
# Export environment variables and extract certificate files from chaincode.json
# Globals:
#   None
# Arguments:
#   METADATA_DIR: Location of the chaincode.json file
# Returns:
#   None
#######################################
function process_chaincode_metadata_json {
    local METADATA_DIR="$1"

    CORE_CHAINCODE_ID_NAME="$(jq -r .chaincode_id "$METADATA_DIR/chaincode.json")"
    CORE_PEER_ADDRESS="$(jq -r .peer_address "$METADATA_DIR/chaincode.json")"
    CORE_PEER_LOCALMSPID="$(jq -r .mspid "$METADATA_DIR/chaincode.json")"
    export CORE_CHAINCODE_ID_NAME
    export CORE_PEER_ADDRESS
    export CORE_PEER_LOCALMSPID

    if [ -z "$(jq -r .client_cert "$METADATA_DIR/chaincode.json")" ]; then
        CORE_PEER_TLS_ENABLED="false"
        export CORE_PEER_TLS_ENABLED
    else
        CORE_PEER_TLS_ENABLED="true"
        CORE_TLS_CLIENT_CERT_FILE="$PWD/client.crt"
        CORE_TLS_CLIENT_KEY_FILE="$PWD/client.key"
        CORE_PEER_TLS_ROOTCERT_FILE="$PWD/root.crt"
        export CORE_PEER_TLS_ENABLED
        export CORE_TLS_CLIENT_CERT_FILE
        export CORE_TLS_CLIENT_KEY_FILE
        export CORE_PEER_TLS_ROOTCERT_FILE

        jq -r .client_cert "$METADATA_DIR/chaincode.json" > "$CORE_TLS_CLIENT_CERT_FILE"
        jq -r .client_key  "$METADATA_DIR/chaincode.json" > "$CORE_TLS_CLIENT_KEY_FILE"
        jq -r .root_cert   "$METADATA_DIR/chaincode.json" > "$CORE_PEER_TLS_ROOTCERT_FILE"
    fi
}
