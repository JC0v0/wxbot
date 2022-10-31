package covid19

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/yqchilde/wxbot/engine"
	"github.com/yqchilde/wxbot/engine/robot"
)

type Covid19 struct{ engine.PluginMagic }

var (
	pluginInfo = &Covid19{
		engine.PluginMagic{
			Desc:     "🚀 输入 {XX疫情} => 获取疫情数据，Ps:济南疫情",
			Commands: []string{`([^\x00-\xff]{0,6})疫情(.*)`},
			Weight:   99,
		},
	}
	plugin = engine.InstallPlugin(pluginInfo)
)

func (p *Covid19) OnRegister() {}

func (p *Covid19) OnEvent(msg *robot.Message) {
	if msg != nil {
		if len(msg.Content.Msg) > 3*6 {
			return
		}
		if idx, ok := msg.MatchRegexCommand(pluginInfo.Commands); ok {
			var re = regexp.MustCompile(pluginInfo.Commands[idx])
			match := re.FindAllStringSubmatch(msg.Content.Msg, -1)
			if len(match) > 0 && len(match[0]) > 1 {
				city := match[0][1]

				var str string
				var ret string
				if len(city) > 0 && city != "全国" {
					data, err := getCityCovid19Info(city)
					if err != nil {
						plugin.Errorf(err.Error())
						msg.ReplyText(fmt.Sprintf("获取%s疫情数据失败", city))
						return
					}
					str += "😦%s疫情今日数据统计如下: \n"
					str += "* %s\n"
					str += "* 新增本土: %s\n"
					str += "* 新增本土无症状: %s\n"
					str += "* 现有确诊: %s\n"
					str += "* 累计确诊: %s\n"
					str += "* 累计治愈: %s\n"
					str += "* 累计死亡: %s\n"
					ret = fmt.Sprintf(str, city, data.LastUpdateTime, data.LocalAdd, data.LocalAddWzz, data.ConfirmNow, data.ConfirmTotal, data.HealTotal, data.DeadTotal)
				} else {
					data, err := getDomesticCovid19Info()
					if err != nil {
						plugin.Errorf(err.Error())
						msg.ReplyText(fmt.Sprintf("获取%s疫情数据失败", city))
						return
					}
					str += "😦全国疫情今日数据统计如下: \n"
					str += "* 病例%s\n"
					str += "* 新增本土: %s\n"
					str += "* 现有本土: %s\n"
					str += "* 新增本土无症状: %s\n"
					str += "* 现有本土无症状: %s\n"
					str += "* 新增境外: %s\n"
					str += "* 现有境外: %s\n"
					str += "* 港澳台新增: %s\n"
					str += "* 现有确诊: %s\n"
					str += "* 累计确诊: %s(%s)\n"
					str += "* 累计境外: %s(%s)\n"
					str += "* 累计治愈: %s(%s)\n"
					str += "* 累计死亡: %s(%s)\n"
					ret = fmt.Sprintf(str, data.LastUpdateTime, data.LocalAdd, data.LocalNow, data.LocalAddWzz, data.LocalNowWzz, data.ForeignAdd, data.ForeignNow, data.HkMacTwAdd, data.ConfirmNow, data.ConfirmTotal, data.ConfirmTotalAdd, data.ForeignTotal, data.ForeignTotalAdd, data.HealTotal, data.HealTotalAdd, data.DeadTotal, data.DeadTotalAdd)
				}
				COVID19DaysCal := time.Now().Local().Sub(time.Date(2019, 12, 16, 0, 0, 0, 0, time.Local)).Hours() / 24
				COVID19Duration := fmt.Sprintf("😷自新冠疫情爆发以来已经过了%d天了，外出记得做好自我防护\n", int(COVID19DaysCal))
				msg.ReplyText(COVID19Duration + ret)
			}
		}
	}
}

