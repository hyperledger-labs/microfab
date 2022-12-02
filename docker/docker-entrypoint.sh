#!/usr/bin/env bash
#
# SPDX-License-Identifier: Apache-2.0
#
set -euo pipefail

# If this image is run with -u <random UID>, as happens on Red Hat OpenShift, then
# the user is not in the /etc/passwd file. This causes Ansible to fail, so we need
# to add the user to /etc/passwd now before Ansible runs.
if ! whoami &> /dev/null; then
    sed '/microfab/d' /etc/passwd > /tmp/passwd
    cat /tmp/passwd > /etc/passwd
    rm -f /tmp/passwd
    echo "microfab:x:$(id -u):0::/home/microfab:/bin/bash" >> /etc/passwd
    export HOME=/home/microfab
fi


if [ -n "${MICROFAB_CONFIG:-}" ]; then
    COUCHDB_ENABLED=$(echo "${MICROFAB_CONFIG}" | jq -r '. | if has("couchdb") then .couchdb else true end')
    if [ "${COUCHDB_ENABLED}" = "true" ]; then
        couchdb &
    fi
else
    couchdb &
fi
exec microfabd
