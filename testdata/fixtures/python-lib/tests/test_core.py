"""Tests for core module."""

from mylib.core import add, fibonacci


def test_add() -> None:
    assert add(2, 3) == 5


def test_add_negative() -> None:
    assert add(-1, 1) == 0


def test_fibonacci_zero() -> None:
    assert fibonacci(0) == []


def test_fibonacci_one() -> None:
    assert fibonacci(1) == [0]


def test_fibonacci_five() -> None:
    assert fibonacci(5) == [0, 1, 1, 2, 3]
