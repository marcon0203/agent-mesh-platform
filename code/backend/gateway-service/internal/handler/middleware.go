package handler

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/go-redis/v9"
)

// apiKeyContextKey 是鉴权通过后，把 API Key 信息挂到 RequestContext.Keys 的键名，
// 供限流中间件区分维度、供 handler 做审计用。
const apiKeyContextKey = "api_key"

type apiKeyInfo struct {
	ID          string
	DeveloperID string
	Scopes      []string
}

// AuthService 校验 API Key（哈希比对 + scope 校验）。
// 实现细节见产品规格文档 §6.1、技术规格文档 §6.4。
//
// Workbench 通道按设计应该走独立的 Session 鉴权分支（不复用 API Key 校验），
// 但当前平台还没有账号登录/Session 体系，这属于本次 M1 范围之外的能力，
// Workbench 页面目前也需要提供一个 API Key 才能调用 sendMsg/streamMsg。
type AuthService struct {
	DB *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{DB: db}
}

func (s *AuthService) Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		authHeader := string(c.GetHeader("Authorization"))
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			abortWithCode(c, 401, 40101, "missing or malformed Authorization header")
			return
		}
		rawKey := strings.TrimPrefix(authHeader, prefix)
		sum := sha256.Sum256([]byte(rawKey))
		keyHash := hex.EncodeToString(sum[:])

		var (
			id, developerID string
			scopesJSON      string
			status          int
		)
		row := s.DB.QueryRow(`SELECT id, developer_id, scopes, status FROM api_key WHERE key_hash = ?`, keyHash)
		if err := row.Scan(&id, &developerID, &scopesJSON, &status); err != nil {
			abortWithCode(c, 401, 40101, "api key invalid or revoked")
			return
		}
		if status != 1 {
			abortWithCode(c, 401, 40101, "api key invalid or revoked")
			return
		}

		var scopes []string
		_ = json.Unmarshal([]byte(scopesJSON), &scopes)

		agentID := c.Param("agent_id")
		if !hasScope(scopes, "agent:"+agentID+":invoke") {
			abortWithCode(c, 401, 40102, "api key has no invoke scope for this agent")
			return
		}

		c.Set(apiKeyContextKey, apiKeyInfo{ID: id, DeveloperID: developerID, Scopes: scopes})
		c.Next(ctx)
	}
}

func hasScope(scopes []string, required string) bool {
	for _, s := range scopes {
		if s == required {
			return true
		}
	}
	return false
}

func abortWithCode(c *app.RequestContext, httpStatus, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, map[string]any{"code": code, "message": message})
}

// RateLimitService 基于 Redis + Lua 实现令牌桶限流，覆盖 apikey / agent 两个维度
// （tenant / token 消耗级维度依赖账号体系和逐 token 计量，留待后续里程碑接入，
// 见技术规格文档 §6.3）。
type RateLimitService struct {
	Redis *redis.Client
}

func NewRateLimitService(rdb *redis.Client) *RateLimitService {
	return &RateLimitService{Redis: rdb}
}

const (
	apiKeyRateLimitPerMinute = 60  // 产品规格文档 §6.2：API Key 级默认 60 次/分钟
	agentRateLimitPerMinute  = 120 // Agent 级默认上限，示意值，后续可由发布方自主设置
	rateLimitWindowSeconds   = 60
)

// tokenBucketScript 是令牌桶算法的原子实现：
//   KEYS[1] = bucket key
//   ARGV[1] = capacity        令牌桶容量（即限流阈值）
//   ARGV[2] = refill_per_sec  每秒补充速率
//   ARGV[3] = now             当前 unix 秒
//   ARGV[4] = window          限流窗口（秒），用于计算 reset 时间和 key 过期
// 返回 {allowed(0/1), remaining, reset_at}
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local window = tonumber(ARGV[4])

local bucket = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(bucket[1])
local ts = tonumber(bucket[2])

if tokens == nil then
	tokens = capacity
	ts = now
end

local delta = math.max(0, now - ts)
tokens = math.min(capacity, tokens + delta * refill_rate)

local allowed = 0
if tokens >= 1 then
	allowed = 1
	tokens = tokens - 1
end

redis.call("HMSET", key, "tokens", tokens, "ts", now)
redis.call("EXPIRE", key, window * 2)

return {allowed, math.floor(tokens), now + window}
`)

func (s *RateLimitService) Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		agentID := c.Param("agent_id")
		apiKeyID := "anonymous"
		if v, ok := c.Get(apiKeyContextKey); ok {
			if info, ok := v.(apiKeyInfo); ok {
				apiKeyID = info.ID
			}
		}

		if !s.checkDimension(ctx, c, "apikey:"+apiKeyID, apiKeyRateLimitPerMinute) {
			return
		}
		if !s.checkDimension(ctx, c, "agent:"+agentID, agentRateLimitPerMinute) {
			return
		}
		c.Next(ctx)
	}
}

// checkDimension 返回 false 时已经写完 429 响应，调用方需要立即 return。
func (s *RateLimitService) checkDimension(ctx context.Context, c *app.RequestContext, dimensionKey string, limit int) bool {
	now := time.Now().Unix()
	refillRate := float64(limit) / float64(rateLimitWindowSeconds)

	result, err := tokenBucketScript.Run(ctx, s.Redis, []string{"ratelimit:" + dimensionKey}, limit, refillRate, now, rateLimitWindowSeconds).Result()
	if err != nil {
		// Redis 故障时降级放行，避免限流组件本身成为单点故障。
		return true
	}
	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		return true
	}
	allowed, _ := values[0].(int64)
	remaining, _ := values[1].(int64)
	resetAt, _ := values[2].(int64)

	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

	if allowed == 0 {
		retryAfter := resetAt - now
		if retryAfter < 1 {
			retryAfter = 1
		}
		c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
		c.AbortWithStatusJSON(429, map[string]any{
			"code":    40002,
			"message": fmt.Sprintf("超出调用配额，请 %d 秒后重试", retryAfter),
		})
		return false
	}
	return true
}
