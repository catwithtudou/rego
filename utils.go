package rego

import (
	"encoding/xml"
	"log"
	"math"
)

func Min(a,b int)int{
	if b>=a{
		return a
	}
	return b
}


func CheckErr(err error,msg string){
	if err!=nil{
		log.Fatalf("%s : %s",msg,err)
	}
}

func PrintErr(format string,values ...interface{}){
	log.Fatalf(format,values)
}


func ContainInt(a []int,b int)bool{
	for _,v:=range a{
		if v==b{
			return true
		}
	}
	return false
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	n := len(p)
	var buf []byte
	r := 1
	w := 1
	if p[0] != '/' {
		r = 0
		buf = make([]byte, n+1)
		buf[0] = '/'
	}
	trailing := n > 1 && p[n-1] == '/'
	for r < n {
		switch {
		case p[r] == '/':
			r++
		case p[r] == '.' && r+1 == n:
			trailing = true
			r++
		case p[r] == '.' && p[r+1] == '/':
			r += 2
		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'):
			r += 3
			if w > 1 {
				w--
				if buf == nil {
					for w > 1 && p[w] != '/' {
						w--
					}
				} else {
					for w > 1 && buf[w] != '/' {
						w--
					}
				}
			}
		default:
			if w > 1 {
				bufApp(&buf, p, w, '/')
				w++
			}
			for r < n && p[r] != '/' {
				bufApp(&buf, p, w, p[r])
				w++
				r++
			}
		}
	}
	if trailing && w > 1 {
		bufApp(&buf, p, w, '/')
		w++
	}

	if buf == nil {
		return p[:w]
	}
	return string(buf[:w])
}

func bufApp(buf *[]byte, s string, w int, c byte) {
	if *buf == nil {
		if s[w] == c {
			return
		}
		*buf = make([]byte, len(s))
		copy(*buf, s[:w])
	}
	(*buf)[w] = c
}


func GetAllKeys(m map[string]HandlerChain)(result []string){
	keys:=make([]string,len(m))
	j:=0
	for k:=range m{
		keys[j]=k
		j++
	}
	return keys
}

func GetParamAllKeys(m map[string]ParamMap)(result []string){
	keys:=make([]string,len(m))
	j:=0
	for k:=range m{
		keys[j]=k
		j++
	}
	return keys
}

func IsKeyExist(m map[string]HandlerChain,key string)bool{
	for k:=range m{
		if k==key{
			return true
		}

	}
	return false
}

type H map[string]interface{}

func (h H) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := encoder.EncodeToken(start); err != nil {
		CheckErr(err,"encode the token failed")
		return err
	}
	for key, value := range h {
		element := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := encoder.EncodeElement(value, element); err != nil {
			CheckErr(err,"encode the element failed")
			return err
		}
	}
	return encoder.EncodeToken(xml.EndElement{Name: start.Name})
}


func ParseIntToBin(i int)(b []bool){
	b = make([]bool,8)
	for pos:=7;i>0;pos,i=pos-1,i/2{
		flag:=false
		if i%2==1{
			flag=true
		}
		b[pos]=flag
	}
	return
}

func ParseBinToInt(b []bool)(i int){
	i=0
	for pos,j:=0,len(b)-1;j>-1;j,pos=j-1,pos+1{
		if b[j]{
			i+=int(math.Pow(float64(2),float64(pos)))
		}
	}
	return
}