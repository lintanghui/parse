package parse

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"go-common/xstr"
)

func TestBind(t *testing.T) {
	type v struct {
		Data16   int8 `params:"aaa;Range(1,10)" default:"10"`
		Data32   int32
		Data64   int64    `params:"data64;Range(1,20)" default:"20"`
		Float32  float32  `params:"ccc"`
		String   string   `params:"sss;Length(5,20)" default:"-"`
		SliceInt []int64  `params:"iii;Length(0,2)" default:"-"`
		SliceStr []string `params:"ttt"`
		Bool     bool     `params:"bbb"`
	}
	req, err := http.NewRequest("GET", "http://api.bilbili.com/x?data64=33&Data32=32&sss=aaaaaa&iii=1,2,3&ttt=a,b,c&bbb=true&ccc=1.2", nil)
	req.ParseForm()
	if err != nil {
		t.Log(err)
	}
	p := New()
	var data = &v{}
	err = p.Bind(data, req.Form)
	t.Logf("data %+v err %+v", data, err)
}

func BenchmarkBind(b *testing.B) {
	type v struct {
		Data16   int16    `params:"int16"`
		Data32   int32    `params:"int32"`
		String   string   `params:"string"`
		SliceInt []int    `params:"sliceInt"`
		SliceStr []string `params:"sliceStr"`
		Bool     bool     `params:"bool"`
	}
	req, err := http.NewRequest("GET", "http://api.bilbili.com/x?int16=11&int32=32&string=aaa&sliceInt=1,2,3&sliceStr=a,b,c&bool=true", nil)
	req.ParseForm()
	if err != nil {
		b.Log(err)
	}
	p := New()
	for i := 0; i < b.N; i++ {
		data := new(v)
		err := p.Bind(data, req.Form)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkCommonGet(b *testing.B) {
	type v struct {
		Data16   int16    `params:"int16"`
		Data32   int32    `params:"int32"`
		String   string   `params:"string"`
		SliceInt []int64  `params:"sliceInt"`
		SliceStr []string `params:"sliceStr"`
		Bool     bool     `params:"bool"`
	}
	req, err := http.NewRequest("GET", "http://api.bilbili.com/x?int16=11&int32=32&string=aaa&sliceInt=1,2,3&sliceStr=a,b,c&bool=true", nil)
	req.ParseForm()
	if err != nil {
		b.Log(err)
	}
	for i := 0; i < b.N; i++ {
		data := new(v)
		params := req.Form
		var (
			i16Str = params.Get("int16")
			i32Str = params.Get("int32")
			str    = params.Get("string")
			sInt   = params.Get("sliceInt")
			sStr   = params.Get("sliceStr")
			bStr   = params.Get("bool")
		)
		i16, _ := strconv.ParseInt(i16Str, 10, 64)
		data.Data16 = int16(i16)
		i32, _ := strconv.ParseInt(i32Str, 10, 64)
		data.Data32 = int32(i32)
		data.String = str
		s := strings.Split(sStr, ",")
		data.SliceStr = s
		si, _ := xstr.SplitInts(sInt)
		data.SliceInt = si
		b, _ := strconv.ParseBool(bStr)
		data.Bool = b
	}
}
