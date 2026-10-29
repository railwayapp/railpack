# Intentionally named server.py to verify startup uses the detected module instead of main.
from fastapi import FastAPI

app = FastAPI()


@app.get("/")
def read_root():
    return {"message": "Welcome to FastAPI"}
