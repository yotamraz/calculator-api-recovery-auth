"""Cleanup script: kill processes on port 8000 and remove stale database."""
import os
import signal
import subprocess

# Kill any process on port 8000
try:
    result = subprocess.run(
        ["lsof", "-ti", ":8000"],
        capture_output=True, text=True, timeout=5
    )
    if result.stdout.strip():
        for pid in result.stdout.strip().split("\n"):
            pid = pid.strip()
            if pid:
                try:
                    os.kill(int(pid), signal.SIGKILL)
                    print(f"Killed process {pid}")
                except (ProcessLookupError, ValueError):
                    pass
except Exception as e:
    print(f"lsof approach failed: {e}")

# Also try pkill
try:
    subprocess.run(["pkill", "-9", "-f", "uvicorn"], timeout=5)
except Exception:
    pass

# Remove stale database
for db_path in ["calculator.db", "calculator.db-shm", "calculator.db-wal"]:
    try:
        os.remove(db_path)
        print(f"Removed {db_path}")
    except FileNotFoundError:
        pass

import time
time.sleep(2)
print("Cleanup complete")
