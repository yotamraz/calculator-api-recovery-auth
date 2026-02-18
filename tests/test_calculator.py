"""Unit tests for core calculator functions and HTTP-level calculator endpoint tests."""

import pytest
from fastapi.testclient import TestClient

from app.core import add, divide, multiply, subtract


class TestAdd:
    def test_positive_numbers(self):
        assert add(2, 3) == 5

    def test_negative_numbers(self):
        assert add(-1, -2) == -3

    def test_mixed_signs(self):
        assert add(-1, 3) == 2

    def test_zeros(self):
        assert add(0, 0) == 0

    def test_floats(self):
        assert add(1.5, 2.5) == 4.0


class TestSubtract:
    def test_positive_numbers(self):
        assert subtract(5, 3) == 2

    def test_negative_result(self):
        assert subtract(3, 5) == -2

    def test_negative_numbers(self):
        assert subtract(-1, -2) == 1

    def test_zeros(self):
        assert subtract(0, 0) == 0

    def test_floats(self):
        assert subtract(5.5, 2.5) == 3.0


class TestMultiply:
    def test_positive_numbers(self):
        assert multiply(2, 3) == 6

    def test_negative_numbers(self):
        assert multiply(-2, -3) == 6

    def test_mixed_signs(self):
        assert multiply(-2, 3) == -6

    def test_by_zero(self):
        assert multiply(5, 0) == 0

    def test_floats(self):
        assert multiply(2.5, 4) == 10.0


class TestDivide:
    def test_positive_numbers(self):
        assert divide(6, 3) == 2.0

    def test_negative_numbers(self):
        assert divide(-6, -3) == 2.0

    def test_mixed_signs(self):
        assert divide(-6, 3) == -2.0

    def test_floats(self):
        assert divide(7.5, 2.5) == 3.0

    def test_divide_by_zero_raises_value_error(self):
        with pytest.raises(ValueError, match="Cannot divide by zero"):
            divide(1, 0)

    def test_divide_zero_by_nonzero(self):
        assert divide(0, 5) == 0.0


# --- HTTP-level calculator endpoint tests ---


class TestAddEndpoint:
    """Tests for POST /add."""

    def test_requires_auth(self, client: TestClient):
        response = client.post("/add", json={"a": 1, "b": 2})
        assert response.status_code == 401

    def test_add_two_numbers(self, client: TestClient, auth_headers: dict):
        response = client.post("/add", json={"a": 5, "b": 3}, headers=auth_headers)
        assert response.status_code == 200
        assert response.json() == {"result": 8.0}

    def test_add_negative_numbers(self, client: TestClient, auth_headers: dict):
        response = client.post("/add", json={"a": -1, "b": -2}, headers=auth_headers)
        assert response.status_code == 200
        assert response.json() == {"result": -3.0}


class TestSubtractEndpoint:
    """Tests for POST /subtract."""

    def test_requires_auth(self, client: TestClient):
        response = client.post("/subtract", json={"a": 5, "b": 3})
        assert response.status_code == 401

    def test_subtract_two_numbers(self, client: TestClient, auth_headers: dict):
        response = client.post("/subtract", json={"a": 10, "b": 4}, headers=auth_headers)
        assert response.status_code == 200
        assert response.json() == {"result": 6.0}


class TestMultiplyEndpoint:
    """Tests for POST /multiply."""

    def test_requires_auth(self, client: TestClient):
        response = client.post("/multiply", json={"a": 2, "b": 3})
        assert response.status_code == 401

    def test_multiply_two_numbers(self, client: TestClient, auth_headers: dict):
        response = client.post("/multiply", json={"a": 7, "b": 6}, headers=auth_headers)
        assert response.status_code == 200
        assert response.json() == {"result": 42.0}


class TestDivideEndpoint:
    """Tests for POST /divide."""

    def test_requires_auth(self, client: TestClient):
        response = client.post("/divide", json={"a": 10, "b": 2})
        assert response.status_code == 401

    def test_divide_two_numbers(self, client: TestClient, auth_headers: dict):
        response = client.post("/divide", json={"a": 10, "b": 4}, headers=auth_headers)
        assert response.status_code == 200
        assert response.json() == {"result": 2.5}

    def test_divide_by_zero(self, client: TestClient, auth_headers: dict):
        response = client.post("/divide", json={"a": 1, "b": 0}, headers=auth_headers)
        assert response.status_code == 400
        assert response.json()["detail"] == "Cannot divide by zero"
