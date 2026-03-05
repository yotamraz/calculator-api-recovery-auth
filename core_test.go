package main

import (
	"math"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -2, -3, -5},
		{"mixed signs", -2, 3, 1},
		{"zeros", 0, 0, 0},
		{"large numbers", 1e15, 1e15, 2e15},
		{"small decimals", 0.1, 0.2, 0.30000000000000004},
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

func TestSubtract(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 5, 3, 2},
		{"negative numbers", -5, -3, -2},
		{"result is negative", 3, 5, -2},
		{"zeros", 0, 0, 0},
		{"large numbers", 2e15, 1e15, 1e15},
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

func TestMultiply(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 2, 3, 6},
		{"negative numbers", -2, -3, 6},
		{"mixed signs", -2, 3, -6},
		{"zero factor", 5, 0, 0},
		{"identity", 7, 1, 7},
		{"large numbers", 1e7, 1e7, 1e14},
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

func TestDivide(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 6, 3, 2},
		{"negative numbers", -6, -3, 2},
		{"mixed signs", -6, 3, -2},
		{"fractional result", 7, 2, 3.5},
		{"identity", 7, 1, 7},
		{"small dividend", 1, 1000, 0.001},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Divide(tt.a, tt.b)
			if err != nil {
				t.Errorf("Divide(%v, %v) unexpected error: %v", tt.a, tt.b, err)
			}
			if result != tt.expected {
				t.Errorf("Divide(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestDivideByZero(t *testing.T) {
	_, err := Divide(5, 0)
	if err == nil {
		t.Error("Divide(5, 0) expected error, got nil")
	}
	if err.Error() != "Cannot divide by zero" {
		t.Errorf("Divide(5, 0) error = %q, want %q", err.Error(), "Cannot divide by zero")
	}
}

func TestDivideByZeroNegative(t *testing.T) {
	_, err := Divide(-5, 0)
	if err == nil {
		t.Error("Divide(-5, 0) expected error, got nil")
	}
}

func TestDivideByZeroReturnValue(t *testing.T) {
	result, err := Divide(0, 0)
	if err == nil {
		t.Error("Divide(0, 0) expected error, got nil")
	}
	if result != 0 {
		t.Errorf("Divide(0, 0) returned %v, want 0 on error", result)
	}
}

func TestAddInfinity(t *testing.T) {
	result := Add(math.MaxFloat64, math.MaxFloat64)
	if !math.IsInf(result, 1) {
		t.Errorf("Add(MaxFloat64, MaxFloat64) = %v, want +Inf", result)
	}
}
