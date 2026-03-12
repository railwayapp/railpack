import os
import psycopg
import ctypes.util

print(f"psycopg version: {psycopg.__version__}")

# Verify libpq is not installed on the system
libpq_path = ctypes.util.find_library('pq')
if libpq_path:
    print(f"ERROR: Found system libpq at {libpq_path}. It should not be present.")
    exit(1)
print("System libpq not found")

database_url = os.environ["DATABASE_URL"]

print("Connecting to database...")
with psycopg.connect(database_url) as conn:
    print("Successfully connected to PostgreSQL!")
    
    with conn.cursor() as cur:
        cur.execute("SELECT 1 + 1 AS result;")
        result = cur.fetchone()[0]
        print(f"Test query result: 1 + 1 = {result}")

print("Connection closed successfully")
