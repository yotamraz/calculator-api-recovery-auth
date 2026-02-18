"""Start the FastAPI app using uvicorn programmatically with robust process management."""
import multiprocessing
import os
import signal
import sys


def run_server():
    """Run the uvicorn server - this function runs in a child process."""
    # Ignore SIGHUP so server survives shell exit
    signal.signal(signal.SIGHUP, signal.SIG_IGN)
    # Create new session to fully detach
    os.setsid()

    import uvicorn
    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, log_level="info")


def main():
    # Ignore SIGHUP so parent survives shell exit
    signal.signal(signal.SIGHUP, signal.SIG_IGN)

    # Use fork-based multiprocessing to start server
    proc = multiprocessing.Process(target=run_server, daemon=False)
    proc.start()

    # Write PID for cleanup
    with open("uvicorn.pid", "w") as f:
        f.write(str(proc.pid))

    print(f"Server started with PID {proc.pid}", flush=True)

    # Parent exits immediately - child continues in its own session


if __name__ == "__main__":
    multiprocessing.set_start_method("fork")
    main()
