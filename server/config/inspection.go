package config

type Inspection struct {
	KubeConfigDir string `mapstructure:"kube-config-dir" json:"kubeConfigDir" yaml:"kube-config-dir"`
	ForceStub     bool   `mapstructure:"force-stub" json:"forceStub" yaml:"force-stub"`
	WebhookURL    string `mapstructure:"webhook-url" json:"webhookUrl" yaml:"webhook-url"`
	AIBaseURL     string `mapstructure:"ai-base-url" json:"aiBaseUrl" yaml:"ai-base-url"`
	AIAPIKey      string `mapstructure:"ai-api-key" json:"aiApiKey" yaml:"ai-api-key"`
	AIModel       string `mapstructure:"ai-model" json:"aiModel" yaml:"ai-model"`
}
