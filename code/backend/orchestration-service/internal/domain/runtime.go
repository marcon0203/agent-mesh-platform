package domain

// AgentRuntime 是领域层对"执行引擎"的抽象端口。
// 领域层只关心「给定一个 Agent 聚合根和运行上下文，跑一次 Agentic Loop 并流式产出结果」，
// 不关心底层是 Eino ADK、还是未来替换成别的框架 —— 这正是把 Eino 隔离在
// infrastructure 层、而不是散落在 domain/application 里的意义所在。
type AgentRuntime interface {
	Run(agent *Agent, ctx RunContext, message string, out chan<- InvokeChunk) error
}
