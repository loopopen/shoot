package restclient

import (
	"reflect"
	"testing"

	"github.com/loopopen/shoot/internal/shoot"
)

func Test_parsePath(t *testing.T) {
	type args struct {
		doc string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 []string
		want3 bool
	}{
		{
			name: "normal case with quotes",
			args: args{
				doc: shoot.Sign + `Get("/api/users{id}")`,
			},
			want:  "GET",
			want1: "/api/users{id}",
			want2: []string{"id"},
			want3: true,
		},
		{
			name: "normal case without quotes",
			args: args{
				doc: shoot.Sign + `Get(/api/users{id})`,
			},
			want:  "GET",
			want1: "/api/users{id}",
			want2: []string{"id"},
			want3: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3 := parsePath(tt.args.doc, shoot.Sign)
			if got != tt.want {
				t.Errorf("parsePath() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parsePath() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("parsePath() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.want3 {
				t.Errorf("parsePath() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

func Test_parseHeaders(t *testing.T) {
	type args struct {
		doc string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "multi-line",
			args: args{
				doc: shoot.Sign + `headers=
 {Content-Type:application/json},
 {Accept:application/json},`,
			},
			want: map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
		},
		{
			name: "one-line",
			args: args{
				doc: shoot.Sign + `headers={A:a},{B:b}`,
			},
			want: map[string]string{
				"A": "a",
				"B": "b",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHeaders(tt.args.doc, shoot.Sign)
			for k, v := range tt.want {
				if v != got[k] {
					t.Errorf("value of %s got = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}

func TestSetBodyParamNameRejectsAmbiguousBinding(t *testing.T) {
	g := &Generator{data: NewTmplData("test", "test")}
	g.data.BodyParamMap = make(map[string]string)

	if err := g.setBodyParamName("Update", "payload"); err != nil {
		t.Fatalf("first body parameter returned error: %v", err)
	}
	if err := g.setBodyParamName("Update", "options"); err == nil {
		t.Fatal("second body parameter did not return an ambiguity error")
	}
}
