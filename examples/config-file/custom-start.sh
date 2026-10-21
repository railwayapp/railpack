#!/bin/bash

if [ "$(cat /hello)" != "world" ]; then
  echo "Error: /hello file does not contain 'world'"
  exit 1
fi

# /boop comes from an unrelated step and must not leak through implicit deploy outputs.
if [ -e /boop ]; then
  echo "Error: implicit deploy output copied /boop"
  exit 1
fi

echo "custom start!"
