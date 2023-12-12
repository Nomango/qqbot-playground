package cqhttp_positive_ws_driver

import (
	"encoding/json" //nolint:gci

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func (b *Bot) SendGroupMsg(groupID int64, message interface{}) int32 {
	return int32(b.CallApi("send_group_msg", map[string]interface{}{"group_id": groupID, "message": message}).Int())
}

func (b *Bot) SendPrivateMsg(userID int64, message interface{}) int32 {
	return int32(b.CallApi("send_private_msg", map[string]interface{}{"user_id": userID, "message": message}).Int())
}

func (b *Bot) CallApi(action string, param interface{}) gjson.Result {
	echo := uuid.NewV4().String()
	type userAPi struct {
		Action string      `json:"action"`
		Params interface{} `json:"params"`
		Echo   string      `json:"echo"`
	}
	var d = userAPi{
		Action: action,
		Params: param,
		Echo:   echo,
	}
	b.Do(d)
	data, err := b.GetResponse(echo)
	content, _ := json.Marshal(d)
	log.Infoln(string(content) + "\n\t\t\t\t\t" + string(data))
	if err != nil {
		// log.Errorln("调用API出现了错误")
		// log.Errorln(err)
		// log.Errorln(string(data))
		return gjson.Result{Type: gjson.Null}
	}
	if gjson.GetBytes(data, "status").String() != "ok" {
		log.Errorln("调用API出现了错误")
		log.Errorln(string(data))
		// log.Errorln(gjson.GetBytes(data, "msg"), ",", gjson.GetBytes(data, "wording"))
	}
	return gjson.GetBytes(data, "data")
}
