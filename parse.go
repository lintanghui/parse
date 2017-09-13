/*
Package parse parse url.Values and bind values to object(struct).

Supported Types
 - int,int8,int16,int32,int64
 - uint,uint8,uint16,uint32,uint64
 - string
 - []string,[]int64
 - bool
 - float32,float64

Support Func
 - Range(min,max) // field value must between min and max.
 - Min(min) // filed must bigger than min
 - Length(min,max) // len(filed) must between min and max.
The object's default key string is the struct field name
but can be specified in the struct field's tag value. The "params" key in
the struct field's tag value is the key name, followed by an optional semicolon
and options. Examples:

 // Field appear in url.Values as key urlField
 Field int `params:"urlFiled"`

 // Field appear in url.Values as key v,and v's value must between 1 and 100
 Field int64 `params:"v;Range(1,100)"`

The Field can specific default value using strcut field's tag with name 'default',
if field's value is not requried in url.Values. You should use default:"-".Example：

 // Field's value must between 1 and 100,if not ,set it to default value 100
 Field int64 `params:"field;Range(1,100)" default:"100"`

 // Field's value is not required in url.Values by using default:"-".
 // if value not apear in url.Values.ingore this field.
 Field string `params:"field" default:"-"`

Example:
 package main
 import (
	 "net/http"
	 "fmt"

	 "github.com/lintanghui/parse
 )
 func main(){
	type v struct {
		Data16   int8 `params:"aaa;Range(1,10)" default:"10"`
		Data32   int32
		Data64   int64    `params:"data64;Range(1,20)" default:"20"`
		Float32  float32  `params:"ccc"`
		String   string   `params:"sss" default:"-"`
		SliceInt []int64  `params:"iii"`
		SliceStr []string `params:"ttt"`
		Bool     bool     `params:"bbb"`
	}
	req, err := http.NewRequest("GET", "http://api.bilbili.com/x?aaa=11&data64=33&Data32=32&string=aaa&iii=1,2,3&ttt=a,b,c&bbb=true&ccc=1.2", nil)
	req.ParseForm()
	if err != nil {
		t.Log(err)
	}
	p := parse.New()
	var data = &v{}
	err = p.Bind(data, req.Form)
	fmt.Printf("%+v",data)
	// OUTPUT:
	// &{Data16:10 Data32:32 Data64:20 Float32:1.2 String: SliceInt:[1 2 3] SliceStr:[a b c] Bool:true}
 }
*/
package parse

import (
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	_omit = "-"
)

// global err define.
var (
	ErrObjType      = errors.New("obj must be ptr to struct")
	ErrInvalidFunc  = errors.New("invalid valid function")
	ErrInvalidParam = errors.New("invalid valid params")
)

// Parse params parser.
type Parse struct {
	objCache map[reflect.Type][]*pfield
	mutex    sync.RWMutex
}

// New new and return parser.
func New() *Parse {
	return &Parse{
		objCache: make(map[reflect.Type][]*pfield),
	}
}

// Bind bind url.Values params to obj(struct).
func (p *Parse) Bind(obj interface{}, req url.Values) (err error) {
	fcs, err := p.fieldCache(reflect.TypeOf(obj))
	if err != nil {
		return
	}
	objV := reflect.ValueOf(obj).Elem()
	for i, f := range fcs {
		value := req.Get(f.name)
		err = setValue(value, objV.Field(i), f)
		if err != nil {
			if !f.def {
				return
			}
			err = nil
			objV.Field(i).Set(f.defValue)
		}
	}
	return
}

type pfield struct {
	ftype    reflect.Type
	name     string
	vfuncs   []validFuncs
	def      bool          // if field had default value
	defValue reflect.Value // field default value
}

func (p *Parse) fieldCache(obj reflect.Type) (fs []*pfield, err error) {
	p.mutex.RLock()
	fs, ok := p.objCache[obj]
	if !ok {
		p.mutex.RUnlock()
		fs, err = p.register(obj)
		return
	}
	p.mutex.RUnlock()
	return
}

// Register register object into parser.
func (p *Parse) Register(objs ...reflect.Type) (err error) {
	for _, obj := range objs {
		_, err = p.register(obj)
		if err != nil {
			return
		}
	}
	return
}

func (p *Parse) register(obj reflect.Type) (fs []*pfield, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err = isValid(obj); err != nil {
		return
	}
	objT := obj.Elem()
	for i := 0; i < objT.NumField(); i++ {
		var (
			vfs []validFuncs
		)
		field := objT.Field(i)
		pStr := field.Tag.Get("params")
		params := strings.Split(pStr, ";")
		name := params[0]
		if len(name) == 0 {
			name = objT.Field(i).Name
		}
		if len(params) > 1 {
			vfs, err = getFuncs(params[1:])
			if err != nil {
				return
			}
		}
		ftype := objT.Field(i).Type
		ofield := &pfield{
			name:   name,
			ftype:  ftype,
			vfuncs: vfs,
		}
		def := field.Tag.Get("default")
		if def != "" {
			dv := reflect.New(ftype).Elem()
			if def != _omit {
				if err = setValue(def, dv, ofield); err != nil {
					return
				}
			}
			ofield.def = true
			ofield.defValue = dv
		}
		fs = append(fs, ofield)
	}
	p.objCache[obj] = fs
	return
}

