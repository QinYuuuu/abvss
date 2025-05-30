package pedersen

import (
	"reflect"
	"testing"

	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
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
			param := &Param{
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

func TestCommitter_Verify(t *testing.T) {
	param := Setup(edwards25519.NewBlakeSHA256Ed25519())
	type args struct {
		m  []kyber.Scalar
		c  *Comm
		pi *Pi
	}
	type testCase struct {
		name string
		args args
		want bool
	}
	testCases := []testCase{
		// TODO: Add test cases.
		{
			name: "test-verify-success",
			args: args{
				m: []kyber.Scalar{
					param.group.Scalar().SetInt64(1),
					param.group.Scalar().SetInt64(2),
					param.group.Scalar().SetInt64(3),
				},
				c:  &Comm{},
				pi: &Pi{},
			},
			want: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.c, tt.args.pi = param.Commit(tt.args.m)
			if got := param.Verify(tt.args.m, tt.args.c, tt.args.pi); got != tt.want {
				t.Errorf("Committer.Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
