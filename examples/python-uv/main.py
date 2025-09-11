from flask import Flask
import os

path = os.environ['PATH']
assert path.startswith("/app/.venv/bin"), f"Expected PATH to start with /app/.venv/bin but got {path}"

app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello from Python Flask!"

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=3333)

print("Hello from Python UV!")
