package infrastructure

import "github.com/agentmesh/orchestration-service/internal/domain"

// 内置 Hook 的具体实现。domain.Agent 只记录"启用了哪些 Hook 的名字"，
// 这些名字对应到具体函数的绑定关系（名字 -> 函数）由本文件维护，
// 并在 EinoRuntime.Run 里通过 Eino Callback 的 OnStart/OnEnd 挂载。
// 对应产品规格文档第三章「中间件 Hooks 层」。

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

func loadMemory(ctx domain.RunContext) error      { return nil /* TODO: 接入 Redis/向量库读取会话记忆 */ }
func checkPermission(ctx domain.RunContext) error { return nil /* TODO: 校验当前 Agent 对目标能力的调用权限 */ }
func buildPrompt(ctx domain.RunContext) error     { return nil /* TODO: 组装系统提示词与上下文 */ }

func reportUsageHook(ctx domain.RunContext) error { return nil /* TODO: 调用 AsyncUsageReporter */ }
func cleanupSandbox(ctx domain.RunContext) error  { return nil /* TODO: 释放本次调用的沙箱资源 */ }
func persistState(ctx domain.RunContext) error    { return nil /* TODO: Checkpoint 状态持久化 */ }
