package mpc

import (
	"testing"
)

func TestDistributedIntMean_Setup(t *testing.T) {
	testParams := NewParameters(512)
	tests := []struct {
		name    string
		param   *Parameters
		args    []int64
		wantErr bool
	}{
		{
			"One test is enougth",
			testParams,
			[]int64{1, 2, 3, 4},
			false,
		},
		{
			"But I'll write two, if that makes you happy",
			&Parameters{},
			[]int64{1, 2, 3, 4, 5},
			true,
		},
		{
			"Or, maybe, even three",
			testParams,
			[]int64{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := NewDistributedIntMean()

			if err := dm.Setup(tt.param, tt.args); (err != nil) != tt.wantErr {
				t.Errorf("DistributedMean.Setup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDistributedIntMean_Run(t *testing.T) {
	testParams := NewParameters(128)
	tests := []struct {
		name    string
		params  *Parameters
		args    []int64
		want    int64
		wantErr bool
	}{
		{
			"The test of our lives",
			testParams,
			[]int64{1, 3},
			int64(2),
			false,
		},
		{
			"The test of our lives",
			testParams,
			[]int64{1, 2, 3, 5, 7},
			int64(2),
			false,
		},
		{
			"The test of our lives",
			testParams,
			[]int64{4, 14, 21},
			int64(2),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := NewDistributedIntMean()
			dm.Setup(tt.params, tt.args)
			if err := dm.Run(); (err != nil) != tt.wantErr {
				t.Errorf("DistributedMean.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
