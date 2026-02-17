"""Tests for calculator endpoints and core logic unit tests."""

import pytest
from fastapi.testclient import TestClient

from app.core import add, divide, multiply, subtract


# --- Unit tests for app.core ---


class TestCoreAdd:
    def test_positive_numbers(self):
        assert add(2, 3) == 5

    def test_negative_numbers(self):
        assert add(-1, -2) == -3

    def test_zero(self):
        assert add(0, 0) == 0

    def test_floats(self):
        assert add(1.5, 2.5) == 4.0


class TestCoreSubtract:
    def test_positive_numbers(self):
        assert subtract(5, 3) == 2

    def test_negative_result(self):
        assert subtract(2, 5) == -3

    def test_zero(self):
        assert subtract(0, 0) == 0


class TestCoreMultiply:
    def test_positive_numbers(self):
        assert multiply(3, 4) == 12

    def test_by_zero(self):
        assert multiply(5, 0) == 0

    def test_negative_numbers(self):
        assert multiply(-2, 3) == -6


class TestCoreDivide:
    def test_positive_numbers(self):
        assert divide(10, 2) == 5.0

    def test_float_result(self):
        assert divide(7, 2) == 3.5

    def test_divide_by_zero_raises(self):
        with pytest.raises(ValueError, match="Cannot divide by zero"):
            divide(5, 0)


# --- Endpoint tests ---


def test_add(client: TestClient, auth_headers: dict[str, str]):
    response = client.post("/add", json={"a": 10, "b": 5}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json() == {"result": 15.0}


def test_subtract(client: TestClient, auth_headers: dict[str, str]):
    response = client.post("/subtract", json={"a": 10, "b": 3}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json() == {"result": 7.0}


def test_multiply(client: TestClient, auth_headers: dict[str, str]):
    response = client.post("/multiply", json={"a": 7, "b": 6}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json() == {"result": 42.0}


def test_divide(client: TestClient, auth_headers: dict[str, str]):
    response = client.post("/divide", json={"a": 20, "b": 4}, headers=auth_headers)
    assert response.status_code == 200
    assert response.json() == {"result": 5.0}


def test_divide_by_zero(client: TestClient, auth_headers: dict[str, str]):
    response = client.post("/divide", json={"a": 10, "b": 0}, headers=auth_headers)
    assert response.status_code == 400
    assert response.json()["detail"] == "Cannot divide by zero"


def test_calculator_unauthenticated(client: TestClient):
    for endpoint in ["/add", "/subtract", "/multiply", "/divide"]:
        response = client.post(endpoint, json={"a": 1, "b": 2})
        assert response.status_code == 401, f"{endpoint} should require authentication"
