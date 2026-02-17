"""Tests for calculator arithmetic endpoints and core unit tests."""

import pytest

from app.core import add, divide, multiply, subtract


# --- Core unit tests ---


class TestCoreAdd:
    """Direct tests for app.core.add."""

    def test_add_positive(self):
        assert add(2, 3) == 5.0

    def test_add_negative(self):
        assert add(-1, -2) == -3.0

    def test_add_zero(self):
        assert add(0, 0) == 0.0


class TestCoreSubtract:
    """Direct tests for app.core.subtract."""

    def test_subtract_positive(self):
        assert subtract(5, 3) == 2.0

    def test_subtract_negative(self):
        assert subtract(-1, -2) == 1.0


class TestCoreMultiply:
    """Direct tests for app.core.multiply."""

    def test_multiply_positive(self):
        assert multiply(3, 4) == 12.0

    def test_multiply_by_zero(self):
        assert multiply(5, 0) == 0.0


class TestCoreDivide:
    """Direct tests for app.core.divide."""

    def test_divide_positive(self):
        assert divide(10, 2) == 5.0

    def test_divide_by_zero(self):
        with pytest.raises(ValueError, match="Cannot divide by zero"):
            divide(10, 0)


# --- Calculator endpoint tests ---


class TestAddEndpoint:
    """Tests for POST /add."""

    def test_add_success(self, auth_client):
        response = auth_client.post("/add", json={"a": 5, "b": 3})
        assert response.status_code == 200
        assert response.json() == {"result": 8.0}

    def test_add_negative_numbers(self, auth_client):
        response = auth_client.post("/add", json={"a": -1, "b": -2})
        assert response.status_code == 200
        assert response.json() == {"result": -3.0}


class TestSubtractEndpoint:
    """Tests for POST /subtract."""

    def test_subtract_success(self, auth_client):
        response = auth_client.post("/subtract", json={"a": 10, "b": 4})
        assert response.status_code == 200
        assert response.json() == {"result": 6.0}


class TestMultiplyEndpoint:
    """Tests for POST /multiply."""

    def test_multiply_success(self, auth_client):
        response = auth_client.post("/multiply", json={"a": 7, "b": 6})
        assert response.status_code == 200
        assert response.json() == {"result": 42.0}


class TestDivideEndpoint:
    """Tests for POST /divide."""

    def test_divide_success(self, auth_client):
        response = auth_client.post("/divide", json={"a": 10, "b": 2})
        assert response.status_code == 200
        assert response.json() == {"result": 5.0}

    def test_divide_by_zero(self, auth_client):
        response = auth_client.post("/divide", json={"a": 10, "b": 0})
        assert response.status_code == 400
        assert response.json()["detail"] == "Cannot divide by zero"


class TestCalculatorUnauthenticated:
    """Tests for unauthenticated access to calculator endpoints."""

    def test_add_no_token(self, client):
        response = client.post("/add", json={"a": 1, "b": 2})
        assert response.status_code == 401

    def test_subtract_no_token(self, client):
        response = client.post("/subtract", json={"a": 1, "b": 2})
        assert response.status_code == 401

    def test_multiply_no_token(self, client):
        response = client.post("/multiply", json={"a": 1, "b": 2})
        assert response.status_code == 401

    def test_divide_no_token(self, client):
        response = client.post("/divide", json={"a": 1, "b": 2})
        assert response.status_code == 401
