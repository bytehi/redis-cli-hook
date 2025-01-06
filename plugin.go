package main

import (
	"log"
	"os"

	"github.com/bytehi/resp"
)

// Plugin 定义了钩子插件的接口
type Plugin interface {
	BeforeCommand(cmd string, args []string) (string, []string, error)
	AfterCommand(cmd string, args []string, ack *resp.RESP) (*resp.RESP, error)
}

// BasePlugin 提供基础实现
type BasePlugin struct {
	enabled bool
}

// 实现一个具体的插件示例：日志插件
type LogPlugin struct {
	BasePlugin
	logger *log.Logger
}

func NewLogPlugin(enabled bool) *LogPlugin {
	return &LogPlugin{
		BasePlugin: BasePlugin{enabled: enabled},
		logger:     log.New(os.Stdout, "[redis-cli] ", log.LstdFlags),
	}
}

func (p *LogPlugin) BeforeCommand(cmd string, args []string) (string, []string, error) {
	if !p.enabled {
		return cmd, args, nil
	}
	p.logger.Printf("执行命令: %s, 参数: %v", cmd, args)
	return cmd, args, nil
}

func (p *LogPlugin) AfterCommand(cmd string, args []string, ack *resp.RESP) (*resp.RESP, error) {
	if !p.enabled {
		return ack, nil
	}
	p.logger.Printf("命令结果: %v", ack)
	return ack, nil
}

var Plugins = []Plugin{
	NewLogPlugin(true),
}
