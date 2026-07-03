package infrastructure

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/agentmesh/orchestration-service/internal/domain"
)

// AgentMySQLRepository 实现 domain.AgentRepository，对接 infra/schema.sql 的
// agent 表（PostgreSQL）。挂载能力列表 + Hook 配置以 JSON 形式存进 config_json 列，
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
		dbID, ownerID    int64
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
		strconv.FormatInt(dbID, 10), name, cfg.LoopTemplate, cfg.Capabilities, cfg.Hooks,
		maxDepth, version, agentStatusFromDB(statusVal), modelProviderID,
	), nil
}

const agentSelectCols = `id, name, owner_id, config_json, version, status, max_depth, model_provider_id`

func (r *AgentMySQLRepository) FindByID(id string) (*domain.Agent, error) {
	row := r.db.QueryRow(`SELECT `+agentSelectCols+` FROM agent WHERE id = $1`, id)
	return r.scanAgent(row)
}

// ListPublished 返回所有已发布的 Agent，供前端 Agent 构建器挑选"挂载哪个 Agent
// 作为 Subagent"时使用，也供 eino_runtime.go 校验一个 Subagent 挂载是否合法
// （只有已发布的 Agent 才允许被递归调用）。
func (r *AgentMySQLRepository) ListPublished() ([]*domain.Agent, error) {
	rows, err := r.db.Query(`SELECT `+agentSelectCols+` FROM agent WHERE status = $1 ORDER BY id DESC`, agentStatusToDB(domain.AgentStatusPublished))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*domain.Agent
	for rows.Next() {
		agent, err := r.scanAgent(rows)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, rows.Err()
}

// ListAll 返回不论状态的全部 Agent，供前端 Agent 列表页使用——草稿态的也要能
// 看到，不然创建到一半没发布的 Agent 就再也找不回来了。
func (r *AgentMySQLRepository) ListAll() ([]*domain.Agent, error) {
	rows, err := r.db.Query(`SELECT ` + agentSelectCols + ` FROM agent ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*domain.Agent
	for rows.Next() {
		agent, err := r.scanAgent(rows)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, rows.Err()
}

// Save 插入或更新一条记录；新建 Agent（ID 为空）时由 Postgres 自增列生成 ID，
// 通过返回值把生成的 ID 回填到一个新的聚合根实例上（聚合根本身不暴露 ID setter，
// 与 ModelProviderRepository.Save 的模式保持一致）。
func (r *AgentMySQLRepository) Save(agent *domain.Agent) (*domain.Agent, error) {
	cfg := agentConfigJSON{
		LoopTemplate:    agent.LoopTemplate(),
		Capabilities:    agent.Capabilities(),
		Hooks:           agent.Hooks(),
		ModelProviderID: agent.ModelProviderID(),
	}
	configRaw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	var modelProviderID sql.NullInt64
	if agent.ModelProviderID() != "" {
		id, err := strconv.ParseInt(agent.ModelProviderID(), 10, 64)
		if err != nil {
			return nil, errors.New("invalid model provider id: " + err.Error())
		}
		modelProviderID = sql.NullInt64{Int64: id, Valid: true}
	}

	// 领域模型目前还没有 Agent 归属者的概念（见 domain/agent.go），owner_id 暂时固定为
	// 平台默认账号；等账号体系接入后这里需要从调用方上下文传入真实 owner_id。
	const placeholderOwnerID = 1

	if agent.ID() == "" {
		var newID int64
		err := r.db.QueryRow(
			`INSERT INTO agent (name, owner_id, config_json, version, status, max_depth, model_provider_id)
			 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
			agent.Name(), placeholderOwnerID, configRaw, agent.Version(), agentStatusToDB(agent.Status()), agent.MaxDepth(), modelProviderID,
		).Scan(&newID)
		if err != nil {
			return nil, err
		}
		return domain.RehydrateAgent(
			strconv.FormatInt(newID, 10), agent.Name(), agent.LoopTemplate(), agent.Capabilities(),
			agent.Hooks(), agent.MaxDepth(), agent.Version(), agent.Status(), agent.ModelProviderID(),
		), nil
	}

	_, err = r.db.Exec(
		`UPDATE agent SET name=$1, config_json=$2, version=$3, status=$4, max_depth=$5, model_provider_id=$6 WHERE id=$7`,
		agent.Name(), configRaw, agent.Version(), agentStatusToDB(agent.Status()), agent.MaxDepth(), modelProviderID, agent.ID(),
	)
	if err != nil {
		return nil, err
	}
	return agent, nil
}
