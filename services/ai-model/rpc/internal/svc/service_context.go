package svc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"community-and-home/services/ai-model/rpc/internal/config"
)

type ServiceContext struct {
	Config       config.Config
	PythonClient *http.Client
	PythonURL    string
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 创建 HTTP 客户端用于调用 Python 引擎
	client := &http.Client{
		Timeout: time.Duration(c.PythonEngine.Timeout) * time.Millisecond,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	pythonURL := fmt.Sprintf("http://%s:%d", c.PythonEngine.Host, c.PythonEngine.Port)

	logx.Infof("Python Engine URL: %s", pythonURL)

	return &ServiceContext{
		Config:       c,
		PythonClient: client,
		PythonURL:    pythonURL,
	}
}
