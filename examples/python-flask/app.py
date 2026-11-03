# intentionally named app.py to catch startup commands that hardcode the main module
import os
from flask import Flask

app = Flask(__name__)


@app.route('/')
def hello():
    return 'Welcome to Flask'
