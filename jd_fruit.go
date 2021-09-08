package main

import (
	"bytes"
	"compress/gzip"
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
	client   http.Client
	farmInfo *simplejson.Json
	farmTask *simplejson.Json
}

func Encode(resp *http.Response) []byte {
	var bodyResp []byte
	defer resp.Body.Close()
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

func request(url string, body []byte, headers map[string]string, cookies []*http.Cookie, method string) *http.Response {
	if body == nil {
		body = []byte("")
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	if len(headers) != 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	if len(cookies) != 0 {
		for idx := range cookies {
			req.AddCookie(cookies[idx])
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func (c *JdFruit) collect() {
	go c.initFarm("POST")
}

func (c *JdFruit) jdFruit() {
	go c.initFarm("POST")
	if farmUserPro, ok := c.farmInfo.CheckGet("farmUserPro"); ok {
		shareCode, _ := farmUserPro.Get("shareCode").String()
		url := fmt.Sprintf(`https://api.sharecode.ga/api/runTimes?activityId=farm&sharecode=%s`, shareCode)
		request(url, nil, nil, nil, "GET")
	}

	treeState, _ := c.farmInfo.Get("treeState").Int()
	if treeState == 2 || treeState == 3 {
		log.Println("水果已可领取")
		return
	} else if treeState == 1 {
		log.Println("种植中")
	} else if treeState == 0 {
		log.Println("未种植水果")
		return
	}
	c.doDailyTask()
}

func (c *JdFruit) doDailyTask() {
	// {
	log.Println("初始化任务列表")
	// functionId := `taskInitForFarm`
	url := utils.JD_API_HOST + `?functionId=taskInitForFarm`
	bodyReq := []byte(`body={ "version": 14, "channel": 1, "babelChannel": "45" }&appid=wh5`)
	headers := map[string]string{
		"user-agent": "jdapp;android;10.0.2;10;network/wifi;Mozilla/5.0 (Linux; Android 10; ONEPLUS A5010 Build/QKQ1.191014.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045230 Mobile Safari/537.36",
	}
	cookies := []*http.Cookie{
		{
			Name:  "pt_pin",
			Value: "xgs951230_m",
		},
		{
			Name:  "pt_key",
			Value: "AAJhLkGsADDGjNiUL6TgYqbkEvmveWHiFsqkihWTgTH84QAxfjJBXUD3LVrhKb7NHTj8k6ihh0c",
		},
	}
	resp := request(url, bodyReq, headers, cookies, "GET")
	// Encode(resp)
	// c.farmTask, _ = simplejson.NewJson(resp)
	// }
	fmt.Print(resp.Header)
	// fmt.Print(c.farmTask)
}

func (c *JdFruit) initFarm(method string) {
	url := utils.JD_API_HOST + "?functionId=initForFarm"
	bodyReq := []byte(`body={ "version": 14 }&appid=wh5&clientVersion=9.1.0`)
	headers := map[string]string{
		"accept":          "*/*",
		"connection":      "keep-alive",
		"user-agent":      "jdapp;android;10.0.2;10;network/wifi;Mozilla/5.0 (Linux; Android 10; ONEPLUS A5010 Build/QKQ1.191014.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045230 Mobile Safari/537.36",
		"accept-language": "zh-CN,zh;q=0.9",
		"accept-encoding": "gzip, deflate, br",
		"cache-control":   "no-cache",
		"origin":          "https://home.m.jd.com",
		"pragma":          "no-cache",
		"referer":         "https://home.m.jd.com/myJd/newhome.action",
		"sec-fetch-dest":  "empty",
		"sec-fetch-mode":  "cors",
		"sec-fetch-site":  "same-site",
		"Content-Type":    "application/x-www-form-urlencoded",
	}

	cookies := []*http.Cookie{
		{
			Name:  "pt_pin",
			Value: "xgs951230_m",
		},
		{
			Name:  "pt_key",
			Value: "AAJhLkGsADDGjNiUL6TgYqbkEvmveWHiFsqkihWTgTH84QAxfjJBXUD3LVrhKb7NHTj8k6ihh0c",
		},
	}

	resp := request(url, bodyReq, headers, cookies, method)
	bodyResp := Encode(resp)
	c.farmInfo, _ = simplejson.NewJson(bodyResp)
}

func (c *JdFruit) totalBean(method string) {
	url := "https://me-api.jd.com/user_new/info/GetJDUserInfoUnion"
	bodyReq := []byte(``)
	headers := map[string]string{
		"host":            "me-api.jd.com",
		"accept":          "*/*",
		"connection":      "keep-alive",
		"user-agent":      "jdapp;android;10.0.2;10;network/wifi;Mozilla/5.0 (Linux; Android 10; ONEPLUS A5010 Build/QKQ1.191014.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045230 Mobile Safari/537.36",
		"accept-language": "zh-cn",
		"referer":         "https://home.m.jd.com/myJd/newhome.action?sceneval=2&ufc=&",
		"accept-encoding": "gzip, deflate, br",
	}
	cookies := []*http.Cookie{
		{
			Name:  "pt_pin",
			Value: "xgs951230_m",
		},
		{
			Name:  "pt_key",
			Value: "AAJhLkGsADDGjNiUL6TgYqbkEvmveWHiFsqkihWTgTH84QAxfjJBXUD3LVrhKb7NHTj8k6ihh0c",
		},
	}

	resp := request(url, bodyReq, headers, cookies, method)

	bodyResp := Encode(resp)

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
	c := new(JdFruit)
	c.totalBean("GET")
	c.initFarm("POST")
	// fmt.Print(c.farmInfo)
	c.jdFruit()
}
