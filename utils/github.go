package utils

import (
	"context"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/glog"
)

// GetLatestVersionInfo 获取github最新版本
func GetLatestVersionInfo() (version string, downloadUrl string, downloadStatus bool) {
	backendURL := "http://120.24.211.49:10441/GetLatestVersion"
	response, err := g.Client().Get(context.TODO(), backendURL)
	if err != nil {
		glog.Warning(context.TODO(), "请求github最新版本失败，原因：", err.Error())
		return "", "", false
	}
	defer func(response *gclient.Response) {
		err := response.Close()
		if err != nil {
			glog.Warning(context.TODO(), "关闭response失败，原因：", err.Error())
		}
	}(response)
	githubResJson, err := gjson.DecodeToJson(response.ReadAllString())
	if err != nil {
		glog.Warning(context.TODO(), "解析response失败，原因：", err.Error())
		return "", "", false
	}

	// 判断GitHub Release可更新二进制文件是否存在
	if len(githubResJson.Get("data.github_res.assets").Array()) == 0 {
		glog.Warning(context.TODO(), "解析response失败，原因：", "github_res.assets为空")
		return "", "", false
	}
	version = githubResJson.Get("data.github_res.tag_name").String()

	// 获取下载文件名是否正确
	downloadFileName := githubResJson.Get("data.github_res.assets.0.name").String()
	if downloadFileName != "speed_cron_windows_amd64.exe" {
		glog.Warning(context.TODO(), "解析response失败，原因：", "downloadFileName不正确")
		return "", "", false
	}

	// 获取下载地址
	downloadUrl = githubResJson.Get("data.github_res.assets.0.browser_download_url").String()
	if version == "" || downloadUrl == "" {
		glog.Warning(context.TODO(), "解析response失败，原因：", "version或downloadUrl为空")
		return "", "", false
	}
	return version, downloadUrl, true
}
