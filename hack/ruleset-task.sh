#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/tasks -d \
'{
    "name":"Windup",
    "locator": "windup",
    "addon": "windup",
    "application": {"id": 4},
    "data": {
        "tagger":{
	    "enabled": true
        },
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
            },
	    "bundles": [
	      {"ID": 12}
	    ],
	    "identity": {"ID": 4},
	    "repository": {
              "kind": "git",
              "url": "https://github.com/konveyor/tackle-testapp.git",
              "branch": "",
              "tag": "",
              "path": "rules"
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
}' | jq -M .
