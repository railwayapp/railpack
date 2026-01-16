#!/bin/bash
echo "Running start.sh"
[ -f artifacts/install.txt ] && cat artifacts/install.txt
[ -f artifacts/build.txt ] && cat artifacts/build.txt