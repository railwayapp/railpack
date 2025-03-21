#!/bin/bash

if [ "$(cat /hello)" != "world" ]; then
  echo "Error: /hello file does not contain 'world'"
  exit 1
fi

echo "custom start!"
