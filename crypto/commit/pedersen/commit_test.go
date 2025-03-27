package pedersen

import (
	"reflect"
	"testing"

	"go.dedis.ch/kyber/v3"
)

func TestCommitter_Commit(t *testing.T) {
	type fields struct {
		group kyber.Group
		g     kyber.Point
		h     kyber.Point
	}
	type args struct {
		m []kyber.Scalar
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Comm
		want1  *Pi
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := &Committer{
				group: tt.fields.group,
				g:     tt.fields.g,
				h:     tt.fields.h,
			}
			got, got1 := param.Commit(tt.args.m)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Committer.Commit() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Committer.Commit() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
