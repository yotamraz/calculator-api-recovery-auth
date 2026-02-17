"""Cleanup script: remove stale database files."""
import os

# Remove stale database files for both old and new DB names
for db_name in ["calculator.db", "calculator2.db"]:
    for suffix in ["", "-shm", "-wal"]:
        db_path = db_name + suffix
        try:
            os.remove(db_path)
            print(f"Removed {db_path}")
        except FileNotFoundError:
            pass

print("Cleanup complete")
