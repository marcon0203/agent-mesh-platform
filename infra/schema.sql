-- Agent 开放平台核心表结构
-- 完整设计说明见 docs/Agent开放平台_技术规格文档.md 第四章

CREATE TABLE capability (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    type          TINYINT         NOT NULL COMMENT '1=tool 2=skill 3=agent',
    name          VARCHAR(100)    NOT NULL,
    publisher_id  BIGINT UNSIGNED NOT NULL,
    schema_json   JSON            NOT NULL COMMENT '输入输出 JSON Schema',
    mcp_endpoint  VARCHAR(255)             COMMENT 'MCP Server 地址，动态发现用',
    status        TINYINT         NOT NULL DEFAULT 1 COMMENT '1待审核 2已上架 3已下线',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_publisher (publisher_id),
    KEY idx_type_status (type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='能力主表';

CREATE TABLE capability_version (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    capability_id  BIGINT UNSIGNED NOT NULL,
    version        VARCHAR(20)     NOT NULL COMMENT '语义化版本号',
    changelog      TEXT,
    is_latest      TINYINT         NOT NULL DEFAULT 0,
    created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_cap_version (capability_id, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='能力版本表';

CREATE TABLE agent (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name          VARCHAR(100)    NOT NULL,
    owner_id      BIGINT UNSIGNED NOT NULL,
    config_json   JSON            NOT NULL COMMENT '挂载能力列表 + Hook 配置 + Loop 模板',
    version       VARCHAR(20)     NOT NULL,
    status        TINYINT         NOT NULL DEFAULT 1 COMMENT '1开发中 2已发布 3已下线',
    max_depth     TINYINT         NOT NULL DEFAULT 3 COMMENT '子 Agent 最大递归深度',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent 主表';

CREATE TABLE agent_run (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    agent_id      BIGINT UNSIGNED NOT NULL,
    channel       TINYINT         NOT NULL COMMENT '1=api 2=workbench',
    trace_id      VARCHAR(64)     NOT NULL,
    caller_key_id BIGINT UNSIGNED          COMMENT '发起调用的 API Key，Workbench 场景为空',
    tokens_used   INT UNSIGNED    NOT NULL DEFAULT 0,
    status        TINYINT         NOT NULL COMMENT '1成功 2失败 3限流拒绝',
    started_at    DATETIME        NOT NULL,
    PRIMARY KEY (id),
    KEY idx_agent_time (agent_id, started_at),
    KEY idx_trace (trace_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent 运行记录表';

CREATE TABLE api_key (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    developer_id  BIGINT UNSIGNED NOT NULL,
    key_prefix    VARCHAR(12)     NOT NULL COMMENT '展示用前缀，如 sk-abc1',
    key_hash      VARCHAR(64)     NOT NULL COMMENT 'Key 的哈希值，不存明文',
    scopes        JSON            NOT NULL COMMENT '如 ["agent:123:invoke"]',
    status        TINYINT         NOT NULL DEFAULT 1 COMMENT '1启用 2已吊销',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_key_hash (key_hash),
    KEY idx_developer (developer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='开发者 API Key 表';

CREATE TABLE subscription (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    consumer_agent_id BIGINT UNSIGNED NOT NULL COMMENT '发起挂载的 Agent（Hub）',
    capability_id     BIGINT UNSIGNED NOT NULL COMMENT '被挂载的能力（Tool/Skill/Agent）',
    pinned_version    VARCHAR(20)     NOT NULL COMMENT '锁定版本，默认不跟随最新',
    status            TINYINT         NOT NULL DEFAULT 1,
    created_at        DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_consumer_cap (consumer_agent_id, capability_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='能力挂载订阅表';

CREATE TABLE usage_record (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    agent_run_id    BIGINT UNSIGNED NOT NULL,
    capability_id   BIGINT UNSIGNED          COMMENT '若该次消耗来自某个被调用能力，记录归属',
    tokens          INT UNSIGNED    NOT NULL DEFAULT 0,
    cost_cents      INT UNSIGNED    NOT NULL DEFAULT 0,
    revenue_share   INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '分给能力提供方的金额（分）',
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_run (agent_run_id),
    KEY idx_capability (capability_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用量与分成记录表';
