#!/bin/bash

set -e

if [ $# -eq 0 ]; then
    theme_json=$(recol -r --show-json)
else
    theme_json=$(recol "$@" --show-json)
fi

jq --indent 4 --argjson theme "$theme_json" '.theme = $theme' config.json > tmp.json
mv tmp.json config.json
