package mpc

import (
	"math"
	"math/big"
	"testing"
)

func TestGenerateShares(t *testing.T) {
	type args struct {
		secret  int64
		nShares int
		modulus *big.Int
	}
	tests := []struct {
		name string
		args args
	}{
		{"small M", args{4, 4, new(big.Int).SetInt64(45)}},
		{"mediam M", args{7658, 3, new(big.Int).SetInt64(6745985156)}},
		{"default M", args{18, 9, new(big.Int).SetInt64(math.MaxInt64)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateShares(tt.args.secret, tt.args.nShares, tt.args.modulus)

			if len(got) != int(tt.args.nShares) {
				t.Errorf("GenerateShares() = wrong number of shares %v, want %v", len(got), tt.args.nShares)
			}
			sum := new(big.Int)
			for _, el := range got {
				sum.Add(sum, el)
			}
			sum.Mod(sum, tt.args.modulus)
			val := new(big.Int).SetInt64(tt.args.secret)
			if sum.Cmp(val) != 0 {
				t.Errorf("GenerateShares() = %v, want %v", sum, val)
			}
		})
	}
}

func TestGenerateBeaverTriplet(t *testing.T) {
	tests := []struct {
		name string
		N    *big.Int
	}{
		{"valid values", new(big.Int).SetInt64(45)},
		{"valid values", new(big.Int).SetInt64(45985156)},
		{"valid values", new(big.Int).SetInt64(1459886875156)},
		{"valid values", new(big.Int).SetInt64(math.MaxInt64)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateBeaverTriplet(tt.N)

			prod := new(big.Int).Mul(got[1], got[2])
			prod.Mod(prod, tt.N)
			if prod.Cmp(got[0]) != 0 {
				t.Errorf("GenerateBeaverTriplet() = %v, want %v", got[0], prod)
			}
		})
	}
}
