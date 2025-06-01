package pedersen

import (
	"reflect"
	"testing"

	"go.dedis.ch/kyber/v4"
)

func TestVectorPCommit(t *testing.T) {
	type args struct {
		param VectorParam
		value []kyber.Scalar
	}
	var tests []struct {
		name string
		args args
		want kyber.Point
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.param.Commit(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VectorPCommit() = %v, want %v", got, tt.want)
			}
		})
	}
}
