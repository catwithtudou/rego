package rego

import (
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"rego/responsedata"
	"strings"
	"time"
)

//路由处理函数阈值
const abortIndex int8 = math.MaxInt8 /2


//处理路由的上下文顺序,返回格式,参数获取和设置,等等关于HandLer的情况

type Context struct{
	Writer http.ResponseWriter
	Request *http.Request
	index  int8
	Params Params
	handlers HandlerChain
	engine *Engine
	keyParam map[string]interface{}
	queryCache url.Values
	formCache url.Values
}

func (c *Context)reset(){
	c.index=-1
	c.Params=c.Params[0:0]
	c.queryCache=nil
	c.formCache=nil
	c.keyParam=nil
}




//路由相关
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) AbortWithStatus(code int) {
	c.Writer.WriteHeader(code)
	c.Abort()
}


//路由上下文设置获取参数相关
func (c *Context) Set(key string, value interface{}) {
	if c.keyParam == nil {
		c.keyParam = make(map[string]interface{})
	}
	c.keyParam[key] = value
}

func (c *Context) Get(key string) (interface{},bool) {
	value, exists := c.keyParam[key]
	return value,exists
}

func (c *Context) GetString(key string) (s string) {
	if value, ok := c.Get(key); ok && value != nil {
		s, _ = value.(string)
	}
	return
}

func (c *Context) GetBool(key string) (b bool) {
	if value, ok := c.Get(key); ok && value != nil {
		b, _ = value.(bool)
	}
	return
}

func (c *Context) GetInt(key string) (i int) {
	if value, ok := c.Get(key); ok && value != nil {
		i, _ = value.(int)
	}
	return
}

func (c *Context) GetInt64(key string) (i64 int64) {
	if value, ok := c.Get(key); ok && value != nil {
		i64, _ = value.(int64)
	}
	return
}

func (c *Context) GetFloat64(key string) (f64 float64) {
	if value, ok := c.Get(key); ok && value != nil {
		f64, _ = value.(float64)
	}
	return
}

func (c *Context) GetTime(key string) (t time.Time) {
	if value, ok := c.Get(key); ok && value != nil {
		t, _ = value.(time.Time)
	}
	return
}

func (c *Context) GetStringArray(key string) (sa []string) {
	if value, ok := c.Get(key); ok && value != nil {
		sa, _ = value.([]string)
	}
	return
}

func (c *Context) GetStringMap(key string) (sm map[string]interface{}) {
	if value, ok := c.Get(key); ok && value != nil {
		sm, _ = value.(map[string]interface{})
	}
	return
}

func (c *Context) GetStringMapString(key string) (sms map[string]string) {
	if value, ok := c.Get(key); ok && value != nil {
		sms, _ = value.(map[string]string)
	}
	return
}

func (c *Context) GetStringMapStringArray(key string) (smsa map[string][]string) {
	if value, ok := c.Get(key); ok && value != nil {
		smsa, _ = value.(map[string][]string)
	}
	return
}


//返回相关
func (c *Context)Header(key,value string){
	if value=="" {
		c.Writer.Header().Del(key)
		return
	}
	c.Writer.Header().Set(key,value)
}

func (c *Context)GetHeader(key string)string{
	return c.Request.Header.Get(key)
}

func (c *Context) GetRawData() ([]byte, error) {
	return ioutil.ReadAll(c.Request.Body)
}

func (c *Context) SetCookie(key, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}else{
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     key,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})}
}

func (c *Context)Cookie(key string)(string,error){
	cookie,err:=c.Request.Cookie(key)
	if err != nil {
		return "",err
	}
	val,_:=url.QueryUnescape(cookie.Value)
	return val,nil
}

func AllowedStatus(status int) (flag bool) {
	flag=true
	switch {
	case status >= 100 && status < 200:
		flag=false
	case status == http.StatusNoContent:
		flag=false
	case status == http.StatusNotModified:
		flag=false
	}
	return
}

func (c *Context)ResponseData(code int,r responsedata.ResponseData){
	if !AllowedStatus(code){
		r.WriteContentType(c.Writer)
		return
	}
	err:=r.ResponseData(c.Writer)
	if err != nil {
		CheckErr(err,"response the data failed")
	}
}

func (c *Context)String(code int,format string,values ...interface{}){
	c.ResponseData(code,responsedata.String{
		Format: format,
		Data:   values,
	})
}

func (c *Context)JSON(code int,obj interface{}){
	c.ResponseData(code,responsedata.JSON{Data:obj})
}

func (c *Context)XML(code int, obj interface{}) {
	c.ResponseData(code, responsedata.XML{Data: obj})
}

func (c *Context)Redirect(code int,finalUrl string){
	c.ResponseData(-1,responsedata.Redirect{
		Code:     code,
		Request:  c.Request,
		FinalUrl: finalUrl,
	})
}

