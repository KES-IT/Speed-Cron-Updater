package main

import (
	"context"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"os/exec"
	"speed-cron-updater/utils"
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
		githubVersion, _, downloadStatus := utils.GetLatestVersionInfo()
		if githubVersion == "" || !downloadStatus {
			glog.Warning(ctx, "获取github最新版本失败，无法比较版本")
			time.Sleep(5 * time.Second)
			return
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
			time.Sleep(5 * time.Second)
			return
		}
		glog.Debug(ctx, "关闭speed_cron结果：", string(stopOut))

		// 下载最新版本
		_, exeUrl, downloadStatus := utils.GetLatestVersionInfo()
		if !downloadStatus {
			glog.Warning(ctx, "下载最新版本: 获取github最新版本失败，无法下载")
			time.Sleep(5 * time.Second)
			return
		}
		exeUrl = "https://gh.xinyu.today/" + exeUrl
		if err := utils.HTTPDownloadFileWithPercent(exeUrl, "client/core_bin/speed_cron.exe"); err != nil {
			glog.Warning(ctx, "下载speed_cron失败，原因：", err.Error())
			time.Sleep(5 * time.Second)
			return
		}
		glog.Info(ctx, "下载speed_cron成功...")

		// 检测版本
		versionCmd := exec.Command("client/core_bin/speed_cron.exe", "version")
		versionOut, err := versionCmd.Output()
		if err != nil {
			glog.Warning(ctx, "检测speed_cron版本失败，原因：", err.Error())
			time.Sleep(5 * time.Second)
			return
		}
		glog.Info(ctx, "最新获取到本地speed_cron版本为：", string(versionOut))

		startCmd := exec.Command("client/speed_cron_process.exe", "start")
		startOut, err := startCmd.Output()
		if err != nil {
			glog.Warning(ctx, "启动speed_cron失败，原因：", err.Error())
			time.Sleep(5 * time.Second)
			return
		}
		glog.Debug(ctx, "启动speed_cron结果：", string(startOut))
		glog.Info(ctx, "更新完成...程序会在5s后自动关闭...")
		time.Sleep(5 * time.Second)
	}

}
