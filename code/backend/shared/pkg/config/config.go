// Package config 给 4 个后端服务提供统一的"YAML 配置文件 + 环境变量覆盖"
// 加载逻辑，避免每个服务各写一遍。约定：
//   1. 调用方先构造一个带本地开发默认值的 struct
//   2. Load 用 config.yaml（或 CONFIG_FILE 指定的路径）里存在的字段覆盖它，
//      文件不存在时直接跳过，保留默认值
//   3. 再用每个字段 `env:"XXX"` tag 对应的环境变量覆盖一次（非空才生效），
//      优先级：环境变量 > 配置文件 > 代码里的默认值
//
// 只处理 string 类型字段——四个服务的配置项目前都是 DSN/地址/密钥这类字符串，
// 没有需要覆盖数字或布尔字段的场景。
package config

import (
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Load 把 path 指向的 YAML 文件内容以及对应的环境变量覆盖合并进 out
// （out 必须是指向 struct 的指针）。
func Load(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read config file %s: %w", path, err)
		}
	} else if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}

	applyEnvOverrides(out)
	return nil
}

// Path 决定配置文件路径：CONFIG_FILE 环境变量优先，否则用调用方传入的默认值
// （通常是服务目录下的 "config.yaml"，相对于 go run/编译后二进制的工作目录）。
func Path(defaultPath string) string {
	if p := os.Getenv("CONFIG_FILE"); p != "" {
		return p
	}
	return defaultPath
}

func applyEnvOverrides(out interface{}) {
	v := reflect.ValueOf(out).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envKey := field.Tag.Get("env")
		if envKey == "" {
			continue
		}
		if val, ok := os.LookupEnv(envKey); ok && val != "" {
			v.Field(i).SetString(val)
		}
	}
}