//返回参数相关
/*获取url里面的动态参数*/
func (c *Context)Param(key string)string{
	return c.Params.ByName(key)
}

/*获取url后面的静态参数*/
func (c *Context)Query(key string)string{
	value,_:=c.GetQuery(key)
	return value
}

/*判断url后面的静态参数,是否存在,若存在且返回参数值*/
func (c *Context)GetQuery(key string)(string, bool){
	if values,flag:=c.GetQueryArray(key);flag{
		return values[0],flag
	}
	return "",false
}

/*获取Url后面的静态参数,并可以设置默认值*/
func (c *Context) DefaultQuery(key, defaultValue string)(value string) {
	value, ok := c.GetQuery(key)
	if ok {
		return value
	}
	value=defaultValue
	return
}

/*获取url后面的静态参数数组*/
func (c *Context)QueryArray(key string)[]string{
	values,_:=c.GetQueryArray(key)
	return values
}

/*判断url后面的静态参数数组是否存在,若存在则返回参数数组*/
func (c *Context)GetQueryArray(key string)([]string,bool){
	c.getQueryCache()
	if values, flag := c.queryCache[key];   len(values) > 0 &&flag {
		return values, true
	}
	return []string{}, false
}
func (c *Context)getQueryCache(){
	if c.queryCache==nil{
		c.queryCache=make(url.Values)
		c.queryCache,_=url.ParseQuery(c.Request.URL.RawQuery)
	}
}


/*获取静态参数给定查询键的Map映射*/
func (c *Context) QueryMap(key string) map[string]string {
	mapQuery, _ := c.GetQueryMap(key)
	return mapQuery
}

/*判断静态参数给定查询键的Map映射是否存在,若存在则返回相应Map*/
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.getQueryCache()
	return c.get(c.queryCache, key)
}

/*获取给定表单参数*/
func (c *Context) PostForm(key string) string {
	value, _ := c.GetPostForm(key)
	return value
}

/*在获取给定表单参数时,可设置默认值*/
func (c *Context) DefaultPostForm(key, defaultValue string) string {
	if value, flag := c.GetPostForm(key); flag {
		return value
	}
	return defaultValue
}

/*判断给定表单参数是否存在并返回相应布尔值,且存在返回参数值*/
func (c *Context) GetPostForm(key string) (string, bool) {
	if values, flag := c.GetPostFormArray(key); flag {
		return values[0], flag
	}
	return "", false
}

/*获取给定表单参数中字符串数组*/
func (c *Context) PostFormArray(key string) []string {
	values, _ := c.GetPostFormArray(key)
	return values
}
func (c *Context) getFormCache() {
	if c.formCache == nil {
		c.formCache = make(url.Values)
		req := c.Request
		if err := req.ParseMultipartForm(c.engine.MaxMultipartMemory); err != nil {
			if err != http.ErrNotMultipart {
				CheckErr(err,"error on parse multipart form array")
			}
		}
		c.formCache = req.PostForm
	}
}

/*判断给定表单参数中字符串数组是否存在并返回相应布尔值,且存在返回字符串值*/
func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.getFormCache()
	if values := c.formCache[key]; len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

/*判读给定表单参数中相应查询键的Map映射是否存在并返回相应布尔值,且存在返回相应Map*/
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	req := c.Request
	if err := req.ParseMultipartForm(c.engine.MaxMultipartMemory); err != nil {
		if err != http.ErrNotMultipart {
			CheckErr(err,"error on parse multipart form map")
		}
	}
	return c.get(req.PostForm, key)
}

/*获取满足key的Map映射*/
func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	for k, v := range m {
		if i := strings.IndexByte(k, '['); i > 0 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j > 0 {
				exist = true
				dicts[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dicts, exist
}

/*处理单一文件的上传*/
func (c *Context)FormFile(name string)(*multipart.FileHeader,error){
	if c.Request.MultipartForm == nil{
		if err:=c.Request.ParseMultipartForm(c.engine.MaxMultipartMemory);err!=nil{
			return nil,err
		}
	}
	_,fileHeader,err:=c.Request.FormFile(name)
	return fileHeader,err
}

/*处理多一文件的上传*/
func (c *Context)MultipartForm()(*multipart.Form,error){
	err:=c.Request.ParseMultipartForm(c.engine.MaxMultipartMemory)
	return c.Request.MultipartForm,err
}

/*处理文件上传到指定的文件位置*/
func (c *Context)SaveUploadedFile(file *multipart.FileHeader,dst string)error{
	src,err:=file.Open()
	CheckErr(err,"open the File Failed")
	defer src.Close()
	out,err:=os.Create(dst)
	CheckErr(err,"can't create the File")
	defer out.Close()
	_,err=io.Copy(out,src)
	return err
}

/*打开静态文件*/
func (c *Context) File(path string) {
	http.ServeFile(c.Writer, c.Request, path)
}