#!/bin/bash
# Copyright 2016 The LUCI Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# Resolved "http://metadata.google.internal", b/c the DNS resolution can flake.
METADATA_URL="http://169.254.169.254"

_help() {
  echo -e "Usage: $0 <logdog-coordinator-path>"
  echo -e ""
  echo -e "  logdog-coordinator-path\tThe path to the application binary."
}

APP=$1; shift
if [ -z "${APP}" ]; then
  echo "ERROR: Missing argument <logdog-coordinator-path>"
  _help
  exit 1
fi
APP=$(readlink -f "${APP}")
if [ ! -x "${APP}" ]; then
  echo "ERROR: Application path is not an executable file [${APP}]"
  exit 1
fi

_load_metadata_domain() {
  local __resultvar=$1; shift
  local domain=$1; shift
  local resource=$1; shift

  OUTPUT=$(curl \
    -s \
    --fail \
    "${METADATA_URL}/computeMetadata/v1/${domain}/${resource}" \
    -H "Metadata-Flavor: Google")
  local RV=$?
  if [ ${RV} != 0 ]; then
    return ${RV}
  fi
  if [ -z "${OUTPUT}" ]; then
    return 1
  fi

  eval $__resultvar="'${OUTPUT}'"
}

# Loads metadata from the GCE metadata server.
_load_metadata() {
  local __resultvar=$1; shift
  local resource=$1; shift

  # Try loading from 'instance' domain.
  _load_metadata_domain "$__resultvar" "instance" "$resource"
  local RV=$?
  if [ ${RV} == 0 ]; then
    return 0
  fi

  # Try loading from 'project' domain.
  _load_metadata_domain "$__resultvar" "project" "$resource"
  RV=$?
  if [ ${RV} == 0 ]; then
    return 0
  fi

  echo "WARN: Failed to load metadata [${resource}]."
  return 1
}

_load_metadata_check() {
  local __resultvar=$1; shift
  local resource=$1; shift

  _load_metadata "$__resultvar" "$resource"
  local RV=$?
  if [ ${RV} != 0 ]; then
    echo "ERROR: Metadata resource [${resource}] is required."
    exit 1
  fi

  return 0
}

# Test if we're running on a GCE instance.
echo "Testing for GCE instance: ${METADATA_URL}"
curl \
  -s \
  --fail \
  "${METADATA_URL}" \
  1>/dev/null
RV=$?
if [ ${RV} != 0 ]; then
  echo "ERROR: Not running on GCE instance (${RV})."
  exit 1
fi

# Load metadata (common).
_load_metadata_check PROJECT_ID "project-id"
_load_metadata_check COORDINATOR_HOST "attributes/logdog_coordinator_host"

# Load metadata (ts-mon).
_load_metadata TSMON_ENDPOINT "attributes/tsmon_endpoint"
_load_metadata TSMON_ACT_AS "attributes/tsmon_act_as"

# Load metadata (app-specific).
_load_metadata LOG_LEVEL "attributes/logdog_archivist_log_level"

# Runtime temporary directory.
TEMPDIR=$(mktemp -d)
trap "rm -rf ${TEMPDIR}" EXIT

# Compose command line.
ARGS=(
  "-coordinator" "${COORDINATOR_HOST}"
  "-config-kill-interval" "5m"
  "-service-account-json" ":gce"
  )

if [ ! -z "${LOG_LEVEL}" ]; then
  ARGS+=("-log-level" "${LOG_LEVEL}")
fi
if [ ! -z "${TSMON_ENDPOINT}" ]; then
  ARGS+=( \
    "-ts-mon-endpoint" "${TSMON_ENDPOINT}" \
    "-ts-mon-act-as" "${TSMON_ACT_AS}" \
    "-ts-mon-credentials" ":gce" \
    "-ts-mon-autogen-hostname" \
    )
fi

echo "INFO: Running command line args: ${APP} ${ARGS[*]}"
"${APP}" ${ARGS[*]}
