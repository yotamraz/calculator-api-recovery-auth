"""Bootstrap script to register the initial test user."""
import requests

resp = requests.post(
    "http://localhost:8000/auth/register",
    json={"username": "testuser", "password": "testpass123"},
)
print(f"Bootstrap register: {resp.status_code} {resp.text}")
