"""Cleanup script: remove stale database files."""
import os

# Remove stale database
for db_path in ["calculator.db", "calculator.db-shm", "calculator.db-wal"]:
    try:
        os.remove(db_path)
        print(f"Removed {db_path}")
    except FileNotFoundError:
        pass

print("Cleanup complete")
