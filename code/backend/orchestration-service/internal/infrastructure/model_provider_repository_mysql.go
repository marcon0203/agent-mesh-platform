package infrastructure

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/agentmesh/orchestration-service/internal/domain"
)

// ModelProviderMySQLRepository 实现 domain.ModelProviderRepository，对接
// infra/schema.sql 里的 model_provider 表。API Key 在写入前用 apiKeyCipher
// 加密，读出后解密，明文只在内存态的聚合根里出现。
type ModelProviderMySQLRepository struct {
	db     *sql.DB
	cipher *apiKeyCipher
}

func NewModelProviderMySQLRepository(db *sql.DB) (*ModelProviderMySQLRepository, error) {
	c, err := newAPIKeyCipher()
	if err != nil {
		return nil, err
	}
	return &ModelProviderMySQLRepository{db: db, cipher: c}, nil
}

func providerTypeToDB(t domain.ProviderType) (int, error) {
	switch t {
	case domain.ProviderTypeOpenAICompatible:
		return 1, nil
	case domain.ProviderTypeAnthropic:
		return 2, nil
	default:
		return 0, errors.New("unknown provider type: " + string(t))
	}
}

func providerTypeFromDB(v int) (domain.ProviderType, error) {
	switch v {
	case 1:
		return domain.ProviderTypeOpenAICompatible, nil
	case 2:
		return domain.ProviderTypeAnthropic, nil
	default:
		return "", errors.New("unknown provider_type value in db: " + strconv.Itoa(v))
	}
}

func providerStatusToDB(s domain.ModelProviderStatus) int {
	if s == domain.ModelProviderStatusDisabled {
		return 2
	}
	return 1
}

func providerStatusFromDB(v int) domain.ModelProviderStatus {
	if v == 2 {
		return domain.ModelProviderStatusDisabled
	}
	return domain.ModelProviderStatusEnabled
}

func (r *ModelProviderMySQLRepository) scanProvider(row rowScanner) (*domain.ModelProvider, error) {
	var (
		dbID, ownerID              uint64
		name, modelName            string
		baseURL                    sql.NullString
		providerTypeVal, statusVal int
		cipherBytes                []byte
	)
	if err := row.Scan(&dbID, &ownerID, &name, &providerTypeVal, &baseURL, &modelName, &cipherBytes, &statusVal); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("model provider not found")
		}
		return nil, err
	}
	providerType, err := providerTypeFromDB(providerTypeVal)
	if err != nil {
		return nil, err
	}
	apiKey, err := r.cipher.Decrypt(cipherBytes)
	if err != nil {
		return nil, err
	}
	return domain.RehydrateModelProvider(
		strconv.FormatUint(dbID, 10),
		strconv.FormatUint(ownerID, 10),
		name, providerType, baseURL.String, apiKey, modelName, providerStatusFromDB(statusVal),
	), nil
}

const modelProviderSelectCols = `id, owner_id, name, provider_type, base_url, model_name, api_key_cipher, status`

func (r *ModelProviderMySQLRepository) FindByID(id string) (*domain.ModelProvider, error) {
	row := r.db.QueryRow(`SELECT `+modelProviderSelectCols+` FROM model_provider WHERE id = ?`, id)
	return r.scanProvider(row)
}

func (r *ModelProviderMySQLRepository) ListByOwner(ownerID string) ([]*domain.ModelProvider, error) {
	rows, err := r.db.Query(`SELECT `+modelProviderSelectCols+` FROM model_provider WHERE owner_id = ? ORDER BY id DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.ModelProvider
	for rows.Next() {
		provider, err := r.scanProvider(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, provider)
	}
	return result, rows.Err()
}

// Save 插入或更新一条记录。新建时（ID 为空）由 MySQL 自增列生成 ID，
// 通过返回值把生成的 ID 回填到一个新的聚合根实例上（聚合根本身不暴露 ID setter）。
func (r *ModelProviderMySQLRepository) Save(provider *domain.ModelProvider) (*domain.ModelProvider, error) {
	typeVal, err := providerTypeToDB(provider.Type())
	if err != nil {
		return nil, err
	}
	cipherBytes, err := r.cipher.Encrypt(provider.APIKey())
	if err != nil {
		return nil, err
	}
	statusVal := providerStatusToDB(provider.Status())
	baseURL := sql.NullString{String: provider.BaseURL(), Valid: provider.BaseURL() != ""}

	if provider.ID() == "" {
		ownerID, err := strconv.ParseUint(provider.OwnerID(), 10, 64)
		if err != nil {
			return nil, errors.New("invalid owner id: " + err.Error())
		}
		res, err := r.db.Exec(
			`INSERT INTO model_provider (owner_id, name, provider_type, base_url, model_name, api_key_cipher, status) VALUES (?,?,?,?,?,?,?)`,
			ownerID, provider.Name(), typeVal, baseURL, provider.ModelName(), cipherBytes, statusVal,
		)
		if err != nil {
			return nil, err
		}
		newID, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		return domain.RehydrateModelProvider(
			strconv.FormatInt(newID, 10), provider.OwnerID(), provider.Name(), provider.Type(),
			provider.BaseURL(), provider.APIKey(), provider.ModelName(), provider.Status(),
		), nil
	}

	_, err = r.db.Exec(
		`UPDATE model_provider SET name=?, provider_type=?, base_url=?, model_name=?, api_key_cipher=?, status=? WHERE id=?`,
		provider.Name(), typeVal, baseURL, provider.ModelName(), cipherBytes, statusVal, provider.ID(),
	)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
