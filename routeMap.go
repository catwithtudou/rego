package rego

import (
	"strings"
	"sync"
)

//参数相关
type Param struct {
	Key   string
	Value string
}
type Params []Param

//路由信息相关
type MethodRouteMap struct {
	httpMethod string
	//routeMap   *RouteMaps
	*RouteMap
}
type MethodRouteMaps []MethodRouteMap

type RouteMap struct{
	sync.RWMutex
	staticMaps map[string]HandlerChain
	paramMaps map[string]ParamMap
	catchAllMaps map[string]HandlerChain
}
type ParamMap struct{
	pathParam map[int]string
	handlers HandlerChain
}
// /usr/:id/name

//
//type RouteMap struct {
//	nodeMap map[string]HandlerChain
//	nodeType
//	path string
//	paramName []string
//	paramIndex []int
//}
//type RouteMaps []RouteMap
//type nodeType uint8
//const (
//	static nodeType = iota
//	param
//	catchAll
//)


type RouteValue struct{
	handlers HandlerChain
	path string
	params Params
}




//得到MethodRouteMaps下某method的RouteMap
func (methodRouteMaps MethodRouteMaps)get(httpMethod string)*RouteMap{
	for _,v:=range methodRouteMaps{
		if v.httpMethod==httpMethod{
			return v.RouteMap
		}
	}
	return nil
}
//func (methodRouteMaps MethodRouteMaps)get(httpMethod string)*RouteMaps{
//	for _,routeMap :=range methodRouteMaps{
//		if routeMap.httpMethod==httpMethod{
//			return routeMap.routeMap
//		}
//	}
//	return nil
//}


//参数相关
func (p Params) Get(name string) (string, bool) {
	for _, value := range p {
		if value.Key == name {
			return value.Value, true
		}
	}
	return "", false
}

func (p Params) ByName(name string) string {
	value, _ := p.Get(name)
	return value
}

//存储路由相关
func (route *RouteMap)reset(){
	route.staticMaps=make(map[string]HandlerChain)
	route.catchAllMaps=make(map[string]HandlerChain)
	route.paramMaps=make(map[string]ParamMap)
}

func (route *RouteMap)addRoute(path string,handlers HandlerChain){
	route.Lock()
	if route.isRawCatchAllRoute(path){
		index:=strings.Index(path,"*")
		route.catchAllMaps[path[:index-1]]=handlers
	}else if ok,pre:=route.isRawParamRoute(path);ok{
		paths:=strings.Split(path,"/")
		for k,v:=range paths{
			if v==""{
				continue
			}
			if v[0]==':'{
				paths[k]=":"
			}
		}
		path=strings.Join(paths,"/")
		route.paramMaps[path]=ParamMap{
			pathParam: pre,
			handlers:  handlers,
		}
	}else {
		route.staticMaps[path] = handlers
	}
	route.Unlock()
}

func (route *RouteMap)isRawParamRoute(path string)(flag bool,paramMap map[int]string){
	paths:=strings.Split(path,"/")
	flag=false
	paramMap=make(map[int]string,len(paths))
	for k,v:=range paths{
		if strings.HasPrefix(v,":")&&strings.Count(v,":")<2{
			paramMap[k]=v
			flag=true
		}
	}
	if flag{
		return flag,paramMap
	}
	return false,nil
}

func (route *RouteMap)isRawCatchAllRoute(path string)bool{
	if strings.Count(path,"*")>1||!strings.Contains(path,"*"){
		return false
	}
	index:=strings.Index(path,"*")
	if path[index-1]=='/'&&len(path)==index+1{
		return true
	}
	return false
}
//func (routeMaps *RouteMaps) addRouteMap(path string, handlers HandlerChain){
//	var routeMap RouteMap
//	routeMap.nodeMap=make(map[string]HandlerChain,1)
//	if strings.Contains(path,":")&&strings.Contains(path,"*"){
//		PrintErr("the route is wrong")
//	}
//	if pre,ok:=routeMap.isAllRoute(path);ok{
//		routeMap.path=pre
//		routeMap.nodeType=catchAll
//		routeMap.nodeMap[pre]=handlers
//		*routeMaps=append(*routeMaps,routeMap)
//		return
//	}
//	if pre,ok,indexes:=routeMap.isParamRoute(path);ok{
//		routeMap.nodeType=param
//		routeMap.paramName=pre
//		routeMap.nodeMap[path]=handlers
//		routeMap.paramIndex=indexes
//		routeMap.path=path
//		*routeMaps=append(*routeMaps,routeMap)
//		return
//	}
//	routeMap.nodeType=static
//	routeMap.nodeMap[path]=handlers
//	*routeMaps=append(*routeMaps,routeMap)
//	fmt.Println(len(*routeMaps))
//}

//func (routeMap *RouteMap)isParamRoute(path string) ([]string, bool,[]int) { //返回是否有动态参数':',并返回此参数名称且是否合法
//	var paths []string
//	var pathIndex []int
//	for k, v := range strings.Split(path, "/") {
//		if strings.HasPrefix(v, ":") {
//			if strings.Count(v, ":") > 1 {
//				return nil, false,nil
//			} else {
//				paths = append(paths, v[1:])
//				pathIndex=append(pathIndex,k)
//			}
//		}
//	}
//	if len(paths) > 0 {
//		return paths, true,pathIndex
//	}
//	return nil, false,nil
//}

//func (routeMap *RouteMap)isAllRoute(path string) (string, bool) { //返回是否有'*'通配符若有是否合法,且返回通配符前面路径包括'/'
//     if strings.Count(path,"*")>1{
//     	return "",false
//	 }
//     if strings.Contains(path,"*"){
//     	index:=strings.Index(path,"*")
//     	if path[index-1]=='/'&&len(path)==index+1{
//     		return path[:index],true
//		}
//	 }
//     return "",false
//}





//获取路由处理函数相关
//这里的path是请求的path,params是context请求的参数
func (route *RouteMap)getValue(path string,params Params)(value RouteValue){
	value.params=params
	catchAllKeys:=GetAllKeys(route.catchAllMaps)
	var catchAllMins []string
	for k,v:=range catchAllKeys{
		catchAllMin:=v
		for _,v:=range catchAllKeys[k+1:]{
			if len(catchAllMin)>len(v)&&strings.HasPrefix(catchAllMin,v){
				catchAllMin=v
			}
		}
		catchAllMins=append(catchAllMins,catchAllMin)
	}
	for _,v:=range catchAllMins{
		if strings.HasPrefix(path,v+"/"){
			route.RLock()
			value.path=path
			value.handlers=route.catchAllMaps[v]
			route.RUnlock()
			return
		}
	}
	paths:=strings.Split(path,"/")
	paramKeys:=GetParamAllKeys(route.paramMaps)
	for _,v:=range paramKeys{
		paramPaths:=strings.Split(v,"/")
		if len(paramPaths)==len(paths){
			flag:=true
			for kk,vv:=range paramPaths{
				if vv==":"{
					continue
				}
				if vv!=paths[kk]{
					flag=false
				}
			}
			if flag{
				route.RLock()
				value.path=v
				value.handlers=route.paramMaps[v].handlers
				for k,v:=range route.paramMaps[v].pathParam{
					if v!=""{
						value.params=append(value.params,Param{
							Key:   v[1:],
							Value: paths[k],
						})
					}
				}
				route.RUnlock()
				return
			}
		}
	}
	value.path=path
	value.handlers=route.staticMaps[path]
	return
}

//func (routeMaps RouteMaps)getValue(path string,params Params)(value RouteValue){
//	value.params=params
//	//得到AllRoute的最小前缀
//	var catchAllMins   []string
//	for k,v:=range routeMaps{
//		if v.nodeType==catchAll{
//			catchAllMin:=v.path
//			for _,i:=range routeMaps[k+1:]{
//				if i.nodeType==catchAll&&strings.HasPrefix(v.path,i.path){
//					catchAllMin=i.path
//				}
//			}
//			catchAllMins=append(catchAllMins,catchAllMin)
//		}
//	}
//	for _,v:=range routeMaps{
//		if v.nodeMap==nil{
//			continue
//		}
//		for _,catchAllMin:=range catchAllMins{
//			if strings.HasPrefix(path,catchAllMin){
//				value.path=path
//				value.handlers=v.nodeMap[catchAllMin]
//				return
//			}
//		}
//		switch v.nodeType {
//		case static:
//			if path==v.path {
//				value.path = path
//				value.handlers = v.nodeMap[path]
//				return
//			}
//		case param:
//			//分段后的数组大小相同
//			//分段后的除参数下标外的字符相同
//			//确认是原路由之后将路由变为原路由Map取值
//			paths:=strings.Split(path,"/")
//			rePaths:=strings.Split(v.path,"/")
//			if len(paths)==len(rePaths){
//				flag:=true
//				for i:=0;i<len(paths);i++{
//					if ContainInt(v.paramIndex,i){
//						continue
//					}
//					if paths[i]!=rePaths[i]{
//						flag=false
//					}
//				}
//				if flag{
//					value.path=path
//					for i:=0;i<len(paths);i++{
//						if ContainInt(v.paramIndex,i){
//							value.params=append(value.params,Param{
//								Key:   v.paramName[i],
//								Value: paths[i],
//							})
//							paths[i]=":"+v.paramName[i]
//						}
//					}
//					path=strings.Join(paths,"/")
//					value.handlers=v.nodeMap[path]
//				}
//			}
//		}
//	}
//	return
//}