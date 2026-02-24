package main

import (
	"errors"
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
		{"zero and positive", 0, 5, 5},
		{"decimals", 1.5, 2.5, 4},
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
		{"negative result", 3, 5, -2},
		{"negative numbers", -2, -3, 1},
		{"zeros", 0, 0, 0},
		{"subtract from zero", 0, 5, -5},
		{"decimals", 5.5, 2.5, 3},
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
		{"multiply by zero", 5, 0, 0},
		{"multiply by one", 5, 1, 5},
		{"decimals", 1.5, 2, 3},
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
		wantErr error
	}{
		{"positive numbers", 6, 3, 2, nil},
		{"negative numbers", -6, -3, 2, nil},
		{"mixed signs", -6, 3, -2, nil},
		{"decimal result", 7, 2, 3.5, nil},
		{"divide zero", 0, 5, 0, nil},
		{"divide by zero", 5, 0, 0, ErrDivideByZero},
		{"zero divide by zero", 0, 0, 0, ErrDivideByZero},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := divide(tt.a, tt.b)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("divide(%v, %v) expected error %v, got nil", tt.a, tt.b, tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("divide(%v, %v) error = %v, want %v", tt.a, tt.b, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("divide(%v, %v) unexpected error: %v", tt.a, tt.b, err)
			}
			if got != tt.want {
				t.Errorf("divide(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
