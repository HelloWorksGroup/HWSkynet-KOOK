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

func (words *TodayWords) MakeKHLCard() string {
	weekday := []string{"日", "一", "二", "三", "四", "五", "六"}
	todayWeekday := time.Now().Weekday()
	card := make([]TWCard, 1)
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
		},
	}
	jsons, _ := json.Marshal(card)
	return string(jsons)
}
