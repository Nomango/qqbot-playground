package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Nomango/ark/slices"
	"github.com/huoxue1/leafbot/chat"
	"github.com/huoxue1/leafbot/driver/cqhttp_positive_ws_driver"
	"github.com/huoxue1/leafbot/message"
	"github.com/sirupsen/logrus"

	"github.com/huoxue1/leafbot"

	_ "github.com/Mrs4s/go-cqhttp/db/leveldb" // leveldb

	// _ "github.com/Mrs4s/go-cqhttp/modules/pprof" // pprof 性能分析

	_ "github.com/Mrs4s/go-cqhttp/modules/silk" // si
)

func init() {
	plugin := leafbot.NewPlugin("小碎花")
	// plugin.OnCommand("测试", leafbot.Option{
	// 	Weight: 0,
	// 	Block:  false,
	// 	Allies: nil,
	// 	Rules: []leafbot.Rule{func(ctx *leafbot.Context) bool {
	// 		return true
	// 	}},
	// }).Handle(func(ctx *leafbot.Context) {
	// 	ctx.Send(message.Text("123"))
	// })
	// plugin.OnStart("开头").Handle(func(ctx *leafbot.Context) {
	// 	ctx.Send(message.Text("onStart匹配成功"))
	// })
	// plugin.OnEnd("结束").Handle(func(ctx *leafbot.Context) {
	// 	ctx.Send("onEnd匹配成功")
	// })
	// plugin.OnRegex(`我的(.*?)时小明`).Handle(func(ctx *leafbot.Context) {
	// 	log.Infoln(ctx.State.RegexResult)
	// 	ctx.Send(message.Text("正则匹配成功"))
	// })
	plugin.OnCommand("重置", leafbot.Option{
		Weight: 0,
		Block:  false,
		Allies: nil,
		Rules: []leafbot.Rule{func(ctx *leafbot.Context) bool {
			return true
		}},
	}).Handle(func(ctx *leafbot.Context) {
		chat.ResetThread(fmt.Sprint(ctx.GroupID))
		logrus.Debugln("重置group:", ctx)
	})
	plugin.OnMessage("", leafbot.Option{Rules: []leafbot.Rule{func(ctx *leafbot.Context) bool {
		return strings.Contains(ctx.Event.RawMessage, "[CQ:at,qq=2591190108]")
		// has := slices.Filter(ctx.Event.Message, func(m message.MessageSegment) bool { return m.Type == "at" && m.Data["qq"] == "2591190108" })
		// return len(has) > 0
	}}}).Handle(func(ctx *leafbot.Context) {
		msgs := slices.Filter(ctx.Event.Message, func(m message.MessageSegment) bool { return m.Type == "text" })
		input := strings.Join(slices.Map(msgs, func(m message.MessageSegment) string { return m.Data["text"] }), "\n")
		c, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		result, err := chat.Chat(c, fmt.Sprint(ctx.GroupID), input)
		if err != nil {
			logrus.Errorln(err)
		}
		ctx.Send(message.Text(strings.Join(result, "\n")))
	})
}

func main() {
	// 创建一个驱动
	// driver := cqhttp_default_driver.NewDriver()
	driver := cqhttp_positive_ws_driver.NewDriver("ws://127.0.0.1:2333", "")
	// 注册驱动
	leafbot.LoadDriver(driver)
	// 初始化Bot
	leafbot.InitBots(leafbot.Config{
		NickName:     []string{"leafBot"},
		Admin:        0,
		SuperUser:    nil,
		CommandStart: []string{"/"},
		LogLevel:     "",
	})
	// 运行驱动
	driver.Run()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
