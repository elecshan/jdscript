package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/bitly/go-simplejson"
	"github.com/elecshan/jdscript/utils"
	httpHeaders "github.com/go-http-utils/headers"
	"github.com/klauspost/compress/flate"
)

type JdFruit struct {
	client  http.Client
	headers map[string]string
}

func request(url string, body []byte, headers map[string]string, method string) *http.Response {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	if len(headers) != 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func collect() {

}

func initFarm(method string) []byte {
	url := utils.JD_API_HOST + "?functionId=initForFarm"
	bodyReq := []byte(`{ "version": 14 }`)
	headers := map[string]string{
		"host":            "me-api.jd.com",
		"accept":          "*/*",
		"connection":      "keep-alive",
		"user-agent":      "jdapp;android;10.0.2;10;network/wifi;Mozilla/5.0 (Linux; Android 10; ONEPLUS A5010 Build/QKQ1.191014.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045230 Mobile Safari/537.36",
		"accept-language": "zh-cn",
		"referer":         "https://home.m.jd.com/myJd/newhome.action?sceneval=2&ufc=&",
		"accept-encoding": "gzip, deflate, br",
	}

	resp := request(url, bodyReq, headers, method)
	var bodyResp []byte
	switch resp.Header.Get(httpHeaders.ContentEncoding) {
	case "br":
		bodyResp, _ = ioutil.ReadAll(brotli.NewReader(resp.Body))
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			panic(err)
		}
		bodyResp, _ = ioutil.ReadAll(gr)
	case "deflate":
		zr := flate.NewReader(resp.Body)
		defer zr.Close()
		bodyResp, _ = ioutil.ReadAll(zr)
	}
	return bodyResp
}

func totalBean(method string) {
	url := "https://me-api.jd.com/user_new/info/GetJDUserInfoUnion"

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("host", "me-api.jd.com")
	req.Header.Add("accept", "*/*")
	req.Header.Add("connection", "keep-alive")
	req.Header.Add("user-agent", "jdapp;android;10.0.2;10;network/wifi;Mozilla/5.0 (Linux; Android 10; ONEPLUS A5010 Build/QKQ1.191014.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045230 Mobile Safari/537.36")
	req.Header.Add("accept-language", "zh-cn")
	req.Header.Add("referer", "https://home.m.jd.com/myJd/newhome.action?sceneval=2&ufc=&")
	req.Header.Add("accept-encoding", "gzip, deflate, br")

	req.AddCookie(&http.Cookie{
		Name:  "pt_pin",
		Value: "xgs951230_m",
	})

	req.AddCookie(&http.Cookie{
		Name:  "pt_key",
		Value: "AAJhLkGsADDGjNiUL6TgYqbkEvmveWHiFsqkihWTgTH84QAxfjJBXUD3LVrhKb7NHTj8k6ihh0c",
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var bodyResp []byte
	switch resp.Header.Get(httpHeaders.ContentEncoding) {
	case "br":
		bodyResp, _ = ioutil.ReadAll(brotli.NewReader(resp.Body))
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			panic(err)
		}
		bodyResp, _ = ioutil.ReadAll(gr)
	case "deflate":
		zr := flate.NewReader(resp.Body)
		defer zr.Close()
		bodyResp, _ = ioutil.ReadAll(zr)
	}

	// res := make(map[string]interface{})
	res, err := simplejson.NewJson(bodyResp)
	if err != nil {
		panic(err)
	}

	retcode, _ := res.Get("retcode").String()
	if retcode == "1001" {
		log.Println("Cookie已过期！")
		return
	}
	if usrInfo, ok := res.Get("data").CheckGet("userInfo"); retcode == "0" && ok {
		nickname, _ := usrInfo.Get("baseInfo").Get("nickname").String()
		fmt.Println(nickname)
	}
	if assetInfo, ok := res.Get("data").CheckGet("assetInfo"); retcode == "0" && ok {
		beanNum, _ := assetInfo.Get("beanNum").String()
		fmt.Println(beanNum)
	}
	// r, _ := res.String()
	// fmt.Println(res)
}

func main() {
	// totalBean("GET")
	var str interface{}
	data := []byte("{ \"version\": 14 }")
	fmt.Print(data)
	json.Unmarshal(data, &str)

	fmt.Print(str)
}
