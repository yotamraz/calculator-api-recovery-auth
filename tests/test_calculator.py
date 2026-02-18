"""Unit tests for core calculator functions in app.core."""

import pytest

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
