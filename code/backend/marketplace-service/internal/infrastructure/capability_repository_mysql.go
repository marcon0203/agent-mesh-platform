package infrastructure

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/agentmesh/marketplace-service/internal/domain"
)

// CapabilityMySQLRepository 实现 domain.CapabilityRepository，对接
// infra/schema.sql 的 capability 表。
type CapabilityMySQLRepository struct {
	db *sql.DB
}

func NewCapabilityMySQLRepository(db *sql.DB) *CapabilityMySQLRepository {
	return &CapabilityMySQLRepository{db: db}
}

func capTypeToDB(t domain.CapabilityType) (int, error) {
	switch t {
	case domain.CapabilityTypeTool:
		return 1, nil
	case domain.CapabilityTypeSkill:
		return 2, nil
	case domain.CapabilityTypeAgent:
		return 3, nil
	default:
		return 0, errors.New("unknown capability type: " + string(t))
	}
}

func capTypeFromDB(v int) (domain.CapabilityType, error) {
	switch v {
	case 1:
		return domain.CapabilityTypeTool, nil
	case 2:
		return domain.CapabilityTypeSkill, nil
	case 3:
		return domain.CapabilityTypeAgent, nil
	default:
		return "", errors.New("unknown capability type value in db: " + strconv.Itoa(v))
	}
}

func capStatusToDB(s domain.CapabilityStatus) int {
	switch s {
	case domain.StatusPublished:
		return 2
	case domain.StatusOffline:
		return 3
	default:
		return 1
	}
}

func capStatusFromDB(v int) domain.CapabilityStatus {
	switch v {
	case 2:
		return domain.StatusPublished
	case 3:
		return domain.StatusOffline
	default:
		return domain.StatusPendingReview
	}
}

const capabilitySelectCols = `id, type, name, publisher_id, schema_json, mcp_endpoint, is_builtin, current_version, status`

func (r *CapabilityMySQLRepository) scanCapability(row rowScanner) (*domain.Capability, error) {
	var (
		dbID, publisherID uint64
		typeVal, statusVal int
		name               string
		schemaJSON         sql.NullString
		mcpEndpoint        sql.NullString
		isBuiltin          bool
		currentVersion     sql.NullString
	)
	if err := row.Scan(&dbID, &typeVal, &name, &publisherID, &schemaJSON, &mcpEndpoint, &isBuiltin, &currentVersion, &statusVal); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("capability not found")
		}
		return nil, err
	}
	capType, err := capTypeFromDB(typeVal)
	if err != nil {
		return nil, err
	}
	return domain.RehydrateCapability(
		strconv.FormatUint(dbID, 10), name, capType, strconv.FormatUint(publisherID, 10),
		schemaJSON.String, mcpEndpoint.String, isBuiltin, capStatusFromDB(statusVal), currentVersion.String,
	), nil
}

func (r *CapabilityMySQLRepository) FindByID(id string) (*domain.Capability, error) {
	row := r.db.QueryRow(`SELECT `+capabilitySelectCols+` FROM capability WHERE id = ?`, id)
	return r.scanCapability(row)
}

func (r *CapabilityMySQLRepository) ListPublished(capType domain.CapabilityType) ([]*domain.Capability, error) {
	typeVal, err := capTypeToDB(capType)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(
		`SELECT `+capabilitySelectCols+` FROM capability WHERE status = ? AND type = ? ORDER BY id DESC`,
		capStatusToDB(domain.StatusPublished), typeVal,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Capability
	for rows.Next() {
		cap, err := r.scanCapability(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, cap)
	}
	return result, rows.Err()
}

// Save 是插入或更新的 upsert：内置能力在启动时以固定数值 ID 写入
// （见 builtin_capabilities.go），后续 SubmitForReview/Approve 等状态流转
// 复用同一行，用 ON DUPLICATE KEY UPDATE 覆盖。
func (r *CapabilityMySQLRepository) Save(cap *domain.Capability) error {
	typeVal, err := capTypeToDB(cap.Type())
	if err != nil {
		return err
	}
	publisherID, err := strconv.ParseUint(cap.PublisherID(), 10, 64)
	if err != nil {
		return errors.New("invalid publisher id: " + err.Error())
	}
	id, err := strconv.ParseUint(cap.ID(), 10, 64)
	if err != nil {
		return errors.New("invalid capability id: " + err.Error())
	}

	isBuiltin := 0
	if cap.IsBuiltin() {
		isBuiltin = 1
	}

	_, err = r.db.Exec(
		`INSERT INTO capability (id, type, name, publisher_id, schema_json, mcp_endpoint, is_builtin, current_version, status)
		 VALUES (?,?,?,?,?,?,?,?,?)
		 ON DUPLICATE KEY UPDATE
		   type=VALUES(type), name=VALUES(name), publisher_id=VALUES(publisher_id),
		   schema_json=VALUES(schema_json), mcp_endpoint=VALUES(mcp_endpoint),
		   is_builtin=VALUES(is_builtin), current_version=VALUES(current_version), status=VALUES(status)`,
		id, typeVal, cap.Name(), publisherID, nullableString(cap.SchemaJSON()), nullableString(cap.MCPEndpoint()),
		isBuiltin, nullableString(cap.Version()), capStatusToDB(cap.Status()),
	)
	return err
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
