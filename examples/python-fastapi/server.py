# intentionally named server.py to catch startup commands that hardcode the main module
from fastapi import FastAPI

app = FastAPI()


@app.get("/")
def read_root():
    return {"message": "Welcome to FastAPI"}
