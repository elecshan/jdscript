package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

func random_string() string {
	t := "0123456789"
	var s []byte

	for {
		s = append(s, t[rand.Intn(len(t))])
		if len(s) >= 16 {
			break
		}
	}

	return string(s)
}

type JdBeanHome struct {
	client  http.Client
	headers map[string]string
	eu      string
	fv      string
}

func (c *JdBeanHome) request(function_id string, body map[string]interface{}, method string) map[string]interface{} {
	params := url.Values{}
	params.Add("functionId", function_id)
	params.Add("appid", "ld")
	params.Add("clientVersion", "10.1.2")
	params.Add("client", "apple")
	params.Add("eu", c.eu)
	params.Add("fv", c.fv)
	params.Add("osVersion", "11")
	params.Add("uuid", c.eu+c.fv)
	params.Add("openudid", c.eu+c.fv)
	bodyJson, _ := json.Marshal(body)
	params.Add("body", string(bodyJson))

	url := "https://api.m.jd.com/client.action?" + params.Encode()
	// log.Print(url)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	res := make(map[string]interface{})
	err = json.Unmarshal(bodyResp, &res)
	return res
}

func (c *JdBeanHome) get_award(source string) {
	body := c.request("beanHomeTask", map[string]interface{}{"source": source, "awardFlag": true}, "GET")
	if _, ok := body["errorCode"]; body["code"] == "0" && !ok {
		log.Printf("领取京豆奖励, 获得京豆:%s!\n", body["data"])
	} else {
		log.Print("领取京豆奖励失败")
	}
}

func main() {
	cookies := make([]*http.Cookie, 0)
	cookies = append(cookies, &http.Cookie{
		Name:  "pt_pin",
		Value: "xgs951230_m",
	})
	cookies = append(cookies, &http.Cookie{
		Name:  "pt_key",
		Value: "AAJhLkGsADDGjNiUL6TgYqbkEvmveWHiFsqkihWTgTH84QAxfjJBXUD3LVrhKb7NHTj8k6ihh0c",
	})

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(nil)
	}

	url, err := url.Parse("https://api.m.jd.com/")
	if err != nil {
		panic(nil)
	}

	jar.SetCookies(url, cookies)
	c := &JdBeanHome{
		client: http.Client{Jar: jar},
		headers: map[string]string{
			"referer":    "https://h5.m.jd.com/rn/2E9A2bEeqQqBP9juVgPJvQQq6fJ/index.html",
			"user-agent": "jdapp;android;10.0.2;10;network/wifi;Mozilla/5.0 (Linux; Android 10; ONEPLUS A5010 Build/QKQ1.191014.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045230 Mobile Safari/537.36",
		},
		eu: random_string(),
		fv: random_string(),
	}
	c.get_award("home")
}
