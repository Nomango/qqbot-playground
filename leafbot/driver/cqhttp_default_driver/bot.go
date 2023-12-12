package cqhttp_default_driver

import (
	"encoding/json"
	"errors"
	"github.com/Mrs4s/go-cqhttp/pkg/onebot"
	"sync"
	"time"

	"github.com/Mrs4s/go-cqhttp/coolq"
	"github.com/Mrs4s/go-cqhttp/modules/api"
	"github.com/tidwall/gjson"
)

type userAPI struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
	Echo   string      `json:"echo"`
}

func (u userAPI) Get(s string) gjson.Result {
	data, _ := json.Marshal(u.Params)
	parse := gjson.Parse(string(data))
	return parse.Get(s)
}

// Bot
// @Description: Bot对象
type Bot struct {
	responses sync.Map
	CQBot     *coolq.CQBot

	call *api.Caller
}

func (b *Bot) Do(i interface{}) {
	data := i.(userAPI)
	call := b.call.Call(data.Action, onebot.V11, data)
	resp, _ := json.Marshal(call)
	b.responses.Store(data.Echo, resp)
}

func (b *Bot) GetResponse(echo string) ([]byte, error) {
	defer func() {
		b.responses.Delete(echo)
	}()

	for i := 0; i < 120; i++ {
		value, ok := b.responses.LoadAndDelete(echo)
		if ok {
			return value.([]byte), nil
		}
		time.Sleep(500)
	}

	return nil, errors.New("get response time out")
}

func (b *Bot) GetSelfId() int64 {
	return b.CQBot.Client.Uin
}
