package main

import "testing"

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"HOLDINGS_EXIST", "HoldingsExist"},
		{"UNAUTHORIZED", "Unauthorized"},
		{"DISTRIBUTION_ALREADY_CONFIRMED", "DistributionAlreadyConfirmed"},
		{"FUND_NOT_FOUND", "FundNotFound"},
		{"", ""},
		{"A", "A"},
		{"A_B_C", "ABC"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
