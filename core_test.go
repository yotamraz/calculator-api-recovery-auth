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
		{"mixed sign", -2, 3, 1},
		{"with zero", 5, 0, 5},
		{"both zero", 0, 0, 0},
		{"decimals", 1.5, 2.5, 4.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("add(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
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
		{"negative result", 3, 5, -2},
		{"negative numbers", -2, -3, 1},
		{"with zero", 5, 0, 5},
		{"both zero", 0, 0, 0},
		{"decimals", 3.5, 1.5, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subtract(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("subtract(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
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
		{"mixed sign", -2, 3, -6},
		{"with zero", 5, 0, 0},
		{"both zero", 0, 0, 0},
		{"decimals", 1.5, 2.0, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := multiply(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("multiply(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name      string
		a, b      float64
		expected  float64
		expectErr bool
	}{
		{"positive numbers", 6, 3, 2, false},
		{"negative numbers", -6, -3, 2, false},
		{"mixed sign", -6, 3, -2, false},
		{"decimal result", 7, 2, 3.5, false},
		{"divide by zero", 5, 0, 0, true},
		{"zero numerator", 0, 5, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := divide(tt.a, tt.b)
			if tt.expectErr {
				if err == nil {
					t.Errorf("divide(%v, %v) expected error, got nil", tt.a, tt.b)
				}
				if err != ErrDivideByZero {
					t.Errorf("divide(%v, %v) error = %v, want ErrDivideByZero", tt.a, tt.b, err)
				}
			} else {
				if err != nil {
					t.Errorf("divide(%v, %v) unexpected error: %v", tt.a, tt.b, err)
				}
				if math.Abs(result-tt.expected) > 1e-9 {
					t.Errorf("divide(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
				}
			}
		})
	}
}
