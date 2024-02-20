package jieba

import (
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/labstack/gommon/log"
	"github.com/mylukin/EchoPilot/helper"
	"github.com/mylukin/gojieba"
	"golang.org/x/text/width"
)

var jbClient *jiebaClient
var reqClient *req.Client
var jbOnece sync.Once
var stopRemoteDict bool

type jiebaClient struct {
	client               *gojieba.Jieba
	dictDir              string
	remoteDict           string
	remoteDictLatestETag string
	isDownloading        bool
	req                  *req.Client
}

// StopRemoteDict
func StopRemoteDict() {
	stopRemoteDict = true
}

func New() *jiebaClient {
	jbOnece.Do(func() {
		reqClient = req.C()
		reqClient.SetTimeout(5 * time.Second)

		jiebaDictDir := strings.Trim(helper.Config("JIEBA_DICT_DIR"), `"`)
		jiebaRemoteDict := strings.Trim(helper.Config("JIEBA_REMOTE_DICT"), `"`)
		jbClient = &jiebaClient{
			client:        gojieba.NewJieba(jiebaDictDir),
			dictDir:       jiebaDictDir,
			remoteDict:    jiebaRemoteDict,
			req:           reqClient,
			isDownloading: false,
		}

		// 每60s检查一次，如果远程词库有变化，则下载
		if !stopRemoteDict && jiebaRemoteDict != "" {
			go func() {
				ticker := time.NewTicker(60 * time.Second)
				for ; true; <-ticker.C {
					jbClient.CheckRemoteDict()
				}
			}()
		}
	})

	return jbClient
}

// CheckRemoteDict
func (o *jiebaClient) CheckRemoteDict() {
	if o.remoteDict == "" || o.isDownloading {
		return
	}

	// 下载完成
	defer func() {
		o.isDownloading = false
	}()

	// 获取head
	resp, err := reqClient.R().Head(o.remoteDict)
	if err != nil {
		log.Errorf("%s: %s", err, o.remoteDict)
		return
	}

	getETag := resp.GetHeader("ETag")
	// 需要下载词库
	if o.remoteDictLatestETag != getETag {
		log.Debugf("old: %s, new: %s", o.remoteDictLatestETag, getETag)
		o.isDownloading = true
		o.remoteDictLatestETag = getETag
		// 进度条设置
		progress := func(info req.DownloadInfo) {
			log.Infof("%s: %.2f %%", getETag, float32(info.DownloadedSize)/float32(info.Response.ContentLength)*100)
		}
		// 下载词库
		_, err := reqClient.R().
			SetOutputFile(gojieba.USER_DICT_PATH).
			SetDownloadCallback(progress).
			Get(o.remoteDict)
		if err == nil {
			log.Infof("download complete: %s", o.remoteDict)
			// 重载词库
			o.Reload()
		} else {
			log.Errorf("%s: %s", err, o.remoteDict)
		}
	}
}

// Reload
func (o *jiebaClient) Reload() {
	if o.client != nil {
		o.client.Free()
	}
	o.client = gojieba.NewJieba(o.dictDir)
}

// Extract is Extract keywords
func (o *jiebaClient) Extract(text string, topk int) []string {
	// 全角转半角
	text = width.Narrow.String(text)
	return o.client.Extract(text, topk)
}

// CutAll
func (o *jiebaClient) CutAll(text string) []string {
	// 全角转半角
	text = width.Narrow.String(text)
	return o.client.CutAll(text)
}

// Cut
func (o *jiebaClient) Cut(text string, hmm bool) []string {
	// 全角转半角
	text = width.Narrow.String(text)
	return o.client.Cut(text, hmm)
}

// Tag
func (o *jiebaClient) Tag(text string) []string {
	// 全角转半角
	text = width.Narrow.String(text)
	return o.client.Tag(text)
}

// CutForSearch
func (o *jiebaClient) CutForSearch(text string, hmm bool) []string {
	// 全角转半角
	text = width.Narrow.String(text)
	return o.client.CutForSearch(text, hmm)
}
