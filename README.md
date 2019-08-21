
# Rego 轻量级Web框架 
此框架是用golang编写的一个轻量级Web框架,大多数Web框架功能已经实现其中包括WebSocket功能也有相应的实现, 框架借鉴了gin框架的设计架构,且框架的使用方法与gin相比类似,具有RestFul的通信风格,使用起来比较简单,其内 部保存和获取路由采用的HashMap来实现的其效率在处理动态参数时可能没有使用树的效果好

## 使用

其开启一个Http服务,只需要几行代码便可以实现

```go
	r:=rego.New()
	r.GET("/echo", func(context *rego.Context) {
	  context.String(200,"hello world")
	})
	r.Run(":8080")
```

### 路由

- 基本路由

  ```GO
  	r:=rego.New()
  	r.GET("/someGet", getting)
  	r.POST("/somePost", posting)
  	r.PUT("/somePut", putting)
  	r.DELETE("/someDelete", deleting)
  	r.PATCH("/somePatch", patching)
  	r.HEAD("/someHead", head)
  	r.OPTIONS("/someOptions", options)
  	r.Run(":8080")
  ```

- 路由参数

  - **api参数**可通过Context的Param方法获取

  ```go
  router.GET("/string/:id", func(c *rego.Context) {
      	name := c.Param("id")
      	fmt.Println(name)
      })
  ```

  - **Url参数**可通过Query方法获取,也可以通过DefaultQuery方法设置默认参数

  ```go
  router.GET("/getParam", func(c *gin.Context) {
  	name := c.DefaultQuery("name", "Guest") 
  	lastname := c.Query("lastname") 
  	fmt.Println("My name is", name,lastname)
  })
  ```

  - **表单参数**通过PostForm方法获取,也可通过DefaultForm方法设置默认参数

  ```go
  router.POST("/getForm", func(c *gin.Context) {
  	type := c.DefaultPostForm("name", "type")//可设置默认值
  	msg := c.PostForm("msg")
  	title := c.PostForm("title")
  	fmt.Println("type is %s, msg is %s, title is %s", type, msg, title)
  })
  ```

- 路由群组

  ```GO
  	rr:=r.Group("/y")
  	{
  		rr.GET("/a", func(context *rego.Context) {
  			context.String(200,"ni hao")
  		})
  		rr.GET("/id", func(context *rego.Context) {
  			context.String(200,"ni h")
  		})
  	}
  ```

### 请求

- 请求头

- 请求参数

  ```go
  c.Set("test","1234")
  c.Get("test")
  //当然利用断言可以获取其他类型
  c.GetString("test")
  ...
  ```

- 上传文件

- Cookies

```go
	r:=rego.New()
	r.POST("/upload", func(context *rego.Context) {
		file,header,err:=context.Request.FormFile("upload")
		fileName:=header.Filename
		out, err := os.Create("./"+fileName+".png")
		defer out.Close()
		if err != nil {
			log.Fatalf("%s : create the file failed",err)
		}
		_, err = io.Copy(out, file)
		if err != nil {
			log.Fatalf("%s : copy the file failed",err)
		}
	})
	r.Run(":8080")
```



### 响应

- 字符串响应

  ```go
  c.String(200,"hello!")
  ```

- JSON/XML响应

  ```go
  	var msg struct {
  		Name    string `json:"name"`
  		Message string
  		Number  int
  	}
  	msg.Name = "Lena"
  	msg.Message = "hey"
  	msg.Number = 123
  	r:=rego.New()
  	r.GET("/moreJSON", func(c *rego.Context) {
  		c.JSON(http.StatusOK, gin.H{"user": "Lena", "Message": "hey", "Number": 123})
  		c.XML(http.StatusOK, gin.H{"user": "Lena", "Message": "hey", "Number": 123})
  		c.JSON(http.StatusOK, msg)
  	})
  	r.Run(":8080")
  ```

- 重定向

  ```go
  	r:=rego.New()
  	r.GET("/Redirect", func(c *rego.Context) {
  		c.Redirect(http.StatusMovedPermanently, "http://129.28.185.138/")
  	})
  	r.Run(":8080")
  ```

- 附加Cookie

### 中间件

- 自定义中间件

  ```go
  func middle()rego.HandlerFunc{
  	return func(context *rego.Context) {
  		context.Set("test","test")
  	}
  }
  
  
  func main(){
  
  	r:=rego.New()
  	r.Use(middle())
  	r.GET("/TEST", func(context *rego.Context) {
  		if value,ok:=context.Get("test");ok{
  			fmt.Println(value)
  		}
  	})
  	r.GET("/te", func(context *rego.Context) {
  		fmt.Println("test")
  	})
  	r.Run(":8080")
  }
  ```

- WebSocket

  ```GO
  	r:=rego.New()
  	r.WebSocket("/test",rego.WSConfig{
  		OnOpen: func(con *rego.WSCon) {
  			fmt.Println("open the connection")
  		},
  		OnClose: func(con *rego.WSCon) {
  			fmt.Println("close the connection")
  		},
  		OnMessage: func(con *rego.WSCon, data []byte) {
  			fmt.Println(con.Conn.RemoteAddr(),"receive the message:",string(data))
  			con.SendIframe([]byte("ni hao ya "))
  		},
  		OnError: func(con *rego.WSCon) {
  			con.Conn.Close()
  		},
  	})
  	r.Run(":8080")
  ```

  




 

 