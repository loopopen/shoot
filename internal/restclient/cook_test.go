package restclient

import (
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
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

func TestParseRestTag(t *testing.T) {
	tests := []struct {
		tag     string
		want    restTag
		wantErr string
	}{
		{tag: `rest:"-"`, want: restTag{Ignored: true}},
		{tag: "`rest:\"-\"`", want: restTag{Ignored: true}},
		{tag: `rest:"alias=lang,omitempty"`, want: restTag{Alias: "lang", OmitEmpty: true}},
		{tag: `rest:"alias=page_size"`, want: restTag{Alias: "page_size"}},
		{tag: `rest:"alias=pub_date,dateformat=2006-01-02,omitempty"`, want: restTag{Alias: "pub_date", DateFormat: "2006-01-02", OmitEmpty: true}},
		{tag: `rest:"unixmilli,utc"`, want: restTag{UnixMilli: true, UTC: true}},
		{tag: `rest:"csv"`, want: restTag{CSV: true}},
		{tag: `json:"id"`},
		{tag: ""},
		{tag: `rest:"unknown=1"`, wantErr: `unknown rest option "unknown"`},
		{tag: `rest:"dateformat=2006-01-02,unix"`, wantErr: "cannot be combined"},
		{tag: `rest:"utc"`, wantErr: "utc requires"},
		{tag: `rest:"-,omitempty"`, wantErr: "cannot be combined"},
		{tag: `rest:"omitempty=true"`, wantErr: "does not take a value"},
	}
	for _, tt := range tests {
		got, err := parseRestTag(tt.tag)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("parseRestTag(%q) error = %v, want %q", tt.tag, err, tt.wantErr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseRestTag(%q): %v", tt.tag, err)
		}
		if got != tt.want {
			t.Fatalf("parseRestTag(%q) = %+v, want %+v", tt.tag, got, tt.want)
		}
	}
}

func TestHandleStructSkipsRestDash(t *testing.T) {
	pkg := types.NewPackage("example.com/books", "books")
	st := types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, pkg, "PageSize", types.Typ[types.Int], false),
		types.NewField(token.NoPos, pkg, "Ignore", types.Typ[types.String], false),
		types.NewField(token.NoPos, pkg, "language", types.Typ[types.String], false),
	}, []string{
		`rest:"alias=page_size"`,
		`rest:"-"`,
		`rest:"alias=lang,omitempty"`,
	})

	g := &Generator{data: NewTmplData("test", "test")}
	g.data.QueryParamsMap = map[string][]string{}
	g.data.IsParamPtrMap = map[string]map[string]bool{}
	g.data.AliasMap = map[string]map[string]string{}
	g.data.QueryOptMap = map[string]map[string]string{}
	if err := g.handleStruct(st, ast.NewIdent("req"), "QueryBooks"); err != nil {
		t.Fatal(err)
	}

	got := g.data.QueryParamsMap["QueryBooks"]
	want := []string{"req.PageSize", "req.Language()"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("query params = %#v, want %#v", got, want)
	}
	if alias := g.data.AliasMap["QueryBooks"]["req.PageSize"]; alias != "page_size" {
		t.Fatalf("PageSize alias = %q", alias)
	}
	if alias := g.data.AliasMap["QueryBooks"]["req.Language()"]; alias != "lang" {
		t.Fatalf("language alias = %q", alias)
	}
	if _, ok := g.data.AliasMap["QueryBooks"]["req.Ignore"]; ok {
		t.Fatal("ignored field was given a query alias")
	}
	if lit := g.data.QueryOptMap["QueryBooks"]["req.Language()"]; lit != "shoot.QueryOpt{OmitEmpty: true}" {
		t.Fatalf("language encoding = %q", lit)
	}
	if _, ok := g.data.QueryOptMap["QueryBooks"]["req.PageSize"]; ok {
		t.Fatal("plain field was given a query encoding")
	}
}

