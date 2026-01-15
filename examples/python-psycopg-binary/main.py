import os
import psycopg

print(f"psycopg version: {psycopg.__version__}")

database_url = os.environ["DATABASE_URL"]

print("Connecting to database...")
with psycopg.connect(database_url) as conn:
    print("Successfully connected to PostgreSQL!")
    
    with conn.cursor() as cur:
        cur.execute("SELECT 1 + 1 AS result;")
        result = cur.fetchone()[0]
        print(f"Test query result: 1 + 1 = {result}")

print("Connection closed successfully")
