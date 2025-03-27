package pedersen

import (
	"reflect"
	"testing"

	"go.dedis.ch/kyber/v3"
)

func TestVectorPCommit(t *testing.T) {
	type args struct {
		param VectorParam
		value []kyber.Scalar
	}
	tests := []struct {
		name string
		args args
		want kyber.Point
	}{
		// TODO: Add test cases.

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VectorPCommit(tt.args.param, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VectorPCommit() = %v, want %v", got, tt.want)
			}
		})
	}
}
