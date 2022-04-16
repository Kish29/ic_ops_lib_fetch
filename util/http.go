package util

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"log"
)

const (
	HttpHeadKeAccept      = "Accept"
	HttpHeadKeContentType = "Content-Type"
	HttpHeadKeyUserAgent  = "User-Agent"
	HttpHeadValJson       = "application/json;charset=UTF-8"
)

func HttpGet2Json(rc *resty.Client, url string, queryParams, headerAttr map[string]string, resultJ interface{}) (err error) {
	if rc == nil {
		return errors.New("client is nil")
	}
	_, err = rc.R().
		SetResult(resultJ).
		SetJSONEscapeHTML(false).
		SetQueryParams(queryParams).
		SetHeader(HttpHeadKeAccept, HttpHeadValJson).
		SetHeaders(headerAttr).
		Get(url)
	if err != nil {
		log.Printf("http get error, error=>%v", err)
	}
	return
}
