import os
import sys
from flask import Flask

path = os.environ['PATH']
assert path.startswith("/app/.venv/bin"), f"Expected PATH to start with /app/.venv/bin but got {path}"

print("Hello from Python UV!")
print(f"Python version: {sys.version.split()[0]}")

app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello from Python Flask!"

if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8000))
    app.run(host="0.0.0.0", port=port)
