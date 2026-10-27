package shoot

import (
	"reflect"
	"strings"
	"testing"
)

func TestWithoutGeneratedType(t *testing.T) {
	src := []byte(`package p

import "encoding/json"

func NewUser(id string) *User { return &User{id: id} }
func (u *User) Id() string { return u.id }
type UserGetter interface{ Id() string }
func NewOrder(id string) *Order { return &Order{id: id} }
func (o *Order) Id() string { return o.id }
func (o Order) MarshalJSON() ([]byte, error) { return json.Marshal(o.Id()) }
`)
	got, err := withoutGeneratedType(src, "User")
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if strings.Contains(text, "NewUser") || strings.Contains(text, "UserGetter") {
		t.Fatalf("user decls still present:\n%s", text)
	}
	if !strings.Contains(text, "NewOrder") || !strings.Contains(text, "json.Marshal") {
		t.Fatalf("order decls missing:\n%s", text)
	}
}

func Test_noopFix(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "oneline",
			args: args{
				src: []byte(`func Func() {}`),
			},
			want: []byte(`func Func() { /*noop*/ }`),
		},
		{
			name: "newline",
			args: args{
				src: []byte(`func Func() {
				}`),
			},
			want: []byte(`func Func() { /*noop*/ }`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := noopFix(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("noopFix() = %v, want %v", got, tt.want)
			}
		})
	}
}
