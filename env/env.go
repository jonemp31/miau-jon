package env

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type E struct {
	Port           string `env:"PORT" envDefault:"8080"`
	DebugMode      bool   `env:"DEBUG_MODE" envDefault:"false"`
	DebugWhatsmeow bool   `env:"DEBUG_WHATSMEOW" envDefault:"false"`

	RedisURL      string `env:"REDIS_URL" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisTLS      bool   `env:"REDIS_TLS" envDefault:"false"`

	ApiKey    string `env:"API_KEY" envDefault:""`
	DBDialect string `env:"DIALECT_DB" envDefault:"sqlite3"`                   // sqlite3 or postgres
	DBURL     string `env:"DB_URL" envDefault:"file:data.db?_foreign_keys=on"` // "postgres://<user>:<pass>@<host>:<port>/<DB>?sslmode=disable

	GCSEnabled bool   `env:"GCS_ENABLED" envDefault:"false"`
	GCSBucket  string `env:"GCS_BUCKET" envDefault:"whatsmiau"`
	GCSURL     string `env:"GCS_URL" envDefault:"https://storage.googleapis.com"`

	GCL          string `json:"GCL_APP_NAME" envDefault:"whatsmiau-br-1"`
	GCLEnabled   bool   `json:"GCL_ENABLED" envDefault:"false"`
	GCLProjectID string `json:"GCL_PROJECT_ID"`

	EmitterBufferSize    int `env:"EMITTER_BUFFER_SIZE" envDefault:"2048"`
	HandlerSemaphoreSize int `env:"HANDLER_SEMAPHORE_SIZE" envDefault:"512"`

	ProxyAddresses []string `env:"PROXY_ADDRESSES" envDefault:""`      // random choices proxies ex: <SOCKS5|HTTP|HTTPS>://<username>:<password>@<host>:<port>
	ProxyStrategy  string   `env:"PROXY_STRATEGY" envDefault:"RANDOM"` // todo: implement BALANCED
	ProxyNoMedia   bool     `env:"PROXY_NO_MEDIA" envDefault:"false"`

	// Default values for new instances
	DefaultWebhookURL      string `env:"DEFAULT_WEBHOOK_URL" envDefault:""`
	DefaultAutoReceipt     bool   `env:"DEFAULT_AUTO_RECEIPT" envDefault:"true"`
	DefaultAutoRead        bool   `env:"DEFAULT_AUTO_READ" envDefault:"true"`
	DefaultAlwaysOnline    bool   `env:"DEFAULT_ALWAYS_ONLINE" envDefault:"false"`
	DefaultReadMessages    bool   `env:"DEFAULT_READ_MESSAGES" envDefault:"true"`
	DefaultRejectCalls     bool   `env:"DEFAULT_REJECT_CALLS" envDefault:"false"`
	DefaultMsgCall         string `env:"DEFAULT_MSG_CALL" envDefault:""`
	DefaultSkipGroups      bool   `env:"DEFAULT_SKIP_GROUPS" envDefault:"false"`
	DefaultSkipBroadcasts  bool   `env:"DEFAULT_SKIP_BROADCASTS" envDefault:"false"`
	DefaultSkipOwnMessages bool   `env:"DEFAULT_SKIP_OWN_MESSAGES" envDefault:"false"`
	DefaultWebhookEvents   string `env:"DEFAULT_WEBHOOK_EVENTS" envDefault:"All"`
	DefaultWebhookByEvents bool   `env:"DEFAULT_WEBHOOK_BY_EVENTS" envDefault:"false"`
}

var Env E

func Load() error {
	_ = godotenv.Load(".env")
	err := env.Parse(&Env)

	return err
}
