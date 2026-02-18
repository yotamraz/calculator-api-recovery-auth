"""Tests for authentication endpoints and auth flow."""

from typing import Annotated

from fastapi import Depends
from fastapi.testclient import TestClient

from app.auth import get_current_user
from app.models import User


class TestRegister:
    """Tests for POST /auth/register."""

    def test_register_success(self, client: TestClient):
        """Registering a new user returns 201 with user info."""
        response = client.post(
            "/auth/register",
            json={"username": "newuser", "password": "secret123"},
        )
        assert response.status_code == 201
        data = response.json()
        assert data["username"] == "newuser"
        assert "id" in data
        assert "created_at" in data
        assert "password" not in data
        assert "hashed_password" not in data

    def test_register_duplicate_username(self, client: TestClient):
        """Registering with an existing username returns 400."""
        client.post(
            "/auth/register",
            json={"username": "dupeuser", "password": "pass1"},
        )
        response = client.post(
            "/auth/register",
            json={"username": "dupeuser", "password": "pass2"},
        )
        assert response.status_code == 400
        assert response.json()["detail"] == "Username already taken"


class TestLogin:
    """Tests for POST /auth/token."""

    def test_login_success(self, client: TestClient):
        """Login with valid credentials returns a token."""
        client.post(
            "/auth/register",
            json={"username": "loginuser", "password": "mypassword"},
        )
        response = client.post(
            "/auth/token",
            data={"username": "loginuser", "password": "mypassword"},
        )
        assert response.status_code == 200
        data = response.json()
        assert "access_token" in data
        assert data["token_type"] == "bearer"

    def test_login_wrong_password(self, client: TestClient):
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

    def test_login_nonexistent_user(self, client: TestClient):
        """Login with a username that doesn't exist returns 401."""
        response = client.post(
            "/auth/token",
            data={"username": "ghost", "password": "nope"},
        )
        assert response.status_code == 401


class TestProtectedAccess:
    """Tests for token-based access to protected endpoints."""

    def _add_protected_endpoint(self, client: TestClient) -> None:
        """Add a temporary protected endpoint for testing auth access."""
        @client.app.get("/test-protected")
        def protected_route(
            current_user: Annotated[User, Depends(get_current_user)],
        ):
            return {"username": current_user.username}

    def test_no_token_returns_401(self, client: TestClient):
        """Accessing a protected endpoint without a token returns 401."""
        self._add_protected_endpoint(client)
        response = client.get("/test-protected")
        assert response.status_code == 401

    def test_invalid_token_returns_401(self, client: TestClient):
        """Accessing a protected endpoint with an invalid token returns 401."""
        self._add_protected_endpoint(client)
        response = client.get(
            "/test-protected",
            headers={"Authorization": "Bearer invalid-token"},
        )
        assert response.status_code == 401

    def test_valid_token_grants_access(
        self, client: TestClient, auth_headers: dict
    ):
        """Accessing a protected endpoint with a valid token succeeds."""
        self._add_protected_endpoint(client)
        response = client.get(
            "/test-protected",
            headers=auth_headers,
        )
        assert response.status_code == 200
        assert response.json()["username"] == "testuser"
