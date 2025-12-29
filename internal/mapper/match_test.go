package mapper

import (
	"go/types"
	"testing"
)

func Test_matchType(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		type1 types.Type
		type2 types.Type
		want  bool
		want2 bool
	}{
		{
			name:  "identical 1",
			type1: types.Typ[types.Int],
			type2: types.Typ[types.Int],
			want:  true,
			want2: true,
		},
		{
			name:  "convertible 1",
			type1: types.Typ[types.Int],
			type2: types.Typ[types.Int32],
			want:  false,
			want2: true,
		},
		{
			name:  "convertible 2",
			type1: types.Typ[types.String],
			type2: types.NewSlice(types.Typ[types.Byte]),
			want:  false,
			want2: true,
		},
		{
			name:  "not convertible 1",
			type1: types.Typ[types.Int64],
			type2: types.Typ[types.String],
			want:  false,
			want2: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := matchType(tt.type1, tt.type2)
			if got != tt.want {
				t.Errorf("matchType() = %v, want %v", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("matchType() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_smartMatch(t *testing.T) {
	type args struct {
		a string
		b string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "id=id",
			args: args{
				a: "id",
				b: "id",
			},
			want: true,
		},
		{
			name: "id~id+",
			args: args{
				a: "id",
				b: "id+",
			},
			want: false,
		},
		{
			name: "id~Id",
			args: args{
				a: "id",
				b: "Id",
			},
			want: false,
		},
		{
			name: "id~ID",
			args: args{
				a: "id",
				b: "ID",
			},
			want: true,
		},
		{
			name: "ID~id",
			args: args{
				a: "id",
				b: "ID",
			},
			want: true,
		},
		{
			name: "Id~ID",
			args: args{
				a: "id",
				b: "ID",
			},
			want: true,
		},
		{
			name: "LoadXml~LoadXML",
			args: args{
				a: "id",
				b: "ID",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := smartMatch(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("smartMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
