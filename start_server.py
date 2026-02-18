"""Start the uvicorn server as a fully detached daemon process."""
import os
import sys
import time


def main():
    # Get the path to the Python interpreter in the venv
    python = sys.executable

    # Double-fork to fully detach from the parent process (Unix daemon pattern)
    pid = os.fork()
    if pid > 0:
        # Parent process: wait briefly for daemon to initialize, then exit
        time.sleep(1)
        print("Server daemon launched")
        sys.exit(0)

    # First child: create new session to detach from terminal
    os.setsid()

    # Second fork to prevent reacquiring a terminal
    pid = os.fork()
    if pid > 0:
        os._exit(0)

    # Grandchild (daemon): fully detached from terminal and parent
    # Redirect file descriptors
    log = open("uvicorn.log", "w")

    # Write PID file
    with open("uvicorn.pid", "w") as f:
        f.write(str(os.getpid()))

    # Redirect stdout/stderr to log file
    os.dup2(log.fileno(), 1)
    os.dup2(log.fileno(), 2)

    # Replace this process with uvicorn
    os.execvp(python, [
        python, "-m", "uvicorn",
        "app.main:app",
        "--host", "0.0.0.0",
        "--port", "8000",
    ])


if __name__ == "__main__":
    main()
