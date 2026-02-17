"""Tests for the auth flow: registration, login, and token-based access."""

from fastapi.testclient import TestClient


def test_register_success(client: TestClient):
    """Successful registration returns 201 with user data."""
    response = client.post(
        "/auth/register",
        json={"username": "newuser", "password": "pass123"},
    )
    assert response.status_code == 201
    data = response.json()
    assert data["username"] == "newuser"
    assert "id" in data
    assert "created_at" in data


def test_register_duplicate_username(client: TestClient):
    """Registering with an existing username returns 400."""
    client.post(
        "/auth/register",
        json={"username": "dupuser", "password": "pass123"},
    )
    response = client.post(
        "/auth/register",
        json={"username": "dupuser", "password": "otherpass"},
    )
    assert response.status_code == 400
    assert response.json()["detail"] == "Username already taken"


def test_login_success(client: TestClient):
    """Successful login returns a token."""
    client.post(
        "/auth/register",
        json={"username": "loginuser", "password": "pass123"},
    )
    response = client.post(
        "/auth/token",
        data={"username": "loginuser", "password": "pass123"},
    )
    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert data["token_type"] == "bearer"


def test_login_wrong_password(client: TestClient):
    """Login with wrong password returns 401."""
    client.post(
        "/auth/register",
        json={"username": "wrongpwuser", "password": "correct"},
    )
    response = client.post(
        "/auth/token",
        data={"username": "wrongpwuser", "password": "incorrect"},
    )
    assert response.status_code == 401
    assert response.json()["detail"] == "Incorrect username or password"


def test_login_nonexistent_user(client: TestClient):
    """Login with a non-existent user returns 401."""
    response = client.post(
        "/auth/token",
        data={"username": "doesnotexist", "password": "anything"},
    )
    assert response.status_code == 401
    assert response.json()["detail"] == "Incorrect username or password"


def test_protected_endpoint_without_token(client: TestClient):
    """Accessing a protected endpoint without a token returns 401."""
    response = client.post("/add", json={"a": 1, "b": 2})
    assert response.status_code == 401


def test_protected_endpoint_with_valid_token(client: TestClient, auth_headers: dict):
    """Accessing a protected endpoint with a valid token succeeds."""
    response = client.post("/add", json={"a": 1, "b": 2}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json()["result"] == 3.0
