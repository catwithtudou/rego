package responsedata

import (
	"fmt"
	"io"
	"net/http"
)

type String struct {
	Format string
	Data   []interface{}
}

var plainContentType = []string{"text/plain; charset=utf-8"}

func (r String) ResponseData(w http.ResponseWriter) error {
	return WriteString(w, r.Format, r.Data)
}

func (r String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, plainContentType)
}

func WriteString(w http.ResponseWriter, format string, data []interface{}) (err error) {
	writeContentType(w, plainContentType)
	if len(data) > 0 {
		_, err = fmt.Fprintf(w, format, data...)
		return
	}
	_, err = io.WriteString(w, format)
	return
}