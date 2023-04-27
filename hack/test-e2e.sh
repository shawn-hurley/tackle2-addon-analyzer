#!/bin/bash

set -o errexit
set -o xtrace

HOST="${HOST:-localhost:8080}"
NAMESPACE="${NAMESPACE:-konveyor-tackle}"

fail_test() {
  APP_ID=$1
  TASK_ID=$2

  set +o errexit

  echo "######################"
  echo "Windup E2E Test Failed"
  echo "######################"

  echo "Addons"
  curl -S -s -X GET ${HOST}/addons
  echo
  echo "Windup Addon"
  curl -S -s -X GET ${HOST}/addons/windup
  echo
  echo "Applications"
  curl -S -s -X GET ${HOST}/applications
  if [ -n ${APP_ID} ] && [ "${APP_ID}" != "null" ]; then
    echo
    echo "# Pathfinder Application"
    curl -S -s -X GET ${HOST}/applications/${APP_ID}
  fi
  if [ -n ${TASK_ID} ] && [ "${TASK_ID}" != "null" ]; then
    echo
    echo "# Task details"
    curl -S -s -X GET ${HOST}/tasks/${TASK_ID} | jq

    echo
    echo "# Pod logs"

    TASK_POD_NAMESPACED_NAME=$(curl -S -s -X GET ${HOST}/tasks/${TASK_ID} | jq --raw-output .pod)
    TASK_POD_NAMESPACE="${TASK_POD_NAMESPACED_NAME%/*}"
    TASK_POD_NAME="${TASK_POD_NAMESPACED_NAME#*/}"

    kubectl logs --namespace ${TASK_POD_NAMESPACE} ${TASK_POD_NAME}
  fi

  echo "Every deployment in ${NAMESPACE}"
  kubectl describe --namespace ${NAMESPACE} deployments

  echo "Logs from the hub"
  kubectl logs --namespace ${NAMESPACE} deployments/tackle-hub

  exit 2
}

# Exit early if kubectl or jq not installed
if ! command -v kubectl >/dev/null 2>&1; then
  echo "Please install kubectl"
  exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "Please install jq"
  exit 1
fi

# Verify we can talk with hub first
if ! (timeout 300s bash -c "until curl -f -S -s -o /dev/null -X GET ${HOST}/addons/windup; do sleep 10; done" || false); then
  echo "Windup addon not found. Is the hub running?"
  fail_test "" ""
fi
echo "Verified windup addon installed."

# Give tackle a minute to give us an application list
# Having issues in testing where it could return 503
if ! (timeout 300s bash -c "until curl -f -S -s -X GET ${HOST}/applications; do sleep 10; done" || false); then
  echo "Tackle won't give a list of applications"
  fail_test "" ""
fi
# At this point, if tackle isn't accepting requests...that's a problem.

# Create pathfinder app if it hasn't been added already
# There is a constraint that only allows one application to have a particular name.
if ! curl -S -s -X GET ${HOST}/applications | jq -e 'any(.[]; .name == "Pathfinder")'; then
  echo "Creating pathfinder application"
  curl -X POST ${HOST}/applications -d \
    '{
        "name":"Pathfinder",
        "description": "Tackle Pathfinder application.",
        "repository": {
          "name": "tackle-pathfinder",
          "url": "https://github.com/konveyor/tackle-pathfinder.git",
          "branch": "1.2.0" }
    }' | jq -M .
fi
APP_ID=$(curl -S -s -X GET ${HOST}/applications | jq --raw-output '.[] | select(.name=="Pathfinder") | .id')
echo "Pathfinder exists with id ${APP_ID}"
# Show the applications in the inventory
curl -S -s -X GET ${HOST}/applications | jq

# Make a request to hub
TASK_ID=$(curl -S -s -X POST ${HOST}/tasks -d \
'{
    "name":"Windup",
    "state": "Ready",
    "locator": "windup",
    "addon": "windup",
    "application": {"id": '$APP_ID'},
    "data": {
        "mode": {
            "artifact": "",
            "binary": false,
            "withDeps": false,
	    "diva": true
        },
        "output": "/windup/report",
        "rules": {
            "path": "",
            "tags": {
                "excluded": [ ]
            }
        },
        "scope": {
            "packages": {
                "excluded": [ ],
                "included": [ ]
            },
            "withKnown": false
        },
        "sources": [ ],
        "targets": [
            "cloud-readiness"
        ]
    }
}' | jq .id)
if [ "${TASK_ID}" = "null" ]; then
  fail_test ${APP_ID} ${TASK_ID}
fi
echo "Task created with id ${TASK_ID}"

# Give windup ten minutes to finish
if ! (timeout 300s bash -c "until curl -S -s -X GET ${HOST}/tasks/${TASK_ID} | jq -e '.state == \"Succeeded\"'; do sleep 30; done" || false); then
  fail_test ${APP_ID} ${TASK_ID}
fi
echo "Windup task completed successfully"
