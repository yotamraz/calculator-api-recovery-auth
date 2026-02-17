"""Tests for auth endpoints (register, login, token usage, errors)."""

from fastapi.testclient import TestClient


def test_register_success(client: TestClient):
    """POST /auth/register creates a new user and returns 201."""
    response = client.post(
        "/auth/register",
        json={"username": "newuser", "password": "newpassword"},
    )
    assert response.status_code == 201
    data = response.json()
    assert data["username"] == "newuser"
    assert "id" in data
    assert "created_at" in data


def test_register_duplicate(client: TestClient):
    """POST /auth/register with an existing username returns 400."""
    client.post(
        "/auth/register",
        json={"username": "dupuser", "password": "pass1"},
    )
    response = client.post(
        "/auth/register",
        json={"username": "dupuser", "password": "pass2"},
    )
    assert response.status_code == 400
    assert response.json()["detail"] == "Username already taken"


def test_login_success(client: TestClient):
    """POST /auth/token with valid credentials returns a token."""
    client.post(
        "/auth/register",
        json={"username": "loginuser", "password": "loginpass"},
    )
    response = client.post(
        "/auth/token",
        data={"username": "loginuser", "password": "loginpass"},
    )
    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert data["token_type"] == "bearer"


def test_login_invalid_credentials(client: TestClient):
    """POST /auth/token with invalid credentials returns 401."""
    client.post(
        "/auth/register",
        json={"username": "authuser", "password": "correctpass"},
    )
    response = client.post(
        "/auth/token",
        data={"username": "authuser", "password": "wrongpass"},
    )
    assert response.status_code == 401
    assert response.json()["detail"] == "Incorrect username or password"


def test_protected_endpoint_without_token(client: TestClient):
    """Accessing a protected endpoint without a token returns 401."""
    response = client.post("/add", json={"a": 1, "b": 2})
    assert response.status_code == 401


def test_protected_endpoint_with_valid_token(auth_client: TestClient):
    """Accessing a protected endpoint with a valid token succeeds."""
    response = auth_client.post("/add", json={"a": 1, "b": 2})
    assert response.status_code == 200
    assert response.json()["result"] == 3.0
