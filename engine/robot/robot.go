package robot

import (
	"github.com/imroc/req/v3"
	"github.com/yqchilde/pkgs/log"
)

var MyRobot BotConf

type BotConf struct {
	Server string `yaml:"server"`
	Token  string `yaml:"token"`
	Bot    Bot    `yaml:"-"`
}

func (b *BotConf) GetRobotInfo() error {
	payload := map[string]interface{}{
		"api":   "GetRobotList",
		"token": MyRobot.Token,
	}

	var resp BotList
	err := req.C().SetBaseURL(MyRobot.Server).Post().SetBody(payload).Do().Into(&resp)
	if err != nil {
		log.Errorf("get robot info error: %v", err)
		return err
	}
	if resp.Code != 0 {
		log.Errorf("get robot info error: %s", resp.Result)
		return err
	}
	MyRobot.Bot = resp.ReturnJson.Data[0]
	return nil
}

func (b *BotConf) GetGroupList() ([]Group, error) {
	payload := map[string]interface{}{
		"api":        "GetGrouplist",
		"token":      MyRobot.Token,
		"robot_wxid": MyRobot.Bot.Wxid,
		"is_refresh": "0",
	}

	var resp GroupList
	err := req.C().SetBaseURL(MyRobot.Server).Post().SetBody(payload).Do().Into(&resp)
	if err != nil {
		log.Errorf("get robot info error: %v", err)
		return nil, err
	}
	log.Printf("info: %#+v", resp)
	if resp.Code != 0 {
		log.Errorf("get robot info error: %s", resp.Result)
		return nil, err
	}
	return resp.ReturnJson, nil
}

// SendText 发送文本消息； to_wxid:好友ID/群ID
func (b *BotConf) SendText(toWxID string, msg string) error {
	payload := map[string]interface{}{
		"api":        "SendTextMsg",
		"token":      MyRobot.Token,
		"msg":        formatTextMessage(msg),
		"robot_wxid": MyRobot.Bot.Wxid,
		"to_wxid":    toWxID,
	}

	var resp MessageResp
	err := req.C().SetBaseURL(MyRobot.Server).Post().SetBody(payload).Do().Into(&resp)
	if err != nil {
		log.Errorf("reply text message error: %v", err)
		return err
	}
	if resp.Code != 0 {
		log.Errorf("reply text message error: %s", resp.Result)
		return err
	}
	return nil
}