package main

import (
	"net/http"
	"rego"
)

func main(){
	r:=rego.New()
	r.GET("/", func(context *rego.Context) {
		context.String(http.StatusOK,"hello world")
	})
	_ = r.Run(":8080")
}