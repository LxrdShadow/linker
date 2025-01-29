package util

import "testing"

type TestCase struct {
	num   uint64
	unit  string
	denom uint64
}

func TestByteDecodeUnit(t *testing.T) {
	cases := []TestCase{
		{
			num:   uint64(200),
			unit:  "B",
			denom: uint64(1),
		},
		{
			num:   uint64(2000),
			unit:  "KB",
			denom: uint64(1000),
		},
		{
			num:   uint64(200000),
			unit:  "KB",
			denom: uint64(1000),
		},
		{
			num:   uint64(2000000),
			unit:  "MB",
			denom: uint64(1000 * 1000),
		},
		{
			num:   uint64(2000000000),
			unit:  "GB",
			denom: uint64(1000 * 1000 * 1000),
		},
		{
			num:   uint64(2000000000000),
			unit:  "TB",
			denom: uint64(1000 * 1000 * 1000 * 1000),
		},
		{
			num:   uint64(2000000000000000),
			unit:  "PB",
			denom: uint64(1000 * 1000 * 1000 * 1000 * 1000),
		},
	}

	t.Run("returns the correct unit and denominator", func(t *testing.T) {
		for _, test := range cases {
			unit, denom := ByteDecodeUnit(test.num)

			if unit != test.unit {
				t.Errorf("unit mismatch: got %s want %s", unit, test.unit)
			}

			if denom != test.denom {
				t.Errorf("denominator mismatch: got %d want %d", denom, test.denom)
			}
		}
	})
}
