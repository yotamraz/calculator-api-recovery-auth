"""Unit tests for the core calculator functions."""

import pytest

from calculator_api.core import add, divide, multiply, subtract


class TestAdd:
    """Tests for the add function."""

    def test_add_positive_numbers(self):
        assert add(2, 3) == 5

    def test_add_negative_numbers(self):
        assert add(-1, -2) == -3

    def test_add_mixed_signs(self):
        assert add(-1, 3) == 2

    def test_add_zeros(self):
        assert add(0, 0) == 0

    def test_add_floats(self):
        assert add(1.5, 2.5) == 4.0


class TestSubtract:
    """Tests for the subtract function."""

    def test_subtract_positive_numbers(self):
        assert subtract(5, 3) == 2

    def test_subtract_negative_result(self):
        assert subtract(3, 5) == -2

    def test_subtract_negative_numbers(self):
        assert subtract(-1, -2) == 1

    def test_subtract_zeros(self):
        assert subtract(0, 0) == 0

    def test_subtract_floats(self):
        assert subtract(5.5, 2.5) == 3.0


class TestMultiply:
    """Tests for the multiply function."""

    def test_multiply_positive_numbers(self):
        assert multiply(2, 3) == 6

    def test_multiply_by_zero(self):
        assert multiply(5, 0) == 0

    def test_multiply_negative_numbers(self):
        assert multiply(-2, -3) == 6

    def test_multiply_mixed_signs(self):
        assert multiply(-2, 3) == -6

    def test_multiply_floats(self):
        assert multiply(1.5, 2.0) == 3.0


class TestDivide:
    """Tests for the divide function."""

    def test_divide_positive_numbers(self):
        assert divide(6, 3) == 2.0

    def test_divide_with_remainder(self):
        assert divide(7, 2) == 3.5

    def test_divide_negative_numbers(self):
        assert divide(-6, -3) == 2.0

    def test_divide_mixed_signs(self):
        assert divide(-6, 3) == -2.0

    def test_divide_by_zero_raises_value_error(self):
        with pytest.raises(ValueError, match="Cannot divide by zero"):
            divide(1, 0)

    def test_divide_zero_by_number(self):
        assert divide(0, 5) == 0.0

    def test_divide_floats(self):
        assert divide(5.0, 2.0) == 2.5
