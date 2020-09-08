#!/usr/bin/env bash
#
# SPDX-License-Identifier: Apache-2.0
#
set -euo pipefail
if [ -n "${MICROFAB_CONFIG:-}" ]; then
    COUCHDB_ENABLED=$(echo "${MICROFAB_CONFIG}" | jq -r '. | if has("couchdb") then .couchdb else true end')
    if [ "${COUCHDB_ENABLED}" = "true" ]; then
        couchdb &
    fi
else
    couchdb &
fi
exec microfabd
