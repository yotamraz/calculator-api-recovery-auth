"""Unit tests for the pure calculator functions in calculator_api.core."""

import pytest

from calculator_api.core import add, divide, multiply, subtract


# --- add ---


class TestAdd:
    """Tests for the add function."""

    def test_positive_numbers(self):
        assert add(2, 3) == 5

    def test_negative_numbers(self):
        assert add(-2, -3) == -5

    def test_mixed_signs(self):
        assert add(-2, 3) == 1

    def test_zeros(self):
        assert add(0, 0) == 0

    def test_float_values(self):
        assert add(1.5, 2.5) == 4.0

    def test_identity(self):
        assert add(42, 0) == 42


# --- subtract ---


class TestSubtract:
    """Tests for the subtract function."""

    def test_positive_numbers(self):
        assert subtract(5, 3) == 2

    def test_negative_result(self):
        assert subtract(3, 5) == -2

    def test_negative_numbers(self):
        assert subtract(-2, -3) == 1

    def test_zeros(self):
        assert subtract(0, 0) == 0

    def test_float_values(self):
        assert subtract(5.5, 2.5) == 3.0

    def test_identity(self):
        assert subtract(42, 0) == 42


# --- multiply ---


class TestMultiply:
    """Tests for the multiply function."""

    def test_positive_numbers(self):
        assert multiply(2, 3) == 6

    def test_negative_numbers(self):
        assert multiply(-2, -3) == 6

    def test_mixed_signs(self):
        assert multiply(-2, 3) == -6

    def test_by_zero(self):
        assert multiply(5, 0) == 0

    def test_by_one(self):
        assert multiply(7, 1) == 7

    def test_float_values(self):
        assert multiply(2.5, 4.0) == 10.0


# --- divide ---


class TestDivide:
    """Tests for the divide function."""

    def test_positive_numbers(self):
        assert divide(6, 3) == 2.0

    def test_negative_numbers(self):
        assert divide(-6, -3) == 2.0

    def test_mixed_signs(self):
        assert divide(-6, 3) == -2.0

    def test_float_result(self):
        assert divide(7, 2) == 3.5

    def test_by_one(self):
        assert divide(42, 1) == 42.0

    def test_divide_by_zero_raises(self):
        with pytest.raises(ValueError, match="Cannot divide by zero"):
            divide(1, 0)

    def test_zero_divided_by_nonzero(self):
        assert divide(0, 5) == 0.0
