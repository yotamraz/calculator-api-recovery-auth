"""Tests for arithmetic endpoints (/add, /subtract, /multiply, /divide)."""

from fastapi.testclient import TestClient


def test_add(auth_client: TestClient):
    """POST /add returns the sum of two numbers."""
    response = auth_client.post("/add", json={"a": 5, "b": 3})
    assert response.status_code == 200
    assert response.json()["result"] == 8.0


def test_subtract(auth_client: TestClient):
    """POST /subtract returns the difference of two numbers."""
    response = auth_client.post("/subtract", json={"a": 10, "b": 4})
    assert response.status_code == 200
    assert response.json()["result"] == 6.0


def test_multiply(auth_client: TestClient):
    """POST /multiply returns the product of two numbers."""
    response = auth_client.post("/multiply", json={"a": 3, "b": 7})
    assert response.status_code == 200
    assert response.json()["result"] == 21.0


def test_divide(auth_client: TestClient):
    """POST /divide returns the quotient of two numbers."""
    response = auth_client.post("/divide", json={"a": 10, "b": 2})
    assert response.status_code == 200
    assert response.json()["result"] == 5.0


def test_divide_by_zero(auth_client: TestClient):
    """POST /divide with b=0 returns 400 with division-by-zero error."""
    response = auth_client.post("/divide", json={"a": 1, "b": 0})
    assert response.status_code == 400
    assert response.json()["detail"] == "Cannot divide by zero"


def test_add_requires_auth(client: TestClient):
    """POST /add without auth returns 401."""
    response = client.post("/add", json={"a": 1, "b": 2})
    assert response.status_code == 401


def test_subtract_requires_auth(client: TestClient):
    """POST /subtract without auth returns 401."""
    response = client.post("/subtract", json={"a": 1, "b": 2})
    assert response.status_code == 401


def test_multiply_requires_auth(client: TestClient):
    """POST /multiply without auth returns 401."""
    response = client.post("/multiply", json={"a": 1, "b": 2})
    assert response.status_code == 401


def test_divide_requires_auth(client: TestClient):
    """POST /divide without auth returns 401."""
    response = client.post("/divide", json={"a": 1, "b": 2})
    assert response.status_code == 401
