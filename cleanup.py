"""Cleanup script: remove stale database files."""
import os

# Remove stale database files for both old and new DB names and paths
paths = [
    "calculator.db", "calculator2.db",
    "/tmp/calculator_test.db",
]
for db_path in paths:
    for suffix in ["", "-shm", "-wal"]:
        full_path = db_path + suffix
        try:
            os.remove(full_path)
            print(f"Removed {full_path}")
        except (FileNotFoundError, PermissionError, OSError) as e:
            if not isinstance(e, FileNotFoundError):
                print(f"Could not remove {full_path}: {e}")

print("Cleanup complete")