func TestHandleStructQueryEncoding(t *testing.T) {
	pkg := types.NewPackage("example.com/books", "books")
	timeType := timeNamedType()
	st := types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, pkg, "PubDate", timeType, false),
		types.NewField(token.NoPos, pkg, "IDs", types.NewSlice(types.Typ[types.String]), false),
		types.NewField(token.NoPos, pkg, "When", types.NewPointer(timeType), false),
	}, []string{
		`rest:"dateformat=2006-01-02,omitempty"`,
		`rest:"csv"`,
		`rest:"unix,utc"`,
	})
	g := newQueryGenerator()
	if err := g.handleStruct(st, ast.NewIdent("req"), "Search"); err != nil {
		t.Fatal(err)
	}
	got := g.data.QueryOptMap["Search"]
	want := map[string]string{
		"req.PubDate": `shoot.QueryOpt{DateFormat: "2006-01-02", OmitEmpty: true}`,
		"req.IDs":     "shoot.QueryOpt{CSV: true}",
		"req.When":    "shoot.QueryOpt{Unix: true, UTC: true}",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("encoding = %#v, want %#v", got, want)
	}

	bad := []struct {
		name string
		typ  types.Type
		tag  string
	}{
		{name: "date on string", typ: types.Typ[types.String], tag: `rest:"dateformat=2006-01-02"`},
		{name: "csv on int", typ: types.Typ[types.Int], tag: `rest:"csv"`},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			st := types.NewStruct([]*types.Var{
				types.NewField(token.NoPos, pkg, "Field", tt.typ, false),
			}, []string{tt.tag})
			g := newQueryGenerator()
			if err := g.handleStruct(st, ast.NewIdent("req"), "Search"); err == nil {
				t.Fatal("expected a type error")
			}
		})
	}
}

func TestTemplateQueryEncoding(t *testing.T) {
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

	data := NewTmplData("shoot rest -type=Client", "test")
	data.SetTypeName("Client")
	data.SetPackageName("client")
	data.MethodList = []string{"Search"}
	data.SigMap = map[string]string{"Search": "Search(ctx context.Context, req SearchReq) (*shoot.Response, error)"}
	data.HTTPMethodMap = map[string]string{"Search": "GET"}
	data.PathMap = map[string]string{"Search": "/search"}
	data.BodyHTTPMethods = []string{"POST", "PUT", "PATCH"}
	data.QueryParamsMap = map[string][]string{"Search": {"req.Language()", "req.PageSize", "pageIdx", "req.IDs"}}
	data.AliasMap = map[string]map[string]string{"Search": {
		"req.Language()": "lang",
		"req.PageSize":   "page_size",
		"pageIdx":        "page_idx",
		"req.IDs":        "id",
	}}
	data.IsParamPtrMap = map[string]map[string]bool{"Search": {"pageIdx": true}}
	data.QueryOptMap = map[string]map[string]string{"Search": {
		"req.Language()": "shoot.QueryOpt{OmitEmpty: true}",
		"req.IDs":        "shoot.QueryOpt{CSV: true}",
	}}
	data.ErrReturnMap = map[string]string{"Search": "nil, err"}
	data.DefaultHeaders = map[string]map[string]string{}
	data.StaticHeaders = map[string]string{}
	data.MethodHeadersMap = map[string]map[string]string{}
	data.ResponseMap = map[string]bool{"Search": true}
	data.CtxParamMap = map[string]string{"Search": "ctx"}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	src := buf.String()
	assertNoBlankLinesInFunctions(t, src)
	body := methodSrc(t, src, "Search")
	for _, line := range []string{
		`if err := shoot.SetQuery(query_, "lang", req.Language(), shoot.QueryOpt{OmitEmpty: true}); err != nil {`,
		`return nil, err`,
		`query_.Set("page_size", fmt.Sprintf("%v",req.PageSize))`,
		`if pageIdx != nil {`,
		`query_.Set("page_idx", fmt.Sprintf("%v", *pageIdx))`,
		`if err := shoot.SetQuery(query_, "id", req.IDs, shoot.QueryOpt{CSV: true}); err != nil {`,
	} {
		if !strings.Contains(body, line) {
			t.Fatalf("Search missing %s\n%s", line, body)
		}
	}
}

func newQueryGenerator() *Generator {
	g := &Generator{data: NewTmplData("test", "test")}
	g.data.QueryParamsMap = map[string][]string{}
	g.data.IsParamPtrMap = map[string]map[string]bool{}
	g.data.AliasMap = map[string]map[string]string{}
	g.data.QueryOptMap = map[string]map[string]string{}
	return g
}

func timeNamedType() types.Type {
	pkg := types.NewPackage("time", "time")
	obj := types.NewTypeName(token.NoPos, pkg, "Time", nil)
	return types.NewNamed(obj, types.NewStruct(nil, nil), nil)
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
