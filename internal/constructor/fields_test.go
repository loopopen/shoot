package constructor

import (
	"testing"
)

func Test_newBody(t *testing.T) {
	type args struct {
		fields  []*Field
		nameMap map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				fields: []*Field{
					{
						name:          "f1",
						qualifiedType: "int",
					},
					{
						name:          "f2",
						qualifiedType: "int",
					},
				},
				nameMap: map[string]string{
					"f1": "f1",
					"f2": "f2",
				},
			},
			want: "f1: f1,\n" +
				"f2: f2,\n",
		},
		{
			name: "with embeded",
			args: args{
				fields: []*Field{
					{
						name:          "f1",
						qualifiedType: "int",
						depth:         0,
					},
					{
						name:          "Embed",
						qualifiedType: "x.Embed",
						depth:         0,
						isEmbeded:     true,
						isPtr:         true,
					},
					{
						name:          "ef1",
						qualifiedType: "int",
						depth:         1,
					},
					{
						name:          "f2",
						qualifiedType: "int",
						depth:         0,
					},
				},
				nameMap: map[string]string{
					"f1":  "f1",
					"ef1": "ef1",
					"f2":  "f2",
				},
			},
			want: "f1: f1,\n" +
				"Embed: &x.Embed{\n" +
				"ef1: ef1,\n" +
				"},\n" +
				"f2: f2,\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newBody(tt.args.fields, tt.args.nameMap); got != tt.want {
				t.Errorf("newBody() = %v, want %v", got, tt.want)
			}
		})
	}
}
