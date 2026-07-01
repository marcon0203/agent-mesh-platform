package infrastructure

import (
	"log"

	"github.com/agentmesh/orchestration-service/internal/domain"
)

// 内置 Hook 的具体实现。domain.Agent 只记录"启用了哪些 Hook 的名字"，
// 这些名字对应到具体函数的绑定关系（名字 -> 函数）由本文件维护，
// 并在 EinoRuntime.Run 里通过 Eino Callback 的 OnStart/OnEnd 挂载。
// 对应产品规格文档第三章「中间件 Hooks 层」。
//
// MVP 阶段这里的实现以可观测的日志埋点为主：Hook 的"挂载机制"（Eino Callback
// OnStart/OnEnd 触发、按名字查表调用）已经完整工作；记忆存储、权限系统、沙箱等
// 深层基础设施还未建设，属于后续里程碑范围，届时只需替换这里的函数体，
// 调用方（EinoRuntime）不需要任何改动。

type HookFunc func(ctx domain.RunContext) error

var PreHookRegistry = map[string]HookFunc{
	"记忆加载":  loadMemory,
	"权限校验":  checkPermission,
	"提示词构建": buildPrompt,
}

var PostHookRegistry = map[string]HookFunc{
	"消费上报": reportUsageHook,
	"沙箱清理": cleanupSandbox,
	"状态回传": persistState,
}

func loadMemory(ctx domain.RunContext) error {
	// TODO: 接入 Redis/向量库读取会话记忆，当前只记录触发事件
	log.Printf("[hook:记忆加载] session=%s trace=%s depth=%d", ctx.SessionID, ctx.TraceID, ctx.Depth)
	return nil
}

func checkPermission(ctx domain.RunContext) error {
	// TODO: 校验当前 Agent 对目标能力的调用权限，当前默认放行
	log.Printf("[hook:权限校验] agent=%s trace=%s allowed=true", ctx.AgentID, ctx.TraceID)
	return nil
}

func buildPrompt(ctx domain.RunContext) error {
	// TODO: 组装系统提示词与上下文；Instruction 目前直接在 EinoRuntime.Run 里静态拼装
	log.Printf("[hook:提示词构建] agent=%s trace=%s", ctx.AgentID, ctx.TraceID)
	return nil
}

func reportUsageHook(ctx domain.RunContext) error {
	// TODO: 调用 AsyncUsageReporter 上报用量；InvokeUseCase 目前统一在调用结束后上报，
	// 这里先记录一次 Hook 触发，避免重复上报
	log.Printf("[hook:消费上报] agent=%s trace=%s depth=%d", ctx.AgentID, ctx.TraceID, ctx.Depth)
	return nil
}

func cleanupSandbox(ctx domain.RunContext) error {
	// TODO: 释放本次调用的沙箱资源；沙箱运行时尚未接入
	log.Printf("[hook:沙箱清理] session=%s trace=%s", ctx.SessionID, ctx.TraceID)
	return nil
}

func persistState(ctx domain.RunContext) error {
	// TODO: Checkpoint 状态持久化；可用 adk.RunnerConfig.CheckPointStore 接入
	log.Printf("[hook:状态回传] session=%s trace=%s", ctx.SessionID, ctx.TraceID)
	return nil
}
