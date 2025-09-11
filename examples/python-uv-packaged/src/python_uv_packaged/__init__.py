import os
path = os.environ['PATH']
assert path.startswith("/app/.venv/bin"), f"Expected PATH to start with /app/.venv/bin but got {path}"

def main() -> None:
    print("Hello from python-uv-packaged!")
