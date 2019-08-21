package responsedata

import (
	"encoding/json"
	"log"
	"net/http"

)

type JSON struct {
	Data interface{}
}


var jsonContentType = []string{"application/json; charset=utf-8"}


func (r JSON) ResponseData(w http.ResponseWriter) (err error) {
	if err = WriteJSON(w, r.Data); err != nil {
		log.Fatalf("%s : %s",err,"response the json failed")
	}
	return
}


func (r JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	writeContentType(w, jsonContentType)
	encoder := json.NewEncoder(w)
	err := encoder.Encode(&obj)
	return err
}