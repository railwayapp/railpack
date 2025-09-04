import sys
import subprocess

print("Hello from Python UV!")
print(f"Python version: {sys.version.split()[0]}")

try:
    result = subprocess.run(["uv", "--version"], capture_output=True, text=True, check=True)
    print(f"UV version: {result.stdout.strip().split()[1]}")
except (subprocess.CalledProcessError, FileNotFoundError):
    print("UV version: not available")
