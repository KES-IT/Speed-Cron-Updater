package main

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/schollz/progressbar/v3"
	"io"
	"os"
	"os/exec"
	"time"
)

func main() {
	var (
		ctx       = context.TODO()
		latestTag = false
	)
	glog.Info(ctx, "开始更新检测...")
	glog.Debug(ctx, "检测speed_cron文件是否存在...")
	if gfile.Exists("client/core_bin/speed_cron.exe") {
		glog.Debug(ctx, "speed_cron文件存在，开始检查版本...")
		// 检测版本
		versionCmd := exec.Command("client/core_bin/speed_cron.exe", "version")
		versionOut, err := versionCmd.Output()
		if err != nil {
			glog.Warning(ctx, "检测speed_cron版本失败，原因：", err.Error())
		}
		if string(versionOut) == "" {
			glog.Info(ctx, "speed_cron版本为空，开始下载...")
		} else {
			glog.Info(ctx, "speed_cron版本为：", string(versionOut))
		}
		// 与服务器版本比较
		githubVersion := getLatestVersion()
		if githubVersion == "" {
			glog.Warning(ctx, "获取github最新版本失败，无法比较版本，将自动下载最新版本")
		} else {
			glog.Info(ctx, "目前最新githubVersion为: ", githubVersion)
			if githubVersion != string(versionOut) {
				glog.Info(ctx, "speed_cron版本不是最新，开始下载...")
			} else {
				glog.Info(ctx, "speed_cron版本是最新，无需下载...")
				latestTag = true
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		glog.Debug(ctx, "speed_cron文件不存在，开始下载...")
	}
	if !latestTag {
		// 关闭原有服务
		stopCmd := exec.Command("client/speed_cron_process.exe", "stop")
		stopOut, err := stopCmd.Output()
		if err != nil {
			glog.Warning(ctx, "关闭speed_cron失败，原因：", err.Error())
		}
		glog.Debug(ctx, "关闭speed_cron结果：", string(stopOut))

		// 下载最新版本
		exeUrl := "https://gh.xinyu.today/https://github.com/hamster1963/Speed-Cron/releases/latest/download/speed_cron_windows_amd64.exe"
		exe, err := g.Client().Get(ctx, exeUrl)
		bar := progressbar.NewOptions(int(exe.ContentLength),
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionEnableColorCodes(false),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionSetWidth(14),
			progressbar.OptionOnCompletion(func() {
				_, _ = fmt.Fprint(os.Stderr, "\n")
			}),
			progressbar.OptionSetDescription("下载最新测速客户端中..."))

		f, _ := os.OpenFile("client/core_bin/speed_cron.exe", os.O_CREATE|os.O_WRONLY, 0644)
		_, err = io.Copy(io.MultiWriter(f, bar), exe.Body)
		if err != nil {
			glog.Warning(ctx, "下载speed_cron失败，原因：", err.Error())
			return
		}
		_ = f.Close()

		glog.Info(ctx, "下载speed_cron成功...")

		versionCmd := exec.Command("client/core_bin/speed_cron.exe", "version")
		versionOut, err := versionCmd.Output()
		if err != nil {
			glog.Warning(ctx, "检测speed_cron版本失败，原因：", err.Error())
		}
		glog.Info(ctx, "当前speed_cron版本为：", string(versionOut))

		startCmd := exec.Command("client/speed_cron_process.exe", "start")
		startOut, err := startCmd.Output()
		if err != nil {
			glog.Warning(ctx, "启动speed_cron失败，原因：", err.Error())
		}
		glog.Debug(ctx, "启动speed_cron结果：", string(startOut))
		glog.Info(ctx, "更新完成...程序会在5s后自动关闭...")
		time.Sleep(5 * time.Second)
	}

}

func getLatestVersion() (version string) {
	url := "http://120.24.211.49:10441/GetLatestVersion"
	response, err := g.Client().Get(context.TODO(), url)
	if err != nil {
		glog.Warning(context.TODO(), "请求github最新版本失败，原因：", err.Error())
		return ""
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
		return ""
	}
	return githubResJson.Get("data.github_res.assets.tag_name").String()
}
