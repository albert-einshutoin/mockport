package config

// DefaultPort is the default Mockport listen port used by generated
// configs, FakeEnv URLs, and the report CLI.
const DefaultPort = 43101

type Config struct {
	Version        string                   `yaml:"version" json:"version"`
	Server         ServerConfig             `yaml:"server" json:"server"`
	Mode           string                   `yaml:"mode" json:"mode"`
	Adapters       map[string]AdapterConfig `yaml:"adapters" json:"adapters"`
	Scenarios      map[string]Scenario      `yaml:"scenarios" json:"scenarios,omitempty"`
	SafetyWarnings []SafetyWarning          `yaml:"-" json:"safety_warnings,omitempty"`
}

type ServerConfig struct {
	Host string `yaml:"host" json:"host"`
	Port int    `yaml:"port" json:"port"`
}

type AdapterConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled"`
	BasePath   string        `yaml:"base_path" json:"base_path"`
	Scenario   string        `yaml:"scenario" json:"scenario"`
	FakeSecret string        `yaml:"fake_secret" json:"-"`
	APIURL     string        `yaml:"api_url" json:"api_url,omitempty"`
	Webhook    WebhookConfig `yaml:"webhook" json:"webhook,omitempty"`
}

type WebhookConfig struct {
	TargetURL     string `yaml:"target_url" json:"target_url,omitempty"`
	SigningSecret string `yaml:"signing_secret" json:"-"`
}

type Scenario struct {
	Adapter  string         `yaml:"adapter" json:"adapter"`
	Response map[string]any `yaml:"response" json:"response,omitempty"`
	Webhook  map[string]any `yaml:"webhook" json:"webhook,omitempty"`
}

type SafetyWarning struct {
	Field    string `json:"field"`
	Category string `json:"category"`
	Message  string `json:"message"`
}
