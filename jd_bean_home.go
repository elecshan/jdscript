package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/bitly/go-simplejson"
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

func (c *JdBeanHome) request(function_id string, body map[string]interface{}, method string) *simplejson.Json {
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

	// res := make(map[string]interface{})
	res, err := simplejson.NewJson(bodyResp)
	if err != nil {
		panic(err)
	}
	return res
}

func (c *JdBeanHome) get_award(source string) {
	body := c.request("beanHomeTask", map[string]interface{}{"source": source, "awardFlag": true}, "GET")
	code, _ := body.Get("code").String()
	if _, ok := body.CheckGet("errorCode"); code == "0" && !ok {
		num, _ := body.Get("data").Get("beanNum").String()
		log.Printf("领取京豆奖励, 获得京豆:%s!\n", num)
	} else {
		msg := "未知"
		_, ok := body.CheckGet("errorMessage")
		if ok {
			msg, _ = body.Get("errorMessage").String()
		}
		log.Printf("领取京豆奖励失败，%s!\n", msg)
	}
}

func (c *JdBeanHome) do_task() {
	body := c.request("findBeanHome", map[string]interface{}{"source": "AppHome", "rnVersion": "4.7", "rnClient": "2"}, "POST")

	if code, _ := body.Get("code").String(); code != "0" {
		log.Print("获取首页数据失败！")
		return
	}

	taskProgress, _ := body.Get("data").Get("taskProgress").Int()
	taskThreshold, _ := body.Get("data").Get("taskThreshold").Int()
	s, _ := body.Get("data").Get("taskProgress").String()
	fmt.Println(s)
	fmt.Println(taskProgress, taskThreshold)
	if taskProgress == taskThreshold {
		log.Print("今日已完成领额外京豆任务!")
		return
	}

	for i := 1; i < 6; i++ {
		bodySub := c.request("beanHomeTask", map[string]interface{}{"type": fmt.Sprintf("%d", i), "source": "home", "awardFlag": false, "itemId": ""}, "GET")
		code, _ := bodySub.Get("code").String()
		if _, ok := bodySub.CheckGet("errorCode"); code == "0" && !ok {
			taskProgress, _ := body.Get("data").Get("taskProgress").Int()
			taskThreshold, _ := body.Get("data").Get("taskThreshold").Int()
			log.Printf("领额外京豆任务进度:%d/%d!\n", taskProgress, taskThreshold)
		} else {
			msg := "未知"
			_, ok := body.CheckGet("errorMessage")
			if ok {
				msg, _ = body.Get("errorMessage").String()
			}
			log.Printf("第%d个领额外京豆任务完成失败, %s!\n", i, msg)
		}
	}
}

func (c *JdBeanHome) do_goods_task() {
	body := c.request("homeFeedsList", map[string]interface{}{"page": 1}, "GET")

	code, _ := body.Get("code").String()
	if _, ok := body.CheckGet("errorCode"); code != "0" || ok {
		log.Printf("无法浏览商品任务!\n")
	}

	taskProgress, _ := body.Get("data").Get("taskProgress").Int()
	taskThreshold, _ := body.Get("data").Get("taskThreshold").Int()
	if taskProgress == taskThreshold {
		log.Print("今日已完成浏览商品任务!")
		return
	}

	for i := 0; i < 3; i++ {
		b := map[string]interface{}{
			"skuId":     fmt.Sprintf("%d", 10000000+rand.Intn(20000000-10000000)),
			"awardFlag": false,
			"type":      "1",
			"source":    "feeds",
			"scanTime":  int(time.Now().Second() * 1000),
		}
		body := c.request("beanHomeTask", b, "GET")

		if msg, ok := body.CheckGet("errorCode"); ok {
			msgStr, _ := msg.String()
			log.Printf("浏览商品任务, %s!\n", msgStr)
			if msgStr == "HT203" {
				break
			}
		} else {
			taskProgress, _ := body.Get("data").Get("taskProgress").Int()
			taskThreshold, _ := body.Get("data").Get("taskThreshold").Int()
			log.Printf("完成浏览商品任务, 进度:%d/%d!\n", taskProgress, taskThreshold)
		}

		time.Sleep(time.Duration(2) * time.Second)
	}
}

func run() {
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

	c.do_task()
	// c.get_award("feeds")
	// c.do_goods_task()
	// c.get_award("feeds")
}

// func main() {
// 	run()
// }