// ValidFuncs define validfuncs .
type validFuncs struct {
	name   string
	params []interface{}
}

func getFuncs(fs []string) (vfs []validFuncs, err error) {
	for _, f := range fs {
		var vf validFuncs
		if len(f) == 0 {
			continue
		}
		vf, err = parseFunc(f)
		if err != nil {
			return
		}
		vfs = append(vfs, vf)
	}
	return
}

func parseFunc(s string) (v validFuncs, err error) {
	s = strings.TrimSpace(s)
	start := strings.Index(s, "(")
	if start == -1 {
		// todo
		return
	}
	var num int
	name := s[:start]
	num, err = funcIn(name)
	end := strings.Index(s, ")")
	if end == -1 {
		err = ErrInvalidFunc
		return
	}
	params := strings.Split(s[start+1:end], ",")
	if len(params) != num {
		err = ErrInvalidFunc
		return
	}
	tParams, err := parseParams(name, params)
	v = validFuncs{name: name, params: tParams}
	return
}
func parseParams(name string, ps []string) (params []interface{}, err error) {
	f, ok := funcs[name]
	if !ok {
		err = ErrInvalidFunc
		return
	}
	for i, p := range ps {
		var param interface{}
		if param, err = parseParam(f.Type().In(i+2), p); err != nil {
			return
		}
		params = append(params, param)
	}
	return
}

func parseParam(t reflect.Type, s string) (i interface{}, err error) {
	switch t.Kind() {
	case reflect.Int:
		i, err = strconv.Atoi(s)
	case reflect.Int8:
		i, err = strconv.ParseInt(s, 10, 8)
		if err == nil {
			i = i.(int8)
		}
	case reflect.Int16:
		i, err = strconv.ParseInt(s, 10, 16)
		if err == nil {
			i = i.(int16)
		}
	case reflect.Int32:
		i, err = strconv.ParseInt(s, 10, 32)
		if err == nil {
			i = i.(int32)
		}
	case reflect.Int64:
		i, err = strconv.ParseInt(s, 10, 64)
	case reflect.Uint:
		i, err = strconv.ParseUint(s, 10, 32)
		if err == nil {
			i = i.(uint)
		}
	case reflect.Uint8:
		i, err = strconv.ParseUint(s, 10, 8)
		if err == nil {
			i = i.(uint32)
		}
	case reflect.Uint16:
		i, err = strconv.ParseUint(s, 10, 16)
		if err == nil {
			i = i.(uint16)
		}
	case reflect.Uint32:
		i, err = strconv.ParseUint(s, 10, 32)
		if err == nil {
			i = i.(uint32)
		}
	case reflect.Uint64:
		i, err = strconv.ParseUint(s, 10, 64)
		if err == nil {
			i = i.(uint64)
		}
	case reflect.String:
		i = s
	case reflect.Interface:
		i = s
	}
	return
}

func funcIn(vf string) (num int, err error) {
	f, ok := funcs[vf]
	if !ok {
		err = ErrInvalidFunc
		return
	}
	num = f.Type().NumIn() - 2
	return
}

var _bitMap = map[reflect.Kind]int{
	reflect.Int:     32,
	reflect.Int16:   16,
	reflect.Int32:   32,
	reflect.Int64:   64,
	reflect.Int8:    8,
	reflect.Uint:    32,
	reflect.Uint16:  16,
	reflect.Uint32:  32,
	reflect.Uint64:  64,
	reflect.Uint8:   8,
	reflect.Float32: 32,
	reflect.Float64: 64,
}

func setValue(data string, v reflect.Value, f *pfield) (err error) {
	kind := v.Kind()
	switch kind {
	case reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int:
		var d int64
		d, err = strconv.ParseInt(data, 10, _bitMap[kind])
		if err != nil {
			return
		}
		v.SetInt(d)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		var d uint64
		d, err = strconv.ParseUint(data, 10, _bitMap[kind])
		if err != nil {
			return
		}
		v.SetUint(d)
	case reflect.Float64, reflect.Float32:
		var d float64
		d, err = strconv.ParseFloat(data, _bitMap[kind])
		if err != nil {
			return
		}
		v.SetFloat(d)
	case reflect.String:
		if data == "" {
			err = ErrInvalidParam
			return
		}
		v.SetString(data)
	case reflect.Bool:
		var b bool
		b, err = strconv.ParseBool(data)
		if err != nil {
			return
		}
		v.SetBool(b)
	case reflect.Slice:
		sl := strings.Split(data, ",")
		switch v.Type().Elem().Kind() {
		case reflect.Int64:
			valSli := make([]int64, len(sl))
			for i := 0; i < len(sl); i++ {
				var d int64
				if d, err = strconv.ParseInt(sl[i], 10, 64); err != nil {
					return
				}
				valSli[i] = d
			}
			v.Set(reflect.ValueOf(valSli))
		case reflect.String:
			v.Set(reflect.ValueOf(sl))
		}
	}
	for _, vf := range f.vfuncs {
		if _, err = funcs.Call(vf.name, append([]interface{}{va, v}, vf.params...)...); err != nil {
			return
		}
	}
	return
}

func isValid(obj reflect.Type) (err error) {
	if obj.Kind() == reflect.Ptr && obj.Elem().Kind() == reflect.Struct {
		return
	}
	err = ErrObjType
	return
}
