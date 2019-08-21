package responsedata

import (
	"log"
	"net/http"
)

type Redirect struct {
	Code     int
	Request  *http.Request
	FinalUrl string
}

func (r Redirect) ResponseData(w http.ResponseWriter) error {
	if (r.Code < http.StatusMultipleChoices || r.Code > http.StatusPermanentRedirect) && r.Code != http.StatusCreated {
		log.Fatalf("can't redirect with status code %d",r.Code)
	}
	http.Redirect(w, r.Request, r.FinalUrl, r.Code)
	return nil
}

func (r Redirect) WriteContentType(http.ResponseWriter) {}
