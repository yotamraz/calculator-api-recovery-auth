package main

import (
	"math"
	"testing"
)

// --- Add ---

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -2, -3, -5},
		{"mixed signs", -2, 3, 1},
		{"with zero", 5, 0, 5},
		{"both zero", 0, 0, 0},
		{"large numbers", 1e15, 2e15, 3e15},
		{"decimal numbers", 1.5, 2.5, 4.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// --- Subtract ---

func TestSubtract(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 5, 3, 2},
		{"negative result", 3, 5, -2},
		{"negative numbers", -2, -3, 1},
		{"with zero", 5, 0, 5},
		{"from zero", 0, 5, -5},
		{"both zero", 0, 0, 0},
		{"decimal numbers", 5.5, 2.5, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Subtract(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Subtract(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// --- Multiply ---

func TestMultiply(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 2, 3, 6},
		{"negative numbers", -2, -3, 6},
		{"mixed signs", -2, 3, -6},
		{"with zero", 5, 0, 0},
		{"both zero", 0, 0, 0},
		{"with one", 7, 1, 7},
		{"decimal numbers", 1.5, 2.0, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Multiply(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Multiply(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// --- Divide ---

func TestDivide(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 6, 3, 2},
		{"negative numbers", -6, -3, 2},
		{"mixed signs", -6, 3, -2},
		{"zero dividend", 0, 5, 0},
		{"decimal result", 7, 2, 3.5},
		{"with one", 7, 1, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Divide(tt.a, tt.b)
			if err != nil {
				t.Errorf("Divide(%v, %v) unexpected error: %v", tt.a, tt.b, err)
			}
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Divide(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestDivideByZero(t *testing.T) {
	result, err := Divide(5, 0)
	if err == nil {
		t.Fatal("Divide(5, 0) expected an error, got nil")
	}
	if err.Error() != "Cannot divide by zero" {
		t.Errorf("Divide(5, 0) error = %q, want %q", err.Error(), "Cannot divide by zero")
	}
	if result != 0 {
		t.Errorf("Divide(5, 0) result = %v, want 0", result)
	}
}

func TestDivideByZeroNegative(t *testing.T) {
	_, err := Divide(-5, 0)
	if err == nil {
		t.Fatal("Divide(-5, 0) expected an error, got nil")
	}
	if err.Error() != "Cannot divide by zero" {
		t.Errorf("Divide(-5, 0) error = %q, want %q", err.Error(), "Cannot divide by zero")
	}
}

func TestDivideZeroByZero(t *testing.T) {
	_, err := Divide(0, 0)
	if err == nil {
		t.Fatal("Divide(0, 0) expected an error, got nil")
	}
	if err.Error() != "Cannot divide by zero" {
		t.Errorf("Divide(0, 0) error = %q, want %q", err.Error(), "Cannot divide by zero")
	}
}
