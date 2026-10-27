package restclient

import (
	"go/format"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/loopopen/shoot/internal/shoot"
	"github.com/loopopen/shoot/internal/transfer"
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

func TestTemplateDefaultHeaders(t *testing.T) {
	tmpl, err := template.New("rest").Funcs(template.FuncMap{
		"camelCase":  transfer.ToCamelCaseGO,
		"pascalCase": transfer.ToPascalCase,
		"firstLower": transfer.FirstLowerLetter,
		"in": func(s string, list []string) bool {
			for _, x := range list {
				if s == x {
					return true
				}
			}
			return false
		},
		"cond": func(cond bool, a, b string) string {
			if cond {
				return a
			}
			return b
		},
	}).Parse(tmplTxt)
	if err != nil {
		t.Fatal(err)
	}

	data := NewTmplData("shoot rest -type=Client", "test")
	data.SetTypeName("Client")
	data.SetPackageName("client")
	data.MethodList = []string{"Get", "Set", "Download"}
	data.SigMap = map[string]string{
		"Get":      "Get(ctx context.Context) (*KV, *shoot.Response, error)",
		"Set":      "Set(ctx context.Context, kv *KV) (*shoot.Response, error)",
		"Download": "Download(ctx context.Context) (*shoot.StreamResponse, error)",
	}
	data.HTTPMethodMap = map[string]string{"Get": "GET", "Set": "POST", "Download": "GET"}
	data.PathMap = map[string]string{"Get": "/get", "Set": "/set", "Download": "/download"}
	data.BodyHTTPMethods = []string{"POST", "PUT", "PATCH"}
	data.BodyParamMap = map[string]string{"Set": "kv"}
	data.ReturnResultMap = map[string]struct {
		Type  string
		IsPtr bool
	}{"Get": {Type: "*KV", IsPtr: true}}
	data.DefaultHeaders = map[string]map[string]string{
		"Get": methodDefaultHeaders(true, false),
		"Set": methodDefaultHeaders(false, true),
	}
	data.ResponseMap = map[string]bool{"Get": true, "Set": true}
	data.StreamResponseMap = map[string]bool{"Download": true}
	data.ErrReturnMap = map[string]string{
		"Get":      "nil, nil, err",
		"Set":      "nil, err",
		"Download": "nil, err",
	}
	data.CtxParamMap = map[string]string{"Get": "ctx", "Set": "ctx", "Download": "ctx"}
	data.MethodHeadersMap = map[string]map[string]string{
		"Download": {"Content-Type": "application/zip"},
	}
	data.StaticHeaders = map[string]string{}
	data.AliasMap = map[string]map[string]string{}
	data.PathParamsMap = map[string][]string{}
	data.QueryParamsMap = map[string][]string{}
	data.IsParamPtrMap = map[string]map[string]bool{}
	data.QueryDictMap = map[string]string{}
	data.RequestOptionsMap = map[string]string{}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	src := buf.String()
	get := methodSrc(t, src, "Get")
	set := methodSrc(t, src, "Set")
	download := methodSrc(t, src, "Download")
	if !strings.Contains(get, `req_.Header.Set("Accept", "application/json")`) {
		t.Fatalf("Get missing Accept:\n%s", get)
	}
	if strings.Contains(set, `req_.Header.Set("Accept", "application/json")`) {
		t.Fatalf("Set has Accept:\n%s", set)
	}
	if !strings.Contains(set, `req_.Header.Set("Content-Type", "application/json")`) {
		t.Fatalf("Set missing Content-Type:\n%s", set)
	}
	if strings.Contains(download, `req_.Header.Set("Accept", "application/json")`) {
		t.Fatalf("Download has Accept:\n%s", download)
	}
	if !strings.Contains(download, `req_.Header.Set("Content-Type", "application/zip")`) {
		t.Fatalf("Download missing explicit Content-Type:\n%s", download)
	}
	assertNoBlankLinesInFunctions(t, src)
}

func TestTemplateFunctionsStayCompact(t *testing.T) {
	tmpl := template.Must(template.New("rest").Funcs(template.FuncMap{
		"camelCase":  transfer.ToCamelCaseGO,
		"pascalCase": transfer.ToPascalCase,
		"firstLower": transfer.FirstLowerLetter,
		"in": func(s string, list []string) bool {
			for _, x := range list {
				if s == x {
					return true
				}
			}
			return false
		},
		"cond": func(cond bool, a, b string) string {
			if cond {
				return a
			}
			return b
		},
	}).Parse(tmplTxt))

	result := func(typ string, ptr bool) struct {
		Type  string
		IsPtr bool
	} {
		return struct {
			Type  string
			IsPtr bool
		}{Type: typ, IsPtr: ptr}
	}

	data := NewTmplData("shoot rest -type=Client", "test")
	data.SetTypeName("Client")
	data.SetPackageName("client")
	data.MethodList = []string{"GetUser", "QueryDict", "Raw", "Values"}
	data.SigMap = map[string]string{
		"GetUser":   "GetUser(ctx context.Context, userID string, pageSize int, pageIdx *int) (*User, *shoot.Response, error)",
		"QueryDict": "QueryDict(ctx context.Context, params *map[string]string) (*QueryUsersResp, *shoot.Response, error)",
		"Raw":       "Raw(ctx context.Context) (*http.Response, error)",
		"Values":    "Values(ctx context.Context, req QueryReq) ([]User, *shoot.Response, error)",
	}
	data.HTTPMethodMap = map[string]string{
		"GetUser": "GET", "QueryDict": "GET", "Raw": "GET", "Values": "GET",
	}
	data.PathMap = map[string]string{
		"GetUser": "/users/{id}", "QueryDict": "/users", "Raw": "/raw", "Values": "/users",
	}
	data.BodyHTTPMethods = []string{"POST", "PUT", "PATCH"}
	data.PathParamsMap = map[string][]string{"GetUser": {"userID"}}
	data.AliasMap = map[string]map[string]string{
		"GetUser": {"userID": "id", "pageSize": "size", "pageIdx": "page_idx"},
	}
	data.QueryParamsMap = map[string][]string{"GetUser": {"pageSize", "pageIdx"}, "Values": {"req"}}
	data.IsParamPtrMap = map[string]map[string]bool{"GetUser": {"pageIdx": true}, "QueryDict": {"params": true}}
	data.QueryDictMap = map[string]string{"QueryDict": "params"}
	data.ReturnResultMap = map[string]struct {
		Type  string
		IsPtr bool
	}{
		"GetUser":   result("*User", true),
		"QueryDict": result("*QueryUsersResp", true),
		"Values":    result("[]User", false),
	}
	data.DefaultHeaders = map[string]map[string]string{
		"GetUser":   methodDefaultHeaders(true, false),
		"QueryDict": methodDefaultHeaders(true, false),
		"Values":    methodDefaultHeaders(true, false),
	}
	data.ResponseMap = map[string]bool{"GetUser": true, "QueryDict": true, "Values": true}
	data.ErrReturnMap = map[string]string{
		"GetUser": "nil, nil, err", "QueryDict": "nil, nil, err", "Raw": "nil, err", "Values": "nil, nil, err",
	}
	data.CtxParamMap = map[string]string{"GetUser": "ctx", "QueryDict": "ctx", "Raw": "ctx", "Values": "ctx"}
	data.StaticHeaders = map[string]string{"Tenant-Id": "123"}
	data.MethodHeadersMap = map[string]map[string]string{}
	data.RequestOptionsMap = map[string]string{}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	raw := buf.String()
	assertNoBlankLinesInFunctions(t, raw)
	if strings.Contains(raw, "}\n\n\nfunc") {
		t.Fatalf("more than one blank line between functions:\n%s", raw)
	}

	formatted, err := format.Source([]byte(raw))
	if err != nil {
		t.Fatalf("format: %v\n%s", err, raw)
	}
	src := string(formatted)
	assertNoBlankLinesInFunctions(t, src)

	getUser := methodSrc(t, src, "GetUser")
	for _, line := range []string{
		`path_ := "/users/{id}"`,
		`path_ = strings.Replace(path_, "{id}", url.PathEscape(fmt.Sprint(userID)), 1)`,
		`query_.Set("size", fmt.Sprintf("%v", pageSize))`,
		`if pageIdx != nil {`,
		`query_.Set("page_idx", fmt.Sprintf("%v", *pageIdx))`,
		`req_.Header.Set("Accept", "application/json")`,
		`req_.Header.Set("Tenant-Id", "123")`,
		`return &r_, response_, nil`,
	} {
		if !strings.Contains(getUser, line) {
			t.Fatalf("GetUser missing %s\n%s", line, getUser)
		}
	}
	if strings.Contains(getUser, "\n\n") {
		t.Fatalf("GetUser has a blank line:\n%s", getUser)
	}

	queryDict := methodSrc(t, src, "QueryDict")
	if !strings.Contains(queryDict, "if params != nil {\n\t\tfor k, v := range *params {") {
		t.Fatalf("QueryDict pointer map:\n%s", queryDict)
	}

	rawFn := methodSrc(t, src, "Raw")
	if !strings.Contains(rawFn, "resp_.Body = io.NopCloser(bytes.NewReader(responseBody_))") ||
		strings.Contains(rawFn, `Header.Set("Accept"`) {
		t.Fatalf("Raw response:\n%s", rawFn)
	}

	values := methodSrc(t, src, "Values")
	if !strings.Contains(values, "return r_, response_, nil") || strings.Contains(values, "return &r_") {
		t.Fatalf("non-pointer result:\n%s", values)
	}
}

func assertNoBlankLinesInFunctions(t *testing.T, src string) {
	t.Helper()
	depth := 0
	for i, line := range strings.Split(src, "\n") {
		if depth > 0 && strings.TrimSpace(line) == "" {
			t.Fatalf("blank line %d inside a function\n%s", i+1, src)
		}
		depth += strings.Count(line, "{") - strings.Count(line, "}")
	}
}

func methodSrc(t *testing.T, src, name string) string {
	t.Helper()
	start := strings.Index(src, "func (_c *client) "+name+"(")
	if start < 0 {
		t.Fatalf("missing %s", name)
	}
	rest := src[start+1:]
	next := strings.Index(rest, "\nfunc ")
	if next < 0 {
		return src[start:]
	}
	return src[start : start+1+next]
}

func TestMethodDefaultHeaders(t *testing.T) {
	if got := methodDefaultHeaders(true, false); got["Accept"] != "application/json" || got["Content-Type"] != "" {
		t.Fatalf("json result headers = %v", got)
	}
	if got := methodDefaultHeaders(false, true); got["Content-Type"] != "application/json" || got["Accept"] != "" {
		t.Fatalf("json body headers = %v", got)
	}
	if got := methodDefaultHeaders(true, true); got["Accept"] != "application/json" || got["Content-Type"] != "application/json" {
		t.Fatalf("json result and body headers = %v", got)
	}
	if got := methodDefaultHeaders(false, false); got != nil {
		t.Fatalf("raw response headers = %v", got)
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
