import os
import psycopg

print("psycopg version:", psycopg.__version__)

major_version = int(psycopg.__version__.split('.')[0])
if major_version != 3:
    print(f"ERROR: Expected psycopg 3.x but got {psycopg.__version__}")
    exit(1)

print("Using psycopg3 - libpq is available")

database_url = os.environ.get("DATABASE_URL")
if not database_url:
    print("No DATABASE_URL provided, skipping connection test")
    exit(0)

print(f"Connecting to database...")
conn = psycopg.connect(database_url)
print("Successfully connected to PostgreSQL!")

with conn.cursor() as cur:
    cur.execute("SELECT version();")
    version = cur.fetchone()[0]
    print(f"PostgreSQL version: {version}")

    cur.execute("SELECT 1 + 1 AS result;")
    result = cur.fetchone()[0]
    print(f"Test query result: 1 + 1 = {result}")

conn.close()
print("Connection closed successfully")
