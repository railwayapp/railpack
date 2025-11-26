import sys
import numpy as np
import pandas as pd
import os

path = os.environ['PATH']
assert path.startswith("/app/.venv/bin"), f"Expected PATH to start with /app/.venv/bin but got {path}"

print(f"Python version: {sys.version.split()[0]}")
print("numpy", np.__version__)
print("pandas", pd.__version__)

print("Hello from pip")
