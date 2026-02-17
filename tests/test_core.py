"""Tests for core arithmetic functions."""

import pytest

from app.core import add, divide, multiply, subtract


def test_add():
    assert add(2, 3) == 5.0
    assert add(-1, 1) == 0.0
    assert add(0, 0) == 0.0


def test_subtract():
    assert subtract(10, 4) == 6.0
    assert subtract(1, 1) == 0.0
    assert subtract(0, 5) == -5.0


def test_multiply():
    assert multiply(3, 7) == 21.0
    assert multiply(0, 100) == 0.0
    assert multiply(-2, 3) == -6.0


def test_divide():
    assert divide(10, 2) == 5.0
    assert divide(7, 2) == 3.5
    assert divide(-6, 3) == -2.0


def test_divide_by_zero():
    with pytest.raises(ValueError, match="Cannot divide by zero"):
        divide(1, 0)
