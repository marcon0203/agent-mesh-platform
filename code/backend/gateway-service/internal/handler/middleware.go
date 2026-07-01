package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// AuthMiddleware 校验 API Key（哈希比对）或 Workbench Session，
// 并将解析出的 scopes 挂到 context 供后续 handler 使用。
// 实现细节见技术规格文档 6.4 节。
func AuthMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// TODO: 解析 Authorization Header -> 哈希 -> 查 api_key 表 -> 校验 scope
		c.Next(ctx)
	}
}

// RateLimitMiddleware 基于 Redis + Lua 实现令牌桶限流，
// 维度覆盖 apikey / agent / tenant / token 四个层级。
// 实现细节见技术规格文档 6.3 节。
func RateLimitMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// TODO: 命中限流时返回 429 + X-RateLimit-* Headers，参考产品规格文档 6.2 节
		c.Next(ctx)
	}
}
