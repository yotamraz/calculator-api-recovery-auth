"""Bootstrap script to register a test user."""
import httpx

resp = httpx.post(
    "http://localhost:8000/auth/register",
    json={"username": "testuser", "password": "testpass123"},
)
print(resp.status_code, resp.text)
