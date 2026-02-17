"""Tests for the authentication endpoints."""

from fastapi.testclient import TestClient


def test_register_success(client: TestClient):
    response = client.post(
        "/auth/register",
        json={"username": "alice", "password": "secret123"},
    )
    assert response.status_code == 201
    data = response.json()
    assert data["username"] == "alice"
    assert "id" in data
    assert "created_at" in data


def test_register_duplicate_username(client: TestClient):
    client.post(
        "/auth/register",
        json={"username": "alice", "password": "secret123"},
    )
    response = client.post(
        "/auth/register",
        json={"username": "alice", "password": "other456"},
    )
    assert response.status_code == 400
    assert response.json()["detail"] == "Username already taken"


def test_login_success(client: TestClient):
    client.post(
        "/auth/register",
        json={"username": "alice", "password": "secret123"},
    )
    response = client.post(
        "/auth/token",
        data={"username": "alice", "password": "secret123"},
    )
    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert data["token_type"] == "bearer"


def test_login_invalid_credentials(client: TestClient):
    client.post(
        "/auth/register",
        json={"username": "alice", "password": "secret123"},
    )
    response = client.post(
        "/auth/token",
        data={"username": "alice", "password": "wrongpassword"},
    )
    assert response.status_code == 401
    assert response.json()["detail"] == "Incorrect username or password"


def test_login_nonexistent_user(client: TestClient):
    response = client.post(
        "/auth/token",
        data={"username": "nobody", "password": "nopass"},
    )
    assert response.status_code == 401
    assert response.json()["detail"] == "Incorrect username or password"


def test_protected_endpoint_without_token(client: TestClient):
    response = client.post("/add", json={"a": 1, "b": 2})
    assert response.status_code == 401


def test_protected_endpoint_with_valid_token(client: TestClient, auth_headers: dict[str, str]):
    response = client.post("/add", json={"a": 1, "b": 2}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json()["result"] == 3.0
