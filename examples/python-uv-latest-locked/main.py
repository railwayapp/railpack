import subprocess

print("Hello from Python UV latest locked!")

python_version = subprocess.run(
    ["python", "--version"],
    capture_output=True,
    text=True,
    check=True,
)
print(python_version.stdout.strip() or python_version.stderr.strip())

uv_version = subprocess.run(
    ["uv", "--version"],
    capture_output=True,
    text=True,
    check=True,
)
print(uv_version.stdout.strip())