func getDomesticCovid19Info() (*EpidemicData, error) {
	api := "https://opendata.baidu.com/data/inner?resource_id=5653&query=国内新型肺炎最新动态&dsp=iphone&tn=wisexmlnew&alr=1&is_opendata=1"
	resp, err := http.Get(api)
	if err != nil {
		plugin.Errorf("failed to get covid19 info api, err: %v", err)
		return nil, err
	}
	readAll, err := io.ReadAll(resp.Body)
	if err != nil {
		plugin.Errorf("failed to read resp body, err: %v", err)
		return nil, err
	}

	var data ApiResponse
	if err := json.Unmarshal(readAll, &data); err != nil {
		plugin.Errorf("failed to unmarshal data, err: %v", err)
		return nil, err
	}

	if len(data.Result) == 0 {
		return nil, errors.New("没有获取到数据")
	}

	tplData := data.Result[0].DisplayData.ResultData.TplData
	covid19Data := &EpidemicData{LastUpdateTime: tplData.Desc}
	for _, d := range tplData.DynamicList[0].DataList {
		switch d.TotalDesc {
		case "新增本土":
			covid19Data.LocalAdd = d.TotalNum
		case "现有本土":
			covid19Data.LocalNow = d.TotalNum
		case "新增本土无症状":
			covid19Data.LocalAddWzz = d.TotalNum
		case "现有本土无症状":
			covid19Data.LocalNowWzz = d.TotalNum
		case "新增境外":
			covid19Data.ForeignAdd = d.TotalNum
		case "现有境外":
			covid19Data.ForeignNow = d.TotalNum
		case "港澳台新增":
			covid19Data.HkMacTwAdd = d.TotalNum
		case "现有确诊":
			covid19Data.ConfirmNow = d.TotalNum
		case "累计确诊":
			covid19Data.ConfirmTotal = d.TotalNum
			covid19Data.ConfirmTotalAdd = d.ChangeNum
		case "累计境外":
			covid19Data.ForeignTotal = d.TotalNum
			covid19Data.ForeignTotalAdd = d.ChangeNum
		case "累计治愈":
			covid19Data.HealTotal = d.TotalNum
			covid19Data.HealTotalAdd = d.ChangeNum
		case "累计死亡":
			covid19Data.DeadTotal = d.TotalNum
			covid19Data.DeadTotalAdd = d.ChangeNum
		}
	}
	return covid19Data, nil
}

func getCityCovid19Info(city string) (*EpidemicData, error) {
	api := "https://opendata.baidu.com/data/inner?resource_id=5653&query=" + city + "新型肺炎最新动态&dsp=iphone&tn=wisexmlnew&alr=1&is_opendata=1"
	resp, err := http.Get(api)
	if err != nil {
		plugin.Errorf("failed to get covid19 info api, err: %v", err)
		return nil, err
	}
	readAll, err := io.ReadAll(resp.Body)
	if err != nil {
		plugin.Errorf("failed to read resp body, err: %v", err)
		return nil, err
	}

	var data ApiResponse
	if err := json.Unmarshal(readAll, &data); err != nil {
		plugin.Errorf("failed to unmarshal data, err: %v", err)
		return nil, err
	}

	if len(data.Result) == 0 {
		return nil, errors.New("没有获取到数据")
	}

	tplData := data.Result[0].DisplayData.ResultData.TplData
	covid19Data := &EpidemicData{LastUpdateTime: tplData.Desc}
	for _, d := range tplData.DataList {
		switch d.TotalDesc {
		case "新增本土":
			covid19Data.LocalAdd = d.TotalNum
		case "新增本土无症状":
			covid19Data.LocalAddWzz = d.TotalNum
		case "现有确诊":
			covid19Data.ConfirmNow = d.TotalNum
		case "累计确诊":
			covid19Data.ConfirmTotal = d.TotalNum
		case "累计治愈":
			covid19Data.HealTotal = d.TotalNum
		case "累计死亡":
			covid19Data.DeadTotal = d.TotalNum
		}
	}
	return covid19Data, nil
}
