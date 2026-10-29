# Intentionally named app.py to verify startup uses the detected module instead of main.
import os
from flask import Flask

app = Flask(__name__)


@app.route('/')
def hello():
    return 'Welcome to Flask'
