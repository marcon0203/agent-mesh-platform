package domain

import "errors"

// CapabilityVersion 是值对象：每次发布都产生一个新版本记录，
// changelog 是必填项——这是产品规格文档 §5.4「版本更新提醒」得以实现的前提，
// 没有 changelog 就无法生成有意义的"有新版本可用"提示。
type CapabilityVersion struct {
	CapabilityID string
	Version      string
	Changelog    string
	IsLatest     bool
}

func NewCapabilityVersion(capabilityID, version, changelog string) (CapabilityVersion, error) {
	if changelog == "" {
		return CapabilityVersion{}, errors.New("changelog is required for every version")
	}
	return CapabilityVersion{CapabilityID: capabilityID, Version: version, Changelog: changelog, IsLatest: true}, nil
}
