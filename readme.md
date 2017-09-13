# Package parse parse url.Values and bind values to object(struct).

# Supported Types
 - int,int8,int16,int32,int64
 - uint,uint8,uint16,uint32,uint64
 - string
 - []string,[]int64
 - bool
 - float32,float64

# Support Func  
 - Range(min,max) // field value must between min and max.  
 - Min(min) // filed must bigger than min  
 - Length(min,max) // len(filed) must between min and max.  

The object's default key string is the struct field name
but can be specified in the struct field's tag value. The "params" key in
the struct field's tag value is the key name, followed by an optional semicolon
and options. Examples:

```
 // Field appear in url.Values as key urlField  
 Field int `params:"urlFiled"`

 // Field appear in url.Values as key v,and v's value must between 1 and 100
 Field int64 `params:"v;Range(1,100)"`

``` 

The Field can specific default value using strcut field's tag with name 'default',
if field's value is not requried in url.Values. You should use default:"-".Example： 

```
 // Field's value must between 1 and 100,if not ,set it to default value 100  
 
 Field int64 `params:"field;Range(1,100)" default:"100"`

 // Field's value is not required in url.Values by using default:"-".
 // if value not apear in url.Values.ingore this field.  
 Field string `params:"field" default:"-"`

```

Example:
```
 package main
 import (
	 "net/http"
	 "fmt"

	 "github.com/lintanghui/parse"
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
	req, err := http.NewRequest("GET", "http://www.linth.top/x?aaa=11&data64=33&Data32=32&string=aaa&iii=1,2,3&ttt=a,b,c&bbb=true&ccc=1.2", nil)
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
```


