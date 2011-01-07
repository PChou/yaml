package goyaml_test


import (
    . "gocheck"
    "testing"
    "goyaml"
    "reflect"
    "math"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

var unmarshalTests = []struct{data string; value interface{}}{
    // It will encode either value as a string if asked for.
    {"hello: world", map[string]string{"hello": "world"}},
    {"hello: true", map[string]string{"hello": "true"}},

    // And when given the option, will preserve the YAML type.
    {"hello: world", map[string]interface{}{"hello": "world"}},
    {"hello: true", map[string]interface{}{"hello": true}},
    {"hello: 10", map[string]interface{}{"hello": 10}},
    {"hello: 0b10", map[string]interface{}{"hello": 2}},
    {"hello: 0xA", map[string]interface{}{"hello": 10}},
    {"hello: 4294967296", map[string]interface{}{"hello": int64(4294967296)}},
    {"hello: 4294967296", map[string]int64{"hello": int64(4294967296)}},
    {"hello: 0.1", map[string]interface{}{"hello": 0.1}},
    {"hello: .1", map[string]interface{}{"hello": 0.1}},
    {"hello: .Inf", map[string]interface{}{"hello": math.Inf(+1)}},
    {"hello: -.Inf", map[string]interface{}{"hello": math.Inf(-1)}},
    {"hello: -10", map[string]interface{}{"hello": -10}},
    {"hello: -.1", map[string]interface{}{"hello": -0.1}},

    // Floats from spec
    {"canonical: 6.8523e+5", map[string]interface{}{"canonical": 6.8523e+5}},
    {"expo: 685.230_15e+03", map[string]interface{}{"expo": 685.23015e+03}},
    {"fixed: 685_230.15", map[string]interface{}{"fixed": 685230.15}},
    //{"sexa: 190:20:30.15", map[string]interface{}{"sexa": 0}}, // Unsupported
    {"neginf: -.inf", map[string]interface{}{"neginf": math.Inf(-1)}},
    {"notanum: .NaN", map[string]interface{}{"notanum": math.NaN}},
    {"fixed: 685_230.15", map[string]float{"fixed": 685230.15}},

    // Bools from spec
    {"canonical: y", map[string]interface{}{"canonical": true}},
    {"answer: NO", map[string]interface{}{"answer": false}},
    {"logical: True", map[string]interface{}{"logical": true}},
    {"option: on", map[string]interface{}{"option": true}},
    {"option: on", map[string]bool{"option": true}},

    // Ints from spec
    {"canonical: 685230", map[string]interface{}{"canonical": 685230}},
    {"decimal: +685_230", map[string]interface{}{"decimal": 685230}},
    {"octal: 02472256", map[string]interface{}{"octal": 685230}},
    {"hexa: 0x_0A_74_AE", map[string]interface{}{"hexa": 685230}},
    {"bin: 0b1010_0111_0100_1010_1110", map[string]interface{}{"bin": 685230}},
    {"bin: -0b101010", map[string]interface{}{"bin": -42}},
    //{"sexa: 190:20:30", map[string]interface{}{"sexa": 0}}, // Unsupported
    {"decimal: +685_230", map[string]int{"decimal": 685230}},

    // Nulls from spec
    {"empty:", map[string]interface{}{"empty": nil}},
    {"canonical: ~", map[string]interface{}{"canonical": nil}},
    {"english: null", map[string]interface{}{"english": nil}},
    {"~: null key", map[interface{}]string{nil: "null key"}},
    {"empty:", map[string]*bool{"empty": nil}},

    // Sequence
    {"seq: [A,B]", map[string]interface{}{"seq": []interface{}{"A", "B"}}},
    {"seq: [A,B,C]", map[string][]string{"seq": []string{"A", "B", "C"}}},
    {"seq: [A,1,C]", map[string][]string{"seq": []string{"A", "1", "C"}}},
    {"seq: [A,1,C]", map[string][]int{"seq": []int{1}}},
    {"seq: [A,1,C]", map[string]interface{}{"seq": []interface{}{"A", 1, "C"}}},

    // Map inside interface with no type hints.
    {"a: {b: c}",
     map[string]interface{}{"a": map[interface{}]interface{}{"b": "c"}}},

    // Structs and type conversions.
    {"hello: world", &struct{Hello string}{"world"}},
    {"a: {b: c}", &struct{A struct{B string}}{struct{B string}{"c"}}},
    {"a: {b: c}", &struct{A *struct{B string}}{&struct{B string}{"c"}}},
    {"a: 1", &struct{A int}{1}},
    {"a: [1, 2]", &struct{A []int}{[]int{1, 2}}},
    {"a: 1", &struct{B int}{0}},
    {"a: 1", &struct{B int "a"}{1}},
    {"a: y", &struct{A bool}{true}},

    // Some cross type conversions
    {"v: 42", map[string]uint{"v": 42}},
    {"v: -42", map[string]uint{}},
    {"v: 4294967296", map[string]uint64{"v": 4294967296}},
    {"v: -4294967296", map[string]uint64{}},

    // Overflow cases.
    {"v: 4294967297", map[string]int32{}},
    {"v: 128", map[string]int8{}},

    // Quoted values.
    {"'1': '2'", map[interface{}]interface{}{"1": "2"}},

    // Explicit tags.
    {"v: !!float '1.1'", map[string]interface{}{"v": 1.1}},
    {"v: !!null ''", map[string]interface{}{"v": nil}},
    {"%TAG !y! tag:yaml.org,2002:\n---\nv: !y!int '1'",
     map[string]interface{}{"v": 1}},
}


func (s *S) TestUnmarshal(c *C) {
    for _, item := range unmarshalTests {
        t := reflect.NewValue(item.value).Type()
        var value interface{}
        if t, ok := t.(*reflect.MapType); ok {
            value = reflect.MakeMap(t).Interface()
        } else {
            pt := reflect.NewValue(item.value).Type().(*reflect.PtrType)
            pv := reflect.MakeZero(pt).(*reflect.PtrValue)
            pv.PointTo(reflect.MakeZero(pt.Elem()))
            value = pv.Interface()
        }
        err := goyaml.Unmarshal([]byte(item.data), value)
        c.Assert(err, IsNil)
        c.Assert(value, Equals, item.value)
    }
}

var unmarshalErrorTests = []struct{data, error string}{
    {"v: !!float 'error'", "Can't decode !!str 'error' as a !!float"},
}

func (s *S) TestUnmarshalErrors(c *C) {
    for _, item := range unmarshalErrorTests {
        var value interface{}
        err := goyaml.Unmarshal([]byte(item.data), &value)
        c.Assert(err, Matches, item.error)
    }
}

var setterTests = []struct{data, tag string; value interface{}}{
    {"_: {hi: there}", "!!map", map[interface{}]interface{}{"hi": "there"}},
    {"_: [1,A]", "!!seq", []interface{}{1, "A"}},
    {"_: 10", "!!int", 10},
    {"_: null", "!!null", nil},
    {"_: !!foo 'BAR!'", "!!foo", "BAR!"},
}

var setterResult = map[int]bool{}

type typeWithSetter struct {
    tag string
    value interface{}
}

func (o *typeWithSetter) SetYAML(tag string, value interface{}) (ok bool) {
    o.tag = tag
    o.value = value
    if i, ok := value.(int); ok {
        if result, ok := setterResult[i]; ok {
            return result
        }
    }
    return true
}

type typeWithSetterField struct {
    Field *typeWithSetter "_"
}

func (s *S) TestUnmarshalWithSetter(c *C) {
    for _, item := range setterTests {
        obj := &typeWithSetterField{}
        err := goyaml.Unmarshal([]byte(item.data), obj)
        c.Assert(err, IsNil)
        c.Assert(obj.Field, NotNil,
                 Bug("Pointer not initialized (%#v)", item.value))
        c.Assert(obj.Field.tag, Equals, item.tag)
        c.Assert(obj.Field.value, Equals, item.value)
    }
}

func (s *S) TestUnmarshalWholeDocumentWithSetter(c *C) {
    obj := &typeWithSetter{}
    err := goyaml.Unmarshal([]byte(setterTests[0].data), obj)
    c.Assert(err, IsNil)
    c.Assert(obj.tag, Equals, setterTests[0].tag)
    value, ok := obj.value.(map[interface{}]interface{})
    c.Assert(ok, Equals, true)
    c.Assert(value["_"], Equals, setterTests[0].value)
}

func (s *S) TestUnmarshalWithFalseSetterIgnoresValue(c *C) {
    setterResult[2] = false
    setterResult[4] = false
    defer func() {
        setterResult[2] = false, false
        setterResult[4] = false, false
    }()

    m := map[string]*typeWithSetter{}
    data := "{abc: 1, def: 2, ghi: 3, jkl: 4}"
    err := goyaml.Unmarshal([]byte(data), m)
    c.Assert(err, IsNil)
    c.Assert(m["abc"], NotNil)
    c.Assert(m["def"], IsNil)
    c.Assert(m["ghi"], NotNil)
    c.Assert(m["jkl"], IsNil)

    c.Assert(m["abc"].value, Equals, 1)
    c.Assert(m["ghi"].value, Equals, 3)
}