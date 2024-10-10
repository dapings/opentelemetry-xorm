package logger

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dapings/opentelemetry-xorm/now"
)

type JSON json.RawMessage

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

type ExampleStruct struct {
	Name string
	Val  string
}

func (s ExampleStruct) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func format(v []byte, escaper string) string {
	return escaper + strings.ReplaceAll(string(v), escaper, escaper+escaper) + escaper
}

func TestFormat(t *testing.T) {
	var (
		jsVal = []byte(`{"Name":"test","Val":"test"}`)
		js    = JSON(jsVal)
		esVal = []byte(`{"Name":"test","Val":"test"}`)
		es    = ExampleStruct{Name: "test", Val: "test"}
	)

	result := format(jsVal, `"`)
	t.Logf("format []byte to string: %v, expected: %s", result, js)

	result = format(esVal, `"`)
	t.Logf("format []byte to string: %v, expected: %+v", result, es)
}

func TestExplainSQ(t *testing.T) {
	type (
		role      string
		password  []byte
		intType   int
		floatType float64
	)
	var (
		tt                 = now.MustParse("2024-09-26 17:45:15")
		myRole             = role("admin")
		pwd                = password("pass")
		jsVal              = []byte(`{"Name":"test","Val":"test"}`)
		js                 = JSON(jsVal)
		esVal              = []byte(`{"Name":"test","Val":"test"}`)
		es                 = ExampleStruct{Name: "test", Val: "test"}
		intVal   intType   = 1
		floatVal floatType = 1.23
	)

	results := []struct {
		SQL           string
		NumericRegexp *regexp.Regexp
		Vars          []any
		Result        string
	}{
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			NumericRegexp: nil,
			Vars:          []interface{}{"test", 1, 999.99, true, []byte("12345"), tt, &tt, nil, "w@g.\"com", myRole, pwd},
			Result:        `create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values ("test", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.""com", "admin", "pass")`,
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			NumericRegexp: nil,
			Vars:          []interface{}{"test?", 1, 999.99, true, []byte("12345"), tt, &tt, nil, "w@g.\"com", myRole, pwd},
			Result:        `create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values ("test?", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.""com", "admin", "pass")`,
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11)",
			NumericRegexp: regexp.MustCompile(`@p(\d+)`),
			Vars:          []interface{}{"test", 1, 999.99, true, []byte("12345"), tt, &tt, nil, "w@g.com", myRole, pwd},
			Result:        `create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values ("test", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.com", "admin", "pass")`,
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values ($3, $4, $1, $2, $7, $8, $5, $6, $9, $10, $11)",
			NumericRegexp: regexp.MustCompile(`\$(\d+)`),
			Vars:          []interface{}{999.99, true, "test", 1, &tt, nil, []byte("12345"), tt, "w@g.com", myRole, pwd},
			Result:        `create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values ("test", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.com", "admin", "pass")`,
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values (@p1, @p11, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10)",
			NumericRegexp: regexp.MustCompile(`@p(\d+)`),
			Vars:          []interface{}{"test", 999.99, true, []byte("12345"), tt, &tt, nil, "w@g.com", myRole, pwd, 1},
			Result:        `create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass) values ("test", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.com", "admin", "pass")`,
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, json_struct, example_struct) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			NumericRegexp: nil,
			Vars:          []interface{}{"test", 1, 999.99, true, []byte("12345"), tt, &tt, nil, "w@g.\"com", myRole, pwd, js, es},
			Result:        fmt.Sprintf(`create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, json_struct, example_struct) values ("test", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.""com", "admin", "pass", %v, %v)`, format(jsVal, `"`), format(esVal, `"`)),
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, json_struct, example_struct) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			NumericRegexp: nil,
			Vars:          []interface{}{"test", 1, 999.99, true, []byte("12345"), tt, &tt, nil, "w@g.\"com", myRole, pwd, &js, &es},
			Result:        fmt.Sprintf(`create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, json_struct, example_struct) values ("test", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.""com", "admin", "pass", %v, %v)`, format(jsVal, `"`), format(esVal, `"`)),
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, json_struct, example_struct) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			NumericRegexp: nil,
			Vars:          []interface{}{"test", 1, 0.1753607109, true, []byte("12345"), tt, &tt, nil, "w@g.\"com", myRole, pwd, &js, &es},
			Result:        fmt.Sprintf(`create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, json_struct, example_struct) values ("test", 1, 0.1753607109, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.""com", "admin", "pass", %v, %v)`, format(jsVal, `"`), format(esVal, `"`)),
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, json_struct, example_struct) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			NumericRegexp: nil,
			Vars:          []interface{}{"test", 1, float32(999.99), true, []byte("12345"), tt, &tt, nil, "w@g.\"com", myRole, pwd, &js, &es},
			Result:        fmt.Sprintf(`create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, json_struct, example_struct) values ("test", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.""com", "admin", "pass", %v, %v)`, format(jsVal, `"`), format(esVal, `"`)),
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, int_val) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			NumericRegexp: nil,
			Vars:          []interface{}{"test?", 1, 999.99, true, []byte("12345"), tt, &tt, nil, "w@g.\"com", myRole, pwd, intVal},
			Result:        `create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, int_val) values ("test?", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.""com", "admin", "pass", 1)`,
		},
		{
			SQL:           "create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, float_val) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			NumericRegexp: nil,
			Vars:          []interface{}{"test?", 1, 999.99, true, []byte("12345"), tt, &tt, nil, "w@g.\"com", myRole, pwd, floatVal},
			Result:        `create table users (name, age, height, actived, bytes, create_at, update_at, deleted_at, email, role, pass, float_val) values ("test?", 1, 999.99, true, "12345", "2024-09-26 17:45:15", "2024-09-26 17:45:15", NULL, "w@g.""com", "admin", "pass", 1.230000)`,
		},
	}

	for idx, r := range results {
		r := r
		t.Run(fmt.Sprintf("#%v", idx), func(t *testing.T) {
			result := ExplainSQL(r.SQL, r.NumericRegexp, `"`, r.Vars...)
			if result != r.Result {
				t.Errorf("explain SQL #%v\nexpected\n%v\nbut got\n%v", idx, r.Result, result)
			}

			// if result := logger.ExplainSQL(r.SQL, nil, `'`, r.Vars...); result != r.Result {
			// 	t.Errorf("explain SQL #%v\nexpected\n%v\nbut got\n%v", idx, r.Result, result)
			// }
		})
	}
}
