package converter

import "testing"

func TestConvertAmountFloatToInt(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   int64
	}{
		{
			name:   "Success",
			amount: 1.23,
			want:   123,
		},
		{
			name:   "Zero",
			amount: 0,
			want:   0,
		},
		{
			name:   "Negative",
			amount: -1.23,
			want:   -123,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ConvertAmountFloatToInt(tc.amount)
			if got != tc.want {
				t.Errorf("unexpected result, want: %d, got: %d", tc.want, got)
			}
		})
	}
}

func TestConvertAmountIntToFloat(t *testing.T) {
	cases := []struct {
		name   string
		amount int64
		want   float64
	}{
		{
			name:   "Success",
			amount: 123,
			want:   1.23,
		},
		{
			name:   "Zero",
			amount: 0,
			want:   0,
		},
		{
			name:   "Negative",
			amount: -123,
			want:   -1.23,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ConvertAmountIntToFloat(tc.amount)
			if got != tc.want {
				t.Errorf("unexpected result, want: %f, got: %f", tc.want, got)
			}
		})
	}
}
