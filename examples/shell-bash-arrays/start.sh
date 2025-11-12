#!/bin/bash
# Intentionally using a bash specific array syntax to ensure the correct shell is picked.

fruits=("apple" "banana" "cherry")

echo "Bash array test:"
for fruit in "${fruits[@]}"; do
  echo "- $fruit"
done

echo "Array length: ${#fruits[@]}"
echo "Test completed successfully"
