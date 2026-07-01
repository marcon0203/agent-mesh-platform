package repository

import (
	"database/sql"
	"errors"
	"strconv"
)

// UsageRepository 落库用量事件。billing-service 是简单三层结构（见
// code/backend/README.md），这里直接写 SQL，不做 DDD 仓储接口抽象。
type UsageRepository struct {
	db *sql.DB
}

func NewUsageRepository(db *sql.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

const (
	channelAPI       = 1
	runStatusSuccess = 1
)

// UpsertAgentRun 按 trace_id 查找或创建一条 agent_run 记录，返回其 id。
// orchestration-service 目前每次调用只上报一次汇总用量（不是逐 chunk 上报），
// 所以这里的"upsert"实际上总是走插入分支；保留查找分支是为了在同一
// trace_id 下未来允许多次上报（例如 Hub + Subagent 分别上报）时不重复建行。
func (r *UsageRepository) UpsertAgentRun(agentID, traceID string, tokens int32) (int64, error) {
	agentIDNum, err := strconv.ParseInt(agentID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid agent id: " + err.Error())
	}

	var existingID int64
	err = r.db.QueryRow(`SELECT id FROM agent_run WHERE trace_id = $1`, traceID).Scan(&existingID)
	switch {
	case err == nil:
		if _, err := r.db.Exec(`UPDATE agent_run SET tokens_used = tokens_used + $1 WHERE id = $2`, tokens, existingID); err != nil {
			return 0, err
		}
		return existingID, nil
	case errors.Is(err, sql.ErrNoRows):
		var newID int64
		err := r.db.QueryRow(
			`INSERT INTO agent_run (agent_id, channel, trace_id, tokens_used, status, started_at) VALUES ($1,$2,$3,$4,$5, now()) RETURNING id`,
			agentIDNum, channelAPI, traceID, tokens, runStatusSuccess,
		).Scan(&newID)
		return newID, err
	default:
		return 0, err
	}
}

// InsertUsageRecord 写入 usage_record 表，capabilityID 为空表示这笔消耗
// 记在 Hub 自身账上（Subagent-as-Tool 分账属于 M2 范围，届时 capabilityID 非空）。
func (r *UsageRepository) InsertUsageRecord(agentRunID int64, tokens int32) error {
	_, err := r.db.Exec(
		`INSERT INTO usage_record (agent_run_id, tokens, cost_cents, revenue_share) VALUES ($1,$2,0,0)`,
		agentRunID, tokens,
	)
	return err
}

// SumTokensByAgent 供对外用量查询接口使用。
func (r *UsageRepository) SumTokensByAgent(agentID string) (int64, error) {
	var total sql.NullInt64
	err := r.db.QueryRow(
		`SELECT SUM(tokens_used) FROM agent_run WHERE agent_id = $1`, agentID,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total.Int64, nil
}
