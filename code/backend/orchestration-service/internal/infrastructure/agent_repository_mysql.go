package infrastructure

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/agentmesh/orchestration-service/internal/domain"
)

// AgentMySQLRepository 实现 domain.AgentRepository，对接 infra/schema.sql 的
// agent 表。挂载能力列表 + Hook 配置以 JSON 形式存进 config_json 列，
// 与技术规格文档第四章的表结构注释一致，避免为每个字段单独建关联表。
type AgentMySQLRepository struct {
	db *sql.DB
}

func NewAgentMySQLRepository(db *sql.DB) *AgentMySQLRepository {
	return &AgentMySQLRepository{db: db}
}

// agentConfigJSON 是 config_json 列的序列化结构。
type agentConfigJSON struct {
	LoopTemplate    domain.LoopTemplate        `json:"loop_template"`
	Capabilities    []domain.MountedCapability `json:"capabilities"`
	Hooks           domain.HookConfig          `json:"hooks"`
	ModelProviderID string                     `json:"model_provider_id"`
}

func agentStatusToDB(s domain.AgentStatus) int {
	switch s {
	case domain.AgentStatusPublished:
		return 2
	case domain.AgentStatusOffline:
		return 3
	default:
		return 1
	}
}

func agentStatusFromDB(v int) domain.AgentStatus {
	switch v {
	case 2:
		return domain.AgentStatusPublished
	case 3:
		return domain.AgentStatusOffline
	default:
		return domain.AgentStatusDraft
	}
}

func (r *AgentMySQLRepository) scanAgent(row rowScanner) (*domain.Agent, error) {
	var (
		dbID, ownerID    uint64
		name, version    string
		configRaw        []byte
		statusVal        int
		maxDepth         int32
		modelProviderRaw sql.NullInt64
	)
	if err := row.Scan(&dbID, &name, &ownerID, &configRaw, &version, &statusVal, &maxDepth, &modelProviderRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("agent not found")
		}
		return nil, err
	}

	var cfg agentConfigJSON
	if len(configRaw) > 0 {
		if err := json.Unmarshal(configRaw, &cfg); err != nil {
			return nil, err
		}
	}

	modelProviderID := cfg.ModelProviderID
	if modelProviderID == "" && modelProviderRaw.Valid {
		modelProviderID = strconv.FormatInt(modelProviderRaw.Int64, 10)
	}

	return domain.RehydrateAgent(
		strconv.FormatUint(dbID, 10), name, cfg.LoopTemplate, cfg.Capabilities, cfg.Hooks,
		maxDepth, version, agentStatusFromDB(statusVal), modelProviderID,
	), nil
}

const agentSelectCols = `id, name, owner_id, config_json, version, status, max_depth, model_provider_id`

func (r *AgentMySQLRepository) FindByID(id string) (*domain.Agent, error) {
	row := r.db.QueryRow(`SELECT `+agentSelectCols+` FROM agent WHERE id = ?`, id)
	return r.scanAgent(row)
}

// Save 插入或更新一条记录；新建 Agent（ID 为空）时由 MySQL 自增列生成 ID。
// domain.Agent 没有暴露 ID 的 setter，调用方需要用返回值里的新 ID 重新
// FindByID 来获得带 ID 的聚合根（与 ModelProviderRepository.Save 的模式保持一致，
// 这里为了不改动已有的 domain.AgentRepository.Save(agent) error 签名，
// 通过 owner 传入的 id 为空时退化为"仅插入，调用方随后自行查询"）。
func (r *AgentMySQLRepository) Save(agent *domain.Agent) error {
	cfg := agentConfigJSON{
		LoopTemplate:    agent.LoopTemplate(),
		Capabilities:    agent.Capabilities(),
		Hooks:           agent.Hooks(),
		ModelProviderID: agent.ModelProviderID(),
	}
	configRaw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	var modelProviderID sql.NullInt64
	if agent.ModelProviderID() != "" {
		id, err := strconv.ParseInt(agent.ModelProviderID(), 10, 64)
		if err != nil {
			return errors.New("invalid model provider id: " + err.Error())
		}
		modelProviderID = sql.NullInt64{Int64: id, Valid: true}
	}

	if agent.ID() == "" {
		_, err := r.db.Exec(
			`INSERT INTO agent (name, owner_id, config_json, version, status, max_depth, model_provider_id) VALUES (?,?,?,?,?,?,?)`,
			agent.Name(), 0, configRaw, agent.Version(), agentStatusToDB(agent.Status()), agent.MaxDepth(), modelProviderID,
		)
		return err
	}

	_, err = r.db.Exec(
		`UPDATE agent SET name=?, config_json=?, version=?, status=?, max_depth=?, model_provider_id=? WHERE id=?`,
		agent.Name(), configRaw, agent.Version(), agentStatusToDB(agent.Status()), agent.MaxDepth(), modelProviderID, agent.ID(),
	)
	return err
}
