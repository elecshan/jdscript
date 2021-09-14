package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
)

type JdSuperBrand struct {
	activityId       uint64
	activityName     string
	encryptProjectId string
	hasAct           bool
	taskList         []*simplejson.Json
}

const signeid = "zFayjeUTzZWJGwv2rVNWY4DNAQw"
const signactid = 1000021
const signenpid = "uK2fYitTgioETuevoY88bGEts3U"
const signdataeid = "47E6skJcyZx7GSUFXyomLgF1FLCA"

func (c *JdSuperBrand) run() {
	log.Printf("开始京东账号【%s】\n", "")
	c.getId("superBrandSecondFloorMainPage", "secondfloor")
	if !c.hasAct {
		log.Println("今天没活动,退出")
		return
	}
	if len(c.encryptProjectId) != 0 {
		c.getCode("secondfloor")
	}
}

func (c *JdSuperBrand) getId(functionId string, source string) {
	url := fmt.Sprintf("https://api.m.jd.com/api?functionId=%s&appid=ProductZ4Brand&client=wh5&t=%d&body=%s", functionId, time.Now().Unix()*1000, []byte(fmt.Sprintf(`{"source":"%s"}`, source)))
	headers := map[string]string{
		"Accept":          "application/json,text/plain, */*",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "zh-cn",
		"Connection":      "keep-alive",
		"Host":            "api.m.jd.com",
		"Referer":         "https://prodev.m.jd.com/mall/active/NrHM6Egy96gxeG4eb7vFX7fYXf3/index.html?activityId=1000007&encryptProjectId=cUNnf3E6aMLQcEQbTVxn8AyhjXb&assistEncryptAssignmentId=2jpJFvC9MBNC7Qsqrt8WzEEcVoiT&assistItemId=S5ijz_8ukVww&tttparams=GgS7lUeyJnTGF0IjoiMzMuMjUyNzYyIiwiZ0xuZyI6IjEwNy4xNjA1MDcifQ6%3D%3D&lng=107.147022&lat=33.255229&sid=e5150a3fdd017952350b4b41294b145w&un_area=27_2442_2444_31912",
		"User-Agent":      "jdapp;android;9.4.4;10;3b78ecc3f490c7ba;network/UNKNOWN;model/M2006J10C;addressid/138543439;aid/3b78ecc3f490c7ba;oaid/7d5870c5a1696881;osVer/29;appBuild/85576;psn/3b78ecc3f490c7ba|541;psq/2;uid/3b78ecc3f490c7ba;adk/;ads/;pap/JA2015_311210|9.2.4|ANDROID 10;osv/10;pv/548.2;jdv/0|iosapp|t_335139774|appshare|CopyURL|1606277982178|1606277986;ref/com.jd.lib.personal.view.fragment.JDPersonalFragment;partner/xiaomi001;apprpd/MyJD_Main;Mozilla/5.0 (Linux; Android 10; M2006J10C Build/QP1A.190711.020; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045227 Mobile Safari/537.36",
	}
	respChan := make(chan *http.Response)
	go request(url, nil, headers, cookies, "POST", respChan)
	resp := <-respChan
	respJson, _ := simplejson.NewJson(Encode(resp))
	code, _ := respJson.Get("code").String()
	res := respJson.Get("data").Get("result")
	if code == "0" && res != nil {
		if info, ok := res.CheckGet("activityBaseInfo"); ok {
			c.activityId, _ = info.Get("activityId").Uint64()
			c.activityName, _ = info.Get("activityName").String()
			c.encryptProjectId, _ = info.Get("encryptProjectId").String()
			log.Printf("当前活动：%s  %d\n", c.activityName, c.activityId)
		}
		c.hasAct = true
	} else {
		log.Println("获取失败")
		c.hasAct = false
	}
}

func (c *JdSuperBrand) getCode(source string) {
	url := fmt.Sprintf("https://api.m.jd.com/api?functionId=superBrandTaskList&appid=ProductZ4Brand&client=wh5&t=%d&body=%s", time.Now().Unix()*1000, []byte(fmt.Sprintf(`{"source":"%s","activityId":%d,"assistInfoFlag":1}`, source, c.activityId)))
	headers := map[string]string{
		"Accept":          "application/json,text/plain, */*",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "zh-cn",
		"Connection":      "keep-alive",
		"Host":            "api.m.jd.com",
		"Referer":         "https://prodev.m.jd.com/mall/active/NrHM6Egy96gxeG4eb7vFX7fYXf3/index.html?activityId=1000007&encryptProjectId=cUNnf3E6aMLQcEQbTVxn8AyhjXb&assistEncryptAssignmentId=2jpJFvC9MBNC7Qsqrt8WzEEcVoiT&assistItemId=S5ijz_8ukVww&tttparams=GgS7lUeyJnTGF0IjoiMzMuMjUyNzYyIiwiZ0xuZyI6IjEwNy4xNjA1MDcifQ6%3D%3D&lng=107.147022&lat=33.255229&sid=e5150a3fdd017952350b4b41294b145w&un_area=27_2442_2444_31912",
		"User-Agent":      "jdapp;android;9.4.4;10;3b78ecc3f490c7ba;network/UNKNOWN;model/M2006J10C;addressid/138543439;aid/3b78ecc3f490c7ba;oaid/7d5870c5a1696881;osVer/29;appBuild/85576;psn/3b78ecc3f490c7ba|541;psq/2;uid/3b78ecc3f490c7ba;adk/;ads/;pap/JA2015_311210|9.2.4|ANDROID 10;osv/10;pv/548.2;jdv/0|iosapp|t_335139774|appshare|CopyURL|1606277982178|1606277986;ref/com.jd.lib.personal.view.fragment.JDPersonalFragment;partner/xiaomi001;apprpd/MyJD_Main;Mozilla/5.0 (Linux; Android 10; M2006J10C Build/QP1A.190711.020; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045227 Mobile Safari/537.36",
	}
	respChan := make(chan *http.Response)
	go request(url, nil, headers, cookies, "POST", respChan)
	resp := <-respChan
	respJson, _ := simplejson.NewJson(Encode(resp))
	code, _ := respJson.Get("code").String()
	log.Print(respJson)
	if code == "0" && source == "secondfloor" {
		res := respJson.Get("data").Get("result")
		if res != nil && res.Get("taskList") != nil {
			tasks := res.Get("taskList")
			taskArray, _ := tasks.Array()

			for i := 0; i < len(taskArray); i++ {
				task := tasks.GetIndex(i)
				t, _ := task.Get("assignmentType").Int()
				if t == 3 || t == 7 || t == 0 {
					c.taskList = append(c.taskList, task)
				}
			}
		}
	}
}

func main() {
	cookies = []*http.Cookie{
		{
			Name:  "pt_pin",
			Value: "xgs951230_m",
		},
		{
			Name:  "pt_key",
			Value: "AAJhLkGsADDGjNiUL6TgYqbkEvmveWHiFsqkihWTgTH84QAxfjJBXUD3LVrhKb7NHTj8k6ihh0c",
		},
	}
	c := new(JdSuperBrand)
	c.run()
}
