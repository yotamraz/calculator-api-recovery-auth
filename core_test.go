package main

import (
	"errors"
	"math"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -2, -3, -5},
		{"mixed signs", -2, 3, 1},
		{"zeros", 0, 0, 0},
		{"with zero", 5, 0, 5},
		{"decimals", 1.5, 2.5, 4.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("add(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"positive numbers", 5, 3, 2},
		{"negative numbers", -2, -3, 1},
		{"mixed signs", -2, 3, -5},
		{"zeros", 0, 0, 0},
		{"from zero", 0, 5, -5},
		{"decimals", 5.5, 2.5, 3.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := subtract(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("subtract(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"positive numbers", 2, 3, 6},
		{"negative numbers", -2, -3, 6},
		{"mixed signs", -2, 3, -6},
		{"by zero", 5, 0, 0},
		{"by one", 7, 1, 7},
		{"decimals", 1.5, 2.0, 3.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := multiply(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("multiply(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool
	}{
		{"positive numbers", 6, 3, 2, false},
		{"negative numbers", -6, -3, 2, false},
		{"mixed signs", -6, 3, -2, false},
		{"result is decimal", 7, 2, 3.5, false},
		{"zero numerator", 0, 5, 0, false},
		{"divide by zero", 5, 0, 0, true},
		{"zero divide by zero", 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := divide(tt.a, tt.b)
			if tt.wantErr {
				if err == nil {
					t.Errorf("divide(%v, %v) expected error, got nil", tt.a, tt.b)
				}
				if !errors.Is(err, ErrDivideByZero) {
					t.Errorf("divide(%v, %v) error = %v, want ErrDivideByZero", tt.a, tt.b, err)
				}
			} else {
				if err != nil {
					t.Errorf("divide(%v, %v) unexpected error: %v", tt.a, tt.b, err)
				}
				if math.Abs(got-tt.want) > 1e-10 {
					t.Errorf("divide(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
				}
			}
		})
	}
}

func TestDivideByZeroErrorMessage(t *testing.T) {
	_, err := divide(1, 0)
	if err == nil {
		t.Fatal("expected error for divide by zero")
	}
	want := "Cannot divide by zero"
	if err.Error() != want {
		t.Errorf("error message = %q, want %q", err.Error(), want)
	}
}
