package responsedata

import "net/http"

type ResponseData interface {
	ResponseData( http.ResponseWriter)error
	WriteContentType(w http.ResponseWriter)
}

var (
	_ ResponseData     = String{}
	_ ResponseData     = JSON{}
	_ ResponseData     = Redirect{}
	_ ResponseData     = XML{}
)



func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
