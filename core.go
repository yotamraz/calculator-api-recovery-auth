package main

import "errors"

// ErrDivideByZero is returned when a division by zero is attempted.
var ErrDivideByZero = errors.New("Cannot divide by zero")

// add returns the sum of a and b.
func add(a, b float64) float64 {
	return a + b
}

// subtract returns a minus b.
func subtract(a, b float64) float64 {
	return a - b
}

// multiply returns the product of a and b.
func multiply(a, b float64) float64 {
	return a * b
}

// divide returns a divided by b. It returns ErrDivideByZero if b is zero.
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}
