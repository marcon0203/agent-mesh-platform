package domain

import "errors"

var (
	ErrEmptyProviderName   = errors.New("model provider name must not be empty")
	ErrEmptyProviderAPIKey = errors.New("model provider api key must not be empty")
	ErrEmptyProviderModel  = errors.New("model provider model name must not be empty")
	ErrProviderBaseURLReq  = errors.New("openai_compatible provider requires base_url")
	ErrUnknownProviderType = errors.New("unknown model provider type")
)

// ProviderType 对应【需人工决策】的落地方案：MVP 阶段支持 OpenAI 兼容协议
// （覆盖大多数可自建 base_url 的模型服务）和 Anthropic 原生协议两种。
type ProviderType string

const (
	ProviderTypeOpenAICompatible ProviderType = "openai_compatible"
	ProviderTypeAnthropic        ProviderType = "anthropic"
)

type ModelProviderStatus string

const (
	ModelProviderStatusEnabled  ModelProviderStatus = "enabled"
	ModelProviderStatusDisabled ModelProviderStatus = "disabled"
)

// ModelProvider 是用户自建的模型供应商配置聚合根：用户提供 API Key + 模型名，
// Agent 发布时引用某个 ModelProvider 的 ID，决定 infrastructure 层的
// EinoRuntime 具体用哪种 Eino ChatModel 组件、以什么凭证初始化。
//
// 明文 APIKey 只存在于聚合根的内存态；是否加密、如何加密是
// infrastructure 层持久化时的职责，不属于领域规则。
type ModelProvider struct {
	id           string
	ownerID      string
	name         string
	providerType ProviderType
	baseURL      string
	apiKey       string
	modelName    string
	status       ModelProviderStatus
}

// NewModelProvider 是唯一合法的构造入口。
func NewModelProvider(id, ownerID, name string, providerType ProviderType, baseURL, apiKey, modelName string) (*ModelProvider, error) {
	if name == "" {
		return nil, ErrEmptyProviderName
	}
	if apiKey == "" {
		return nil, ErrEmptyProviderAPIKey
	}
	if modelName == "" {
		return nil, ErrEmptyProviderModel
	}
	switch providerType {
	case ProviderTypeOpenAICompatible:
		if baseURL == "" {
			return nil, ErrProviderBaseURLReq
		}
	case ProviderTypeAnthropic:
		// baseURL 可选：为空时 infrastructure 层使用 Anthropic 官方端点
	default:
		return nil, ErrUnknownProviderType
	}
	return &ModelProvider{
		id:           id,
		ownerID:      ownerID,
		name:         name,
		providerType: providerType,
		baseURL:      baseURL,
		apiKey:       apiKey,
		modelName:    modelName,
		status:       ModelProviderStatusEnabled,
	}, nil
}

// RehydrateModelProvider 供仓储层从持久化存储重建聚合根，
// 跳过构造校验（数据在写入时已经校验过一次）。
func RehydrateModelProvider(id, ownerID, name string, providerType ProviderType, baseURL, apiKey, modelName string, status ModelProviderStatus) *ModelProvider {
	return &ModelProvider{
		id:           id,
		ownerID:      ownerID,
		name:         name,
		providerType: providerType,
		baseURL:      baseURL,
		apiKey:       apiKey,
		modelName:    modelName,
		status:       status,
	}
}

func (p *ModelProvider) Disable() { p.status = ModelProviderStatusDisabled }
func (p *ModelProvider) Enable()  { p.status = ModelProviderStatusEnabled }

func (p *ModelProvider) ID() string                 { return p.id }
func (p *ModelProvider) OwnerID() string             { return p.ownerID }
func (p *ModelProvider) Name() string                { return p.name }
func (p *ModelProvider) Type() ProviderType          { return p.providerType }
func (p *ModelProvider) BaseURL() string             { return p.baseURL }
func (p *ModelProvider) APIKey() string              { return p.apiKey }
func (p *ModelProvider) ModelName() string           { return p.modelName }
func (p *ModelProvider) Status() ModelProviderStatus { return p.status }
func (p *ModelProvider) IsEnabled() bool             { return p.status == ModelProviderStatusEnabled }

// ModelProviderRepository 是领域层定义、infrastructure 层实现的仓储接口。
type ModelProviderRepository interface {
	FindByID(id string) (*ModelProvider, error)
	ListByOwner(ownerID string) ([]*ModelProvider, error)
	Save(provider *ModelProvider) (*ModelProvider, error)
}
