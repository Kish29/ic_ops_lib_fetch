package util

import (
	"errors"
	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
	"log"
	"net/http"
)

const (
	HttpHeadKeAccept      = "Accept"
	HttpHeadKeContentType = "Content-Type"
	HttpHeadKeyUserAgent  = "User-Agent"
	HttpHeadValJson       = "application/json;charset=UTF-8"
)

func HttpGETToJson(rc *resty.Client, url string, queryParams, headerAttr map[string]string, resultJ interface{}) (err error) {
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

func HttpRawGET(rc *resty.Client, url string, queryParams, headerAttr map[string]string) (body string, err error) {
	if rc == nil {
		return "", errors.New("client is nil")
	}
	var resp *resty.Response
	resp, err = rc.R().SetQueryParams(queryParams).SetHeaders(headerAttr).Get(url)
	if err != nil {
		return "", err
	}
	return Bytes2Str(resp.Body()), nil
}

func HttpGETNode(url string) *html.Node {
	log.Println("Fetch Url", url)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", RandomFakeAgent())
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[error] Http get err=>%v", err)
		return nil
	}
	if resp.StatusCode != 200 {
		log.Printf("[error] Http status code=>%v", resp.StatusCode)
		return nil
	}
	defer resp.Body.Close()
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Printf("[error] parse error, err=>%v", err)
		return nil
	}
	return doc
}
