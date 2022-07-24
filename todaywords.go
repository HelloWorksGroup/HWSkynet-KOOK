package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type OneQuote struct {
	Quote  string
	From   string
	Author string
}

type Newslist struct {
	Content string `json:"content"`
	Source  string `json:"source"`
	Author  string `json:"author"`
}
type TodayPoem struct {
	Code      string
	Msg       string
	Newslists []Newslist `json:"newslist"`
	Poem      string     `json:"content"`
	Author    string     `json:"author"`
	From      string     `json:"source"`
}

type TodayWords struct {
	IsTodaySaid bool
	Prob        int
	TodaySays   OneQuote
}

func (words *TodayWords) TrySay() OneQuote {
	words.TodaySays.Quote = ""
	if rand.Intn(100) >= words.Prob {
		// var obj Hitokoto
		// err := apiGet("https://v1.hitokoto.cn/?c=a", &obj)

		var obj TodayPoem
		err := apiGet(`http://api.tianapi.com/verse/index?key=b3caa56cb74c5e70594183505422a362`, &obj)
		if err != nil {
			fmt.Println("TodayWords Error:", err.Error())
			return words.TodaySays
		}
		words.TodaySays = OneQuote{
			Quote:  obj.Newslists[0].Content,
			From:   obj.Newslists[0].Source,
			Author: obj.Newslists[0].Author,
		}
		words.Prob = 101
		words.IsTodaySaid = true
		return words.TodaySays
	} else if words.Prob > 0 {
		words.Prob -= 1
	}
	return words.TodaySays
}

func (words *TodayWords) NewDay() {
	words.IsTodaySaid = false
	words.Prob = 100
}

type TWCField struct {
	Type    string `json:"type,omitempty"`
	Content string `json:"content,omitempty"`
	Src     string `json:"src,omitempty"`
}
type TWCModule struct {
	Type     string     `json:"type,omitempty"`
	Text     TWCField   `json:"text,omitempty"`
	Elements []TWCField `json:"elements,omitempty"`
}
type TWCard struct {
	Type    string      `json:"type"`
	Theme   string      `json:"theme"`
	Size    string      `json:"size"`
	Modules []TWCModule `json:"modules"`
}

var itemAtWeekday = map[string][]time.Weekday{
	"https://uploadstatic.mihoyo.com/ys-obc/2021/08/22/6381993/71f29b74fb8d894262a0e843bc9436d4_2507497043023918071.png":  {0, 1, 4},
	"https://uploadstatic.mihoyo.com/ys-obc/2020/04/13/4820086/15672798bec3a90c82c4dd7b00ce6640_4280874343063396873.png":  {0, 1, 4},
	"https://uploadstatic.mihoyo.com/ys-obc/2020/04/10/15568211/1145c4b5ac9012f7a30f0b7e480e5b95_5114100765295382355.png": {0, 1, 4},
	"https://uploadstatic.mihoyo.com/ys-obc/2020/04/10/15568211/26f915d93ca74fc9a520f3a1bdc5427f_350339505179470767.png":  {0, 1, 4},

	"https://uploadstatic.mihoyo.com/ys-obc/2021/08/22/6381993/451b6018f62f60061aa9f10bbad8871d_9006959577038210665.png":  {0, 2, 5},
	"https://uploadstatic.mihoyo.com/ys-obc/2021/08/22/6381993/e573023ff97a249e1be14dd202f7b198_1575676778480351549.png":  {0, 2, 5},
	"https://uploadstatic.mihoyo.com/ys-obc/2020/04/10/15568211/022b46bfd2a41d7729113adf96d19e7a_428450291747485.png":     {0, 2, 5},
	"https://uploadstatic.mihoyo.com/ys-obc/2020/04/10/15568211/a28104baf04e93246d2fe4fb0e0e90ab_8086194865950147291.png": {0, 2, 5},

	"https://uploadstatic.mihoyo.com/ys-obc/2021/08/22/6381993/6ad72ef6621e581601eca9074c253a70_2433491468200478226.png":  {0, 3, 6},
	"https://uploadstatic.mihoyo.com/ys-obc/2020/04/13/4820086/db0dfd84e7aa2dd9f2b331e19d6e2072_7104433016319434571.png":  {0, 3, 6},
	"https://uploadstatic.mihoyo.com/ys-obc/2020/04/10/15568211/7b726840ef48b8c43f702dafd02d3fd1_4580180554051984582.png": {0, 3, 6},
	"https://uploadstatic.mihoyo.com/ys-obc/2020/04/10/15568211/ad5f25e5cfa9531092ec9fabaa835852_5196014927639726004.png": {0, 3, 6},
}

func (words *TodayWords) MakeKHLCard() string {
	weekday := []string{"日", "一", "二", "三", "四", "五", "六"}
	todayWeekday := time.Now().Weekday()
	card := make([]TWCard, 1)
	todayItemsImageGroup := make([]TWCField, 0)
	for k, v := range itemAtWeekday {
		for _, day := range v {
			if day == todayWeekday {
				todayItemsImageGroup = append(todayItemsImageGroup, TWCField{
					Type: "image",
					Src:  k,
				})
			}
		}
	}
	card[0] = TWCard{
		Type:  "card",
		Theme: "primary",
		Size:  "lg",
		Modules: []TWCModule{
			{
				Type: "header",
				Text: TWCField{
					Type:    "plain-text",
					Content: `"` + words.TodaySays.Quote + `"`,
				},
			},
			{
				Type: "context",
				Elements: []TWCField{
					{
						Type:    "kmarkdown",
						Content: "「" + words.TodaySays.From + "」 " + words.TodaySays.Author,
					},
				},
			},
			{Type: "divider"},
			{
				Type: "section",
				Text: TWCField{
					Type:    "kmarkdown",
					Content: "现在是 `" + time.Now().Format("2006年01月02日15点04分") + " 星期" + weekday[todayWeekday] + "`",
				},
			},
			{
				Type: "section",
				Text: TWCField{
					Type:    "kmarkdown",
					Content: "**今天能刷下面这些材料哦:**\n---",
				},
			},
			{
				Type:     "image-group",
				Elements: todayItemsImageGroup,
			},
		},
	}
	jsons, _ := json.Marshal(card)
	return string(jsons)
}
