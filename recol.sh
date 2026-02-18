#!/bin/bash
jq --argjson theme "$(recol -r --show-json)" '.theme = $theme' config.json > tmp.json && mv tmp.json config.json
