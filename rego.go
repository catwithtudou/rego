package rego

import (
	"net/http"
	"sync"
)

//定义处理路由函数结构体
type HandlerFunc func(*Context)
type HandlerChain []HandlerFunc

// 32 MB
const defaultMultipartMemory = 32 << 20

//定义404与405消息体
var (
	default404 = []byte("404 page not found")
	default405 = []byte("404 method not allowed")
)

//框架核心引擎
type Engine struct {
	RouteGroup
	pool               sync.Pool
	MaxMultipartMemory int64
	WS
}

//初始化
func New() *Engine {
	engine := &Engine{
		RouteGroup: RouteGroup{
			handlers:        nil,
			path:            "/",
			methodRouteMaps: make(MethodRouteMaps, 0, 9),
		},
		MaxMultipartMemory: defaultMultipartMemory,
		WS:WS{
			wsMap: make(map[string]WSConfig),
		},
	}
	engine.RouteGroup.engine=engine
	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}
	return engine
}
func (engine *Engine) allocateContext() *Context {
	return &Context{engine: engine}
}

//启动引擎
func (engine *Engine) Run(addr string) (err error) {
	if addr == "" {
		addr = ":8080"
	}
	err = http.ListenAndServe(addr, engine)
	CheckErr(err, "start the engine failed")
	return
}


func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := engine.pool.Get().(*Context)
	c.Writer = w
	c.Request = r
	c.reset()
	engine.HandlerHTTPRequest(c)
	engine.pool.Put(c)
}

//处理请求
func (engine *Engine) HandlerHTTPRequest(c *Context) {
	httpMethod := c.Request.Method
	rPath := c.Request.URL.Path
	rPath = cleanPath(rPath)
	AllMaps := engine.methodRouteMaps
	upgrade := c.Request.Header.Get("Upgrade")
	if upgrade == "websocket" {
		config, ok := engine.wsMap[rPath];
		if !ok {
			_, err := c.Writer.Write(default404)
			CheckErr(err, "can't find the config")
			return
		}
		hijacker:=c.Writer.(http.Hijacker)
		con,buf,err:=hijacker.Hijack()
		if err != nil {
			CheckErr(err,"get the connection is failed")
			_ = con.Close()
		}
		engine.Conn=con
		engine.Buf=buf
		engine.handleConnection(c, config)
		if err != nil {
			CheckErr(err,"get the connection is failed")
			_ = con.Close()
		}
	} else {
		for _, v := range AllMaps {
			if v.httpMethod != httpMethod {
				continue
			}
			routeMaps := v.RouteMap
			value := routeMaps.getValue(rPath, c.Params)
			if value.handlers != nil {
				c.handlers = value.handlers
				c.Params = value.params
				c.Next()
				return
			}
		}
		_, err := c.Writer.Write(default404)
		CheckErr(err, "can't find the route")
	}

}
