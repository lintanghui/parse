package parse

import "reflect"

// Validation valida struct.
type Validation struct{}

// Result valida result.
type Result struct {
	err error
	OK  bool
}

var (
	funcs = make(Funcs)
	va    = &Validation{}
)

// Funcs valida func list.
type Funcs map[string]reflect.Value

// Call call func by func name and params.
func (f Funcs) Call(name string, params ...interface{}) (result []reflect.Value, err error) {
	if _, ok := f[name]; !ok {
		err = ErrInvalidFunc
		return
	}
	in := make([]reflect.Value, len(params))
	for i, p := range params {
		in[i] = reflect.ValueOf(p)
	}
	result = f[name].Call(in)
	if !result[0].Bool() {
		err = ErrInvalidParam
	}
	return
}

// Range check if value between min and max.
func (v *Validation) Range(obj reflect.Value, min, max int64) (ok bool) {
	i := obj.Int()
	if i <= max && i >= min {
		ok = true
	}
	return
}

// Length check if obj's len is between min and max.
func (v *Validation) Length(obj reflect.Value, min, max int) (ok bool) {
	var l int
	switch obj.Kind() {
	case reflect.String:
		l = len([]rune(obj.String()))
	default:
		l = obj.Len()
	}
	if l <= max && l >= min {
		ok = true
	}
	return
}

// Min check if obj value bigger than min.
func (v *Validation) Min(obj reflect.Value, min int64) (ok bool) {
	i := obj.Int()
	if i >= min {
		ok = true
	}
	return
}
func init() {
	t := reflect.TypeOf(va)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		funcs[m.Name] = m.Func
	}
}
