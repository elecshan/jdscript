package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/bitly/go-simplejson"
	"github.com/elecshan/jdscript/utils"
	httpHeaders "github.com/go-http-utils/headers"
	"github.com/klauspost/compress/flate"
)

type JdFruit struct {
	client                http.Client
	farmInfo              *simplejson.Json
	farmTask              *simplejson.Json
	signResult            *simplejson.Json
	goalResult            *simplejson.Json
	browseAdResult        *simplejson.Json
	browseAdRewardResult  *simplejson.Json
	threeMeal             *simplejson.Json
	friendList            *simplejson.Json
	waterFriendForFarmRes *simplejson.Json
}

var wg sync.WaitGroup

var cookies []*http.Cookie

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
	default:
		bodyResp, _ = ioutil.ReadAll(resp.Body)
	}
	return bodyResp
}

func request(url string, body []byte, headers map[string]string, cookies []*http.Cookie, method string, resp chan *http.Response) {
	if body == nil {
		body = []byte("")
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	// fmt.Println(req.URL)

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

	res, err := http.DefaultClient.Do(req)
	// log.Print("运行request结束！")
	if err != nil {
		log.Println("API请求失败，请检查网路重试")
	}
	resp <- res
}

func (c *JdFruit) collect() {
	c.initFarm("POST")
}

func (c *JdFruit) jdFruit() {
	c.initFarm("POST")
	if farmUserPro, ok := c.farmInfo.CheckGet("farmUserPro"); ok {
		shareCode, _ := farmUserPro.Get("shareCode").String()
		url := fmt.Sprintf(`https://api.sharecode.ga/api/runTimes?activityId=farm&sharecode=%s`, shareCode)
		respChan := make(chan *http.Response)
		defer close(respChan)
		go request(url, nil, nil, nil, "GET", respChan)
		<-respChan
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
	wg.Done()
}

func (c *JdFruit) doDailyTask() {
	log.Println("初始化任务列表")
	// functionId := `taskInitForFarm`
	url := utils.JD_API_HOST + `?functionId=taskInitForFarm&appid=wh5`
	bodyReq := []byte(`body={ "version": 14, "channel": 1, "babelChannel": "45" }`)
	headers := map[string]string{
		"user-agent": "jdapp;android;10.0.2;10;network/wifi;Mozilla/5.0 (Linux; Android 10; ONEPLUS A5010 Build/QKQ1.191014.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045230 Mobile Safari/537.36",
	}
	respChan := make(chan *http.Response)
	defer close(respChan)
	go request(url, bodyReq, headers, cookies, "GET", respChan)
	resp := <-respChan
	c.farmTask, _ = simplejson.NewJson(Encode(resp))
	// fmt.Print(c.farmTask)

	if ok, _ := c.farmTask.Get("signInit").Get("todaySigned").Bool(); !ok {
		// respChan := make(chan *http.Response)
		// defer close(respChan)
		functionId := `signForFarm`
		url := utils.JD_API_HOST + `?functionId=` + functionId + `&appid=wh5`
		go request(url, nil, nil, cookies, "GET", respChan)
		resp := <-respChan
		c.signResult, _ = simplejson.NewJson(Encode(resp))

		if ok, _ := c.signResult.Get("code").String(); ok == "0" {
			beanNum, _ := c.signResult.Get("amount").Int()
			log.Printf(`【签到成功】获得%dg💧\n`, beanNum)
		} else {
			log.Println(`签到结果： `, c.signResult)
		}
	} else {
		log.Printf(`今日已签到`)
	}
	canPop, _ := c.farmTask.Get("todayGotWaterGoalTask").Get("canPop").Bool()
	log.Printf(`被水滴砸中：`)
	if canPop {
		log.Println(`是`)
		url := utils.JD_API_HOST + `?functionId=gotWaterGoalTaskForFarm&appid=wh5`
		bodyReq := []byte(`body={type:3}`)
		go request(url, bodyReq, nil, cookies, "GET", respChan)
		resp := <-respChan
		c.goalResult, _ = simplejson.NewJson(Encode(resp))
		if code, _ := c.goalResult.Get("code").String(); code == "0" {
			water, _ := c.goalResult.Get("addEnergy").Int()
			log.Printf("被水滴砸中，获得%dg💧\n", water)
		}
	} else {
		log.Println(`否`)
	}

	log.Println(`签到结束，开始广告浏览`)
	browseReward := 0
	browseSuccess := 0
	browseFail := 0
	if f, _ := c.farmTask.Get("gotBrowseTaskAdInit").Get("f").Bool(); !f {
		log.Println(`今日已做过浏览广告任务`)
	} else {
		// log.Println(c.farmTask)
		// var ads []map[string]interface{}
		array, _ := c.farmTask.Get("gotBrowseTaskAdInit").Get("userBrowseTaskAds").Array()
		for i := 0; i < len(array); i++ {
			ad, _ := c.farmTask.Get("gotBrowseTaskAdInit").Get("userBrowseTaskAds").GetIndex(i).Map()
			// ads = append(ads, ad)
			if ad["limit"].(json.Number) <= ad["hadFinishedTimes"].(json.Number) {
				log.Printf("%s已完成\n", ad["mainTitle"])
				continue
			}
			log.Println(`正在浏览广告：`, ad["mainTitle"])
			url := utils.JD_API_HOST + `?functionId=browseAdTaskForFarm&appid=wh5`
			bodyReq := []byte(fmt.Sprintf(`body={%s,0,"version":14,"channel":1,"babelChannel":"45"}`, ad["advertId"]))
			go request(url, bodyReq, nil, cookies, "GET", respChan)
			resp := <-respChan
			c.browseAdResult, _ = simplejson.NewJson(Encode(resp))
			if code, _ := c.browseAdResult.Get("code").String(); code == "0" {
				fmt.Println(ad["mainTitle"], `浏览任务完成`)
				bodyResp := []byte(fmt.Sprintf(`body={%s,1,"version":14,"channel":1,"babelChannel":"45"}`, ad["advertId"]))
				go request(url, bodyResp, nil, cookies, "GET", respChan)
				resp := <-respChan
				c.browseAdRewardResult, _ = simplejson.NewJson(Encode(resp))
				if code, _ := c.browseAdRewardResult.Get("code").String(); code == "0" {
					amount, _ := c.browseAdRewardResult.Get("amount").Int()
					log.Printf("领取浏览广告%s奖励成功，获得%dg💧\n", ad["mainTitle"], amount)
					browseReward += amount
					browseSuccess++
				} else {
					browseFail++
					log.Printf("领取浏览广告%s奖励失败\n", ad["mainTitle"])
				}
			} else {
				browseFail++
				log.Println(ad["mainTitle"], `浏览任务失败,`, c.browseAdResult)
			}
		}

		log.Printf("广告浏览完成%d个，失败%d个，获得%dg💧\n", browseSuccess, browseFail, browseReward)
	}

	log.Println("开始定时领水任务")
	if f, _ := c.farmTask.Get("gotThreeMealInit").Get("f").Bool(); !f {
		log.Println("当前不在定时领水时间段内或者已经领取水滴")
	} else {
		url := utils.JD_API_HOST + `?functionId=gotThreeMealForFarm&appid=wh5`
		go request(url, nil, nil, cookies, "GET", respChan)
		resp := <-respChan
		c.threeMeal, _ = simplejson.NewJson(Encode(resp))
		if code, _ := c.threeMeal.Get("code").String(); code == "0" {
			amount, _ := c.threeMeal.Get("amount").Int()
			log.Printf("定时领水，获得%dg💧\n", amount)
		} else {
			log.Println("定时领水结果：", c.threeMeal)
		}
	}

	log.Println("开始给好友浇水")
	if f, _ := c.farmTask.Get("waterFriendTaskInit").Get("f").Bool(); !f {
		log.Println("给好友浇水任务已完成")
	} else {
		waterFriendCountKey, _ := c.farmTask.Get("waterFriendTaskInit").Get("waterFriendCountKey").Int()
		waterFriendMax, _ := c.farmTask.Get("waterFriendTaskInit").Get("waterFriendMax").Int()
		if waterFriendCountKey <= waterFriendMax {
			ch := make(chan struct{})
			go c.doFriendsWater(waterFriendMax-waterFriendCountKey, ch)
			<-ch
		} else {
			log.Printf("今日为好友浇水量已达%d个\n", waterFriendMax)
		}
	}

	log.Println("开始邀请好友任务")
}

func (c *JdFruit) doFriendsWater(length int, ch chan struct{}) {
	url := utils.JD_API_HOST + "?functionId=friendListInitForFarm&appid=wh5"
	bodyReq := []byte(`body={ "version": 4, "channel": 1 }`)
	respChan := make(chan *http.Response)
	go request(url, bodyReq, nil, cookies, "GET", respChan)
	resp := <-respChan
	c.friendList, _ = simplejson.NewJson(Encode(resp))
	// log.Print(c.friendList)
	var needWaterFriends []string
	friends, _ := c.friendList.Get("friends").Array()
	if len(friends) > 0 {
		for i := 0; i < len(friends); i++ {
			friend, _ := c.friendList.Get("friends").GetIndex(i).Map()
			if friend["friendState"] == 1 && len(needWaterFriends) < length {
				needWaterFriends = append(needWaterFriends, friend["shareCode"].(string))
			}
		}
	} else {
		log.Println("您的好友列表暂无好友,快去邀请您的好友吧!")
		ch <- struct{}{}
		return
	}
	// log.Println(needWaterFriends)
	waterFriendsCount := 0
	cardInfoStr := ""
	for i := 0; i < len(needWaterFriends); i++ {
		url := utils.JD_API_HOST + "?functionId=waterFriendForFarm&appid=wh5"
		bodyReq := []byte(fmt.Sprintf(`body = { "shareCode": %s, "version": 6, "channel": 1 }`, needWaterFriends[i]))
		go request(url, bodyReq, nil, cookies, "GET", respChan)
		resp := <-respChan
		c.waterFriendForFarmRes, _ = simplejson.NewJson(Encode(resp))
		log.Printf(`为第%d个玩家浇水结果：%s\n`, i+1, c.waterFriendForFarmRes)
		code, _ := c.waterFriendForFarmRes.Get("code").String()
		if code == "0" {
			waterFriendsCount++
			rule, _ := c.waterFriendForFarmRes.Get("cardInfo").Get("rule").String()
			switch info, _ := c.waterFriendForFarmRes.Get("cardInfo").Get("type").String(); info {
			case "beanCard":
				log.Println("获取道具卡：", rule)
				cardInfoStr += "水滴换豆卡，"
			case "fastCard":
				log.Println("获取道具卡：", rule)
				cardInfoStr += "快速浇水卡，"
			case "doubleCard":
				log.Println("获取道具卡：", rule)
				cardInfoStr += "水滴翻倍卡，"
			case "signCard":
				log.Println("获取道具卡：", rule)
				cardInfoStr += "加签卡，"
			}
		} else if code == "11" {
			log.Println("水滴不够,跳出浇水")
		}
	}
	log.Printf("【好友浇水】已给%d个好友浇水,消耗%dg水\n", waterFriendsCount, waterFriendsCount*10)
	if len(cardInfoStr) > 0 {
		log.Println("【好友浇水奖励】", cardInfoStr[0:len(cardInfoStr)-1])
	}

	ch <- struct{}{}
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

	respChan := make(chan *http.Response)
	defer close(respChan)
	go request(url, bodyReq, headers, cookies, method, respChan)
	resp := <-respChan
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

	respChan := make(chan *http.Response)
	defer close(respChan)
	go request(url, bodyReq, headers, cookies, method, respChan)
	resp := <-respChan
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
	wg.Done()
}

// func main() {
// 	cookies = []*http.Cookie{
// 		{
// 			Name:  "pt_pin",
// 			Value: "xgs951230_m",
// 		},
// 		{
// 			Name:  "pt_key",
// 			Value: "AAJhLkGsADDGjNiUL6TgYqbkEvmveWHiFsqkihWTgTH84QAxfjJBXUD3LVrhKb7NHTj8k6ihh0c",
// 		},
// 	}
// 	c := new(JdFruit)
// 	go c.totalBean("GET")
// 	// fmt.Print(c.farmInfo)
// 	go c.jdFruit()
// 	wg.Add(2)
// 	wg.Wait()
// 	fmt.Print(time.Now().Unix() * 1000)
// }
