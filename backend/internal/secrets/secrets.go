package secrets

// Credentials 包含了連線到外部服務所需的認證資訊
type Credentials struct {
	Username string
	Password string
	Token    string
}

// SecretsManager 是憑證管理的介面，符合 Factory Provider 模式
type SecretsManager interface {
	// GetCredentials 根據一個引用路徑（例如 Vault 的路徑）來獲取憑證
	GetCredentials(ref string) (*Credentials, error)
}
