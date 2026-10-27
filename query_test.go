package shoot

import (
	"net/url"
	"testing"
	"time"
)

func TestSetQueryEncoding(t *testing.T) {
	when := time.Date(2024, 3, 8, 1, 2, 3, 0, time.FixedZone("CST", 8*3600))

	tests := []struct {
		name  string
		value any
		opt   QueryOpt
		want  url.Values
	}{
		{name: "string", value: "go", want: url.Values{"q": {"go"}}},
		{name: "empty omitempty", value: "", opt: QueryOpt{OmitEmpty: true}, want: url.Values{}},
		{name: "zero int omitempty", value: 0, opt: QueryOpt{OmitEmpty: true}, want: url.Values{}},
		{name: "false omitempty", value: false, opt: QueryOpt{OmitEmpty: true}, want: url.Values{}},
		{name: "kept zero", value: 0, want: url.Values{"q": {"0"}}},
		{
			name:  "dateformat",
			value: when,
			opt:   QueryOpt{DateFormat: "2006-01-02"},
			want:  url.Values{"q": {"2024-03-08"}},
		},
		{
			name:  "dateformat utc",
			value: when,
			opt:   QueryOpt{DateFormat: "2006-01-02 15", UTC: true},
			want:  url.Values{"q": {"2024-03-07 17"}},
		},
		{
			name:  "zero time",
			value: time.Time{},
			opt:   QueryOpt{DateFormat: "2006-01-02"},
			want:  url.Values{"q": {"0001-01-01"}},
		},
		{
			name:  "zero time omitempty",
			value: time.Time{},
			opt:   QueryOpt{DateFormat: "2006-01-02", OmitEmpty: true},
			want:  url.Values{},
		},
		{
			name:  "unix",
			value: when,
			opt:   QueryOpt{Unix: true},
			want:  url.Values{"q": {"1709830923"}},
		},
		{
			name:  "unixmilli",
			value: when,
			opt:   QueryOpt{UnixMilli: true, UTC: true},
			want:  url.Values{"q": {"1709830923000"}},
		},
		{
			name:  "repeat",
			value: []string{"a", "b"},
			want:  url.Values{"q": {"a", "b"}},
		},
		{
			name:  "csv",
			value: []string{"a", "b"},
			opt:   QueryOpt{CSV: true},
			want:  url.Values{"q": {"a,b"}},
		},
		{
			name:  "empty slice omitempty",
			value: []string{},
			opt:   QueryOpt{OmitEmpty: true},
			want:  url.Values{},
		},
		{
			name:  "time slice",
			value: []time.Time{when},
			opt:   QueryOpt{DateFormat: "2006-01-02", CSV: true},
			want:  url.Values{"q": {"2024-03-08"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := url.Values{}
			if err := SetQuery(q, "q", tt.value, tt.opt); err != nil {
				t.Fatal(err)
			}
			if q.Encode() != tt.want.Encode() {
				t.Fatalf("query = %v, want %v", q, tt.want)
			}
		})
	}
}

func TestSetQuerySkipsNilPointers(t *testing.T) {
	q := url.Values{}
	var s *string
	if err := SetQuery(q, "q", s, QueryOpt{}); err != nil {
		t.Fatal(err)
	}
	text := "go"
	if err := SetQuery(q, "name", &text, QueryOpt{}); err != nil {
		t.Fatal(err)
	}
	when := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if err := SetQuery(q, "day", &when, QueryOpt{DateFormat: "2006-01-02"}); err != nil {
		t.Fatal(err)
	}
	items := []*string{nil, &text}
	if err := SetQuery(q, "item", items, QueryOpt{}); err != nil {
		t.Fatal(err)
	}
	got := q.Encode()
	want := url.Values{"name": {"go"}, "day": {"2024-01-02"}, "item": {"go"}}.Encode()
	if got != want {
		t.Fatalf("query = %s, want %s", got, want)
	}
}
