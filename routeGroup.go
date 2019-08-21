package rego

import (
	path2 "path"
)

//主要负责将得到路由信息后,将信息存入MethodRouteMaps
//最后通过engine来得到此MethodRouteMaps信息
//且结构体参数初始化在engine中初始化
type RouteGroup struct{
	//路由函数链
	handlers HandlerChain
	//路由路径
	path string
	//路由匹配Hash
	methodRouteMaps MethodRouteMaps
	//路由引擎
	engine *Engine
}



//添加路由信息相关
func (group *RouteGroup)addRoute(httpMethod,rePath string,handlers HandlerChain){
	result:=group.engine.methodRouteMaps.get(httpMethod)
	if result==nil{
		result=new(RouteMap)
		group.engine.methodRouteMaps=append(group.methodRouteMaps,MethodRouteMap{
			httpMethod: httpMethod,
			RouteMap:   result,
		})
		result.reset()
	}
	result.addRoute(rePath,handlers)
}
//func (group *RouteGroup)addRoute(httpMethod,rePath string,handlers HandlerChain){
//	result:=group.methodRouteMaps.get(httpMethod)
//	if result==nil{
//		result=new(RouteMaps)
//		group.methodRouteMaps=append(group.methodRouteMaps,MethodRouteMap{
//			httpMethod: httpMethod,
//			routeMap:   result,
//		})
//	}
//	result.addRouteMap(rePath,handlers)
//}


//处理路由信息相关
func (group *RouteGroup) Use(middleware ...HandlerFunc) {
	group.handlers = append(group.handlers, middleware...)
}

func (group *RouteGroup)Group(path string,handlers ...HandlerFunc)*RouteGroup{
	return &RouteGroup{
		handlers: group.combineHandlers(handlers),
		path:     group.absolutePath(path),
		methodRouteMaps:group.methodRouteMaps,
		engine: group.engine,
	}
}


func (group *RouteGroup)handle(httpMethod,path string,handlers HandlerChain){
	rePath:=group.absolutePath(path)
	reHandlers:=group.combineHandlers(handlers)
	group.addRoute(httpMethod,rePath,reHandlers)
}


func (group *RouteGroup)GET(path string,handlers ...HandlerFunc){
	 group.handle("GET",path,handlers)
}

func (group *RouteGroup)POST(path string,handlers ...HandlerFunc){
	group.handle("POST",path,handlers)
}

func (group *RouteGroup) DELETE(path string, handlers ...HandlerFunc) {
	 group.handle("DELETE", path, handlers)
}

func (group *RouteGroup) PATCH(path string, handlers ...HandlerFunc) {
	 group.handle("PATCH", path, handlers)
}

func (group *RouteGroup) PUT(path string, handlers ...HandlerFunc) {
	group.handle("PUT", path, handlers)
}

func (group *RouteGroup) OPTIONS(path string, handlers ...HandlerFunc) {
	group.handle("OPTIONS", path, handlers)
}

func (group *RouteGroup) HEAD(path string, handlers ...HandlerFunc) {
	group.handle("HEAD", path, handlers)
}

func (group *RouteGroup) ANY(path string, handlers ...HandlerFunc) {
	group.handle("GET", path, handlers)
	group.handle("POST", path, handlers)
	group.handle("PUT", path, handlers)
	group.handle("PATCH", path, handlers)
	group.handle("HEAD", path, handlers)
	group.handle("OPTIONS", path, handlers)
	group.handle("DELETE", path, handlers)
	group.handle("CONNECT", path, handlers)
	group.handle("TRACE", path, handlers)
}


func (group *RouteGroup)absolutePath(path string)string{
	repath:=group.path
	if path==""{
		return  repath
	}
	finalPath:=path2.Join(repath,path)
	pathLen:=len(path)-1
	finalPathLen:=len(finalPath)-1
	if path[pathLen]=='/'&&finalPath[finalPathLen]!='/'{
		return finalPath+"/"
	}
	return finalPath
}


func (group *RouteGroup)combineHandlers(handlers HandlerChain)HandlerChain{
	Size:=len(group.handlers)+len(handlers)
	if Size>=int(abortIndex){
		PrintErr(" too many  handlers")
	}
	reHandlers:=make(HandlerChain,Size)
	copy(reHandlers,group.handlers)
	copy(reHandlers[len(group.handlers):],handlers)
	return reHandlers
}


