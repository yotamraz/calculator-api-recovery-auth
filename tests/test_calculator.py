"""Tests for calculator endpoints and core logic unit tests."""

import pytest
from fastapi.testclient import TestClient

from app.core import add, divide, multiply, subtract


# --- Core function unit tests ---


class TestCoreFunctions:
    """Unit tests for app.core arithmetic functions."""

    def test_add(self):
        assert add(1, 2) == 3
        assert add(-1, 1) == 0
        assert add(0.1, 0.2) == pytest.approx(0.3)

    def test_subtract(self):
        assert subtract(5, 3) == 2
        assert subtract(0, 5) == -5

    def test_multiply(self):
        assert multiply(4, 3) == 12
        assert multiply(0, 100) == 0

    def test_divide(self):
        assert divide(10, 2) == 5.0
        assert divide(7, 2) == 3.5

    def test_divide_by_zero(self):
        with pytest.raises(ValueError, match="Cannot divide by zero"):
            divide(1, 0)


# --- Calculator endpoint tests ---


def test_add_endpoint(client: TestClient, auth_headers: dict):
    """POST /add returns the sum."""
    response = client.post("/add", json={"a": 5, "b": 3}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json() == {"result": 8.0}


def test_subtract_endpoint(client: TestClient, auth_headers: dict):
    """POST /subtract returns the difference."""
    response = client.post("/subtract", json={"a": 10, "b": 4}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json() == {"result": 6.0}


def test_multiply_endpoint(client: TestClient, auth_headers: dict):
    """POST /multiply returns the product."""
    response = client.post("/multiply", json={"a": 7, "b": 6}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json() == {"result": 42.0}


def test_divide_endpoint(client: TestClient, auth_headers: dict):
    """POST /divide returns the quotient."""
    response = client.post("/divide", json={"a": 10, "b": 2}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json() == {"result": 5.0}


def test_divide_by_zero_endpoint(client: TestClient, auth_headers: dict):
    """POST /divide with b=0 returns 400 with error message."""
    response = client.post("/divide", json={"a": 1, "b": 0}, headers=auth_headers)
    assert response.status_code == 400
    assert response.json()["detail"] == "Cannot divide by zero"


def test_add_requires_auth(client: TestClient):
    """POST /add without token returns 401."""
    response = client.post("/add", json={"a": 1, "b": 2})
    assert response.status_code == 401


def test_subtract_requires_auth(client: TestClient):
    """POST /subtract without token returns 401."""
    response = client.post("/subtract", json={"a": 1, "b": 2})
    assert response.status_code == 401


def test_multiply_requires_auth(client: TestClient):
    """POST /multiply without token returns 401."""
    response = client.post("/multiply", json={"a": 1, "b": 2})
    assert response.status_code == 401


def test_divide_requires_auth(client: TestClient):
    """POST /divide without token returns 401."""
    response = client.post("/divide", json={"a": 1, "b": 2})
    assert response.status_code == 401
