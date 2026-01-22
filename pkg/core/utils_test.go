package core

import (
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestValidateDecimalPrecision(t *testing.T) {
	tests := []struct {
		name        string
		amount      string
		maxDecimals uint8
		expectError bool
		description string
	}{
		{
			name:        "valid_6_decimals",
			amount:      "1.123456",
			maxDecimals: 6,
			expectError: false,
			description: "Amount with exactly 6 decimals should be valid for 6 decimal limit",
		},
		{
			name:        "valid_less_than_max",
			amount:      "1.123",
			maxDecimals: 6,
			expectError: false,
			description: "Amount with 3 decimals should be valid for 6 decimal limit",
		},
		{
			name:        "valid_whole_number",
			amount:      "100",
			maxDecimals: 6,
			expectError: false,
			description: "Whole number should be valid for any decimal limit",
		},
		{
			name:        "valid_zero",
			amount:      "0",
			maxDecimals: 6,
			expectError: false,
			description: "Zero should be valid for any decimal limit",
		},
		{
			name:        "invalid_too_many_decimals",
			amount:      "1.1234567",
			maxDecimals: 6,
			expectError: true,
			description: "Amount with 7 decimals should be invalid for 6 decimal limit",
		},
		{
			name:        "invalid_8_decimals",
			amount:      "0.12345678",
			maxDecimals: 6,
			expectError: true,
			description: "Amount with 8 decimals should be invalid for 6 decimal limit",
		},
		{
			name:        "valid_18_decimals_eth",
			amount:      "1.123456789012345678",
			maxDecimals: 18,
			expectError: false,
			description: "ETH amount with 18 decimals should be valid for 18 decimal limit",
		},
		{
			name:        "invalid_19_decimals_eth",
			amount:      "1.1234567890123456789",
			maxDecimals: 18,
			expectError: true,
			description: "Amount with 19 decimals should be invalid for 18 decimal limit",
		},
		{
			name:        "valid_usdc_6_decimals",
			amount:      "1000.123456",
			maxDecimals: 6,
			expectError: false,
			description: "USDC amount with 6 decimals should be valid",
		},
		{
			name:        "valid_small_amount",
			amount:      "0.000001",
			maxDecimals: 6,
			expectError: false,
			description: "Very small amount with 6 decimals should be valid",
		},
		{
			name:        "invalid_one_over_limit",
			amount:      "0.0000001",
			maxDecimals: 6,
			expectError: true,
			description: "Amount with one more decimal than allowed should be invalid",
		},
		{
			name:        "valid_large_number_no_decimals",
			amount:      "1000000000",
			maxDecimals: 2,
			expectError: false,
			description: "Large whole number should be valid regardless of decimal limit",
		},
		{
			name:        "valid_2_decimals",
			amount:      "99.99",
			maxDecimals: 2,
			expectError: false,
			description: "Amount with 2 decimals should be valid for 2 decimal limit",
		},
		{
			name:        "invalid_3_decimals_when_2_allowed",
			amount:      "99.999",
			maxDecimals: 2,
			expectError: true,
			description: "Amount with 3 decimals should be invalid for 2 decimal limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := decimal.NewFromString(tt.amount)
			assert.NoError(t, err, "Test setup error: invalid amount string")

			err = ValidateDecimalPrecision(amount, tt.maxDecimals)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), "amount exceeds maximum decimal precision")
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestValidateDecimalPrecision_EdgeCases(t *testing.T) {
	t.Run("negative_amount", func(t *testing.T) {
		amount := decimal.NewFromFloat(-1.123456)
		err := ValidateDecimalPrecision(amount, 6)
		assert.NoError(t, err, "Negative amounts should be validated the same as positive")
	})

	t.Run("negative_amount_too_many_decimals", func(t *testing.T) {
		amount := decimal.NewFromFloat(-1.1234567)
		err := ValidateDecimalPrecision(amount, 6)
		assert.Error(t, err, "Negative amount with too many decimals should fail")
	})

	t.Run("very_large_amount", func(t *testing.T) {
		amount, err := decimal.NewFromString("999999999999999999.123456")
		assert.NoError(t, err)
		err = ValidateDecimalPrecision(amount, 6)
		// This should pass as long as decimals are within limit
		assert.NoError(t, err)
	})

	t.Run("zero_decimal_limit", func(t *testing.T) {
		amount := decimal.NewFromInt(100)
		err := ValidateDecimalPrecision(amount, 0)
		assert.NoError(t, err, "Whole number should be valid for 0 decimal limit")

		amountWithDecimals := decimal.NewFromFloat(100.1)
		err = ValidateDecimalPrecision(amountWithDecimals, 0)
		assert.Error(t, err, "Amount with decimals should fail for 0 decimal limit")
	})
}

func TestDecimalToBigInt(t *testing.T) {
	tests := []struct {
		name        string
		amount      string
		decimals    uint8
		expected    string
		description string
	}{
		{
			name:        "usdc_whole_number",
			amount:      "100",
			decimals:    6,
			expected:    "100000000", // 100 * 10^6
			description: "100 USDC should be 100000000 in smallest unit",
		},
		{
			name:        "usdc_with_decimals",
			amount:      "1.23",
			decimals:    6,
			expected:    "1230000", // 1.23 * 10^6
			description: "1.23 USDC should be 1230000 in smallest unit",
		},
		{
			name:        "usdc_max_decimals",
			amount:      "1.123456",
			decimals:    6,
			expected:    "1123456", // 1.123456 * 10^6
			description: "1.123456 USDC should be 1123456 in smallest unit",
		},
		{
			name:        "usdc_small_amount",
			amount:      "0.000001",
			decimals:    6,
			expected:    "1", // 0.000001 * 10^6 = 1
			description: "0.000001 USDC (smallest unit) should be 1",
		},
		{
			name:        "eth_whole_number",
			amount:      "1",
			decimals:    18,
			expected:    "1000000000000000000", // 1 * 10^18
			description: "1 ETH should be 1000000000000000000 wei",
		},
		{
			name:        "eth_with_decimals",
			amount:      "1.5",
			decimals:    18,
			expected:    "1500000000000000000", // 1.5 * 10^18
			description: "1.5 ETH should be 1500000000000000000 wei",
		},
		{
			name:        "eth_gwei",
			amount:      "0.000000001",
			decimals:    18,
			expected:    "1000000000", // 1 gwei = 10^9 wei
			description: "1 gwei (0.000000001 ETH) should be 1000000000 wei",
		},
		{
			name:        "eth_max_precision",
			amount:      "1.123456789012345678",
			decimals:    18,
			expected:    "1123456789012345678", // 1.123456789012345678 * 10^18
			description: "ETH with 18 decimals should preserve full precision",
		},
		{
			name:        "zero_amount",
			amount:      "0",
			decimals:    6,
			expected:    "0",
			description: "Zero amount should be zero",
		},
		{
			name:        "large_amount",
			amount:      "1000000",
			decimals:    6,
			expected:    "1000000000000", // 1000000 * 10^6
			description: "Large amount should be handled correctly",
		},
		{
			name:        "btc_like_8_decimals",
			amount:      "0.00000001",
			decimals:    8,
			expected:    "1", // 0.00000001 * 10^8 = 1 satoshi
			description: "1 satoshi (0.00000001 BTC) should be 1",
		},
		{
			name:        "btc_like_full_amount",
			amount:      "21.12345678",
			decimals:    8,
			expected:    "2112345678", // 21.12345678 * 10^8
			description: "BTC amount with 8 decimals should convert correctly",
		},
		{
			name:        "two_decimals_currency",
			amount:      "99.99",
			decimals:    2,
			expected:    "9999", // 99.99 * 10^2
			description: "Currency with 2 decimals (like cents) should convert correctly",
		},
		{
			name:        "zero_decimals",
			amount:      "100",
			decimals:    0,
			expected:    "100", // 100 * 10^0 = 100
			description: "Token with 0 decimals should remain unchanged",
		},
		{
			name:        "fractional_less_than_decimals",
			amount:      "1.1",
			decimals:    6,
			expected:    "1100000", // 1.1 * 10^6
			description: "Amount with fewer decimals than max should be scaled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := decimal.NewFromString(tt.amount)
			assert.NoError(t, err, "Test setup error: invalid amount string")

			result, err := DecimalToBigInt(amount, tt.decimals)
			assert.NoError(t, err)

			expected, ok := new(big.Int).SetString(tt.expected, 10)
			assert.True(t, ok, "Test setup error: invalid expected value")

			assert.Equal(t, expected.String(), result.String(), tt.description)
		})
	}
}

func TestDecimalToBigInt_NegativeAmounts(t *testing.T) {
	t.Run("negative_usdc", func(t *testing.T) {
		amount := decimal.NewFromFloat(-1.23)
		result, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		expected := big.NewInt(-1230000)
		assert.Equal(t, expected.String(), result.String(), "Negative amounts should be handled correctly")
	})

	t.Run("negative_eth", func(t *testing.T) {
		amount := decimal.NewFromFloat(-0.5)
		result, err := DecimalToBigInt(amount, 18)
		assert.NoError(t, err)
		expected, _ := new(big.Int).SetString("-500000000000000000", 10)
		assert.Equal(t, expected.String(), result.String(), "Negative ETH amount should convert correctly")
	})

	t.Run("negative_zero", func(t *testing.T) {
		amount := decimal.NewFromInt(0)
		result, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		expected := big.NewInt(0)
		assert.Equal(t, expected.String(), result.String(), "Zero should always be zero")
	})
}

func TestDecimalToBigInt_EdgeCases(t *testing.T) {
	t.Run("very_large_amount", func(t *testing.T) {
		// Test with a very large amount
		amount, err := decimal.NewFromString("999999999999999999.123456")
		assert.NoError(t, err)
		result, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		// 999999999999999999.123456 * 10^6 = 999999999999999999123456
		expected, ok := new(big.Int).SetString("999999999999999999123456", 10)
		assert.True(t, ok)
		assert.Equal(t, expected.String(), result.String(), "Very large amounts should be handled")
	})

	t.Run("very_small_amount", func(t *testing.T) {
		// Test with an amount that has more decimals than supported
		amount := decimal.NewFromFloat(0.0000001) // 7 decimals
		_, err := DecimalToBigInt(amount, 6)      // Only 6 decimal precision
		// This should return an error because the amount has more decimal places than allowed
		// After scaling: 0.0000001 * 10^6 = 0.1, which still has a fractional part
		assert.Error(t, err, "Amount with more decimals than supported should return an error")
		assert.Contains(t, err.Error(), "precision", "Error should mention precision")
	})

	t.Run("max_uint8_decimals", func(t *testing.T) {
		// Test with maximum uint8 value for decimals (not practical, but edge case)
		amount := decimal.NewFromInt(1)
		result, err := DecimalToBigInt(amount, 255)
		assert.NoError(t, err)
		// 1 * 10^255 should work
		expected := new(big.Int).Exp(big.NewInt(10), big.NewInt(255), nil)
		assert.Equal(t, expected.String(), result.String(), "Maximum decimals should work")
	})

	t.Run("precision_preservation", func(t *testing.T) {
		// Test that we don't lose precision during conversion
		amount, err := decimal.NewFromString("123.456789")
		assert.NoError(t, err)
		result, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		// 123.456789 * 10^6 = 123456789
		expected := big.NewInt(123456789)
		assert.Equal(t, expected.String(), result.String(), "Precision should be preserved")
	})
}

func TestDecimalToBigInt_RoundTrip(t *testing.T) {
	t.Run("usdc_round_trip", func(t *testing.T) {
		// Test that we can convert back and forth without losing precision
		original := "1.123456"
		amount, err := decimal.NewFromString(original)
		assert.NoError(t, err)

		// Convert to big.Int
		bigIntValue, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		// Convert back to decimal
		divisor := decimal.New(1, 6) // 10^6
		recovered := decimal.NewFromBigInt(bigIntValue, 0).Div(divisor)

		assert.Equal(t, original, recovered.String(), "Round trip conversion should preserve value")
	})
}
