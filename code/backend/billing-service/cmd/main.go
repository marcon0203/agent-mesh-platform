// billing-service 异步消费用量事件，按 Agent / 能力维度聚合计费与分成。
// 用量事件由 orchestration-service 在每次运行结束后写入消息队列，
// 避免同步计费拖慢主链路。详见技术规格文档 6.6 节。
package main

import "log"

func main() {
	// TODO: 启动消息队列消费者，消费用量事件写入 usage_record 表
	// TODO: 暴露内部用量上报接口 + 对外用量查询接口
	// TODO: 定时任务：按能力维度聚合分成，生成结算单
	log.Println("billing-service consumer started")
}
