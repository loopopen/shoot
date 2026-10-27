package shoot

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// QueryOpt controls how one value is encoded into a query string.
// Zero options use the default encoding: scalars are printed, and slices
// repeat the key once per element.
type QueryOpt struct {
	DateFormat string
	Unix       bool
	UnixMilli  bool
	UTC        bool
	CSV        bool
	OmitEmpty  bool
}

// SetQuery writes value into q under key.
// Nil pointers are skipped. OmitEmpty also skips zero values.
// time.Time uses DateFormat, Unix, or UnixMilli when set; UTC converts the
// time before that formatting. Slices and arrays repeat the key, unless CSV
// joins them with commas.
func SetQuery(q url.Values, key string, value any, opt QueryOpt) error {
	if q == nil {
		return fmt.Errorf("query values is nil")
	}
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil
	}
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if opt.OmitEmpty && queryZero(v) {
		return nil
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return setQueryList(q, key, v, opt)
	}
	if opt.CSV {
		return fmt.Errorf("csv requires a slice or array")
	}
	s, skip, err := formatQueryValue(v, opt)
	if err != nil || skip {
		return err
	}
	q.Set(key, s)
	return nil
}

func setQueryList(q url.Values, key string, v reflect.Value, opt QueryOpt) error {
	elemOpt := opt
	elemOpt.OmitEmpty = false
	elemOpt.CSV = false
	var parts []string
	for i := 0; i < v.Len(); i++ {
		s, skip, err := formatQueryValue(v.Index(i), elemOpt)
		if err != nil {
			return err
		}
		if skip {
			continue
		}
		parts = append(parts, s)
	}
	if opt.CSV {
		q.Set(key, strings.Join(parts, ","))
		return nil
	}
	for _, part := range parts {
		q.Add(key, part)
	}
	return nil
}

func formatQueryValue(v reflect.Value, opt QueryOpt) (string, bool, error) {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "", true, nil
		}
		v = v.Elem()
	}
	if t, ok := v.Interface().(time.Time); ok {
		return formatQueryTime(t, opt), false, nil
	}
	if opt.DateFormat != "" || opt.Unix || opt.UnixMilli || opt.UTC {
		return "", false, fmt.Errorf("dateformat, unix, unixmilli, and utc require time.Time")
	}
	return fmt.Sprint(v.Interface()), false, nil
}

func formatQueryTime(t time.Time, opt QueryOpt) string {
	if opt.UTC {
		t = t.UTC()
	}
	switch {
	case opt.DateFormat != "":
		return t.Format(opt.DateFormat)
	case opt.Unix:
		return strconv.FormatInt(t.Unix(), 10)
	case opt.UnixMilli:
		return strconv.FormatInt(t.UnixMilli(), 10)
	default:
		return fmt.Sprint(t)
	}
}

func queryZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	if t, ok := v.Interface().(time.Time); ok {
		return t.IsZero()
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	default:
		return v.IsZero()
	}
}
