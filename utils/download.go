package utils

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// HTTPDownloadFileWithPercent downloads target url file to local path with percent process printing.
func HTTPDownloadFileWithPercent(url string, localSaveFilePath string) error {
	start := time.Now()
	out, err := os.Create(localSaveFilePath)
	if err != nil {
		return gerror.Wrapf(err, `download "%s" to "%s" failed`, url, localSaveFilePath)
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			glog.Warning(context.Background(), err)
		}
	}(out)

	headResp, err := http.Head(url)
	if err != nil {
		return gerror.Wrapf(err, `download "%s" to "%s" failed`, url, localSaveFilePath)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			glog.Warning(context.Background(), err)
		}
	}(headResp.Body)

	size, err := strconv.Atoi(headResp.Header.Get("Content-Length"))
	if err != nil {
		return gerror.Wrap(err, "retrieve Content-Length failed")
	}
	doneCh := make(chan int64)

	go doPrintDownloadPercent(doneCh, localSaveFilePath, int64(size))

	resp, err := http.Get(url)
	if err != nil {
		return gerror.Wrapf(err, `download "%s" to "%s" failed`, url, localSaveFilePath)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			glog.Warning(context.Background(), err)
		}
	}(resp.Body)

	wroteBytesCount, err := io.Copy(out, resp.Body)
	if err != nil {
		return gerror.Wrapf(err, `download "%s" to "%s" failed`, url, localSaveFilePath)
	}

	doneCh <- wroteBytesCount
	elapsed := time.Since(start)
	if elapsed > time.Minute {
		glog.Printf(context.Background(), `download completed in %.0fm`, float64(elapsed)/float64(time.Minute))
	} else {
		glog.Printf(context.Background(), `download completed in %.0fs`, elapsed.Seconds())
	}

	return nil
}

func doPrintDownloadPercent(doneCh chan int64, localSaveFilePath string, total int64) {
	var (
		stop           = false
		lastPercentFmt string
	)
	for {
		select {
		case <-doneCh:
			stop = true

		default:
			file, err := os.Open(localSaveFilePath)
			if err != nil {
				glog.Fatal(context.Background(), err)
				time.Sleep(5 * time.Second)
			}
			fi, err := file.Stat()
			if err != nil {
				glog.Fatal(context.Background(), err)
				time.Sleep(5 * time.Second)
			}
			size := fi.Size()
			if size == 0 {
				size = 1
			}
			var (
				percent    = float64(size) / float64(total) * 100
				percentFmt = fmt.Sprintf(`%.0f`, percent) + "%"
			)
			if lastPercentFmt != percentFmt {
				lastPercentFmt = percentFmt
				glog.Debug(context.Background(), "下载最新二进制文件进度: "+percentFmt)
			}
		}

		if stop {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}
