#!/bin/bash

# Output the container architecture information
echo "=== Container Architecture Information ==="
echo "Architecture: $(uname -m)"
echo "OS: $(uname -s)"
echo "Platform: $(uname -s)/$(uname -m)"

# Also check /proc/version for more details
if [ -f /proc/version ]; then
    echo "Kernel: $(cat /proc/version)"
fi

# Check if we can determine the specific architecture variant
case "$(uname -m)" in
    "x86_64")
        echo "Architecture Type: AMD64"
        ;;
    "aarch64")
        echo "Architecture Type: ARM64"
        ;;
    "armv7l")
        echo "Architecture Type: ARMv7"
        ;;
    *)
        echo "Architecture Type: Unknown ($(uname -m))"
        ;;
esac

echo "=== End Architecture Information ==="