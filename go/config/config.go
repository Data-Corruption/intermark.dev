package config

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/Data-Corruption/rlog/logger"
	toml "github.com/pelletier/go-toml/v2"
)

type noCopy struct{} // see https://github.com/golang/go/issues/8005#issuecomment-190753527

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type Timeouts struct {
	Tail  int `toml:"tail"`
	Fetch int `toml:"fetch"`
	Reset int `toml:"reset"`
	Lfs   int `toml:"lfs"`
	Lunr  int `toml:"lunr"`
}

type Data struct {
	Addr         string   `toml:"addr"`
	UseSshKey    bool     `toml:"use-ssh-key"    comment:"set to true if repo is private"`
	PageCacheMB  int      `toml:"page-cache-mb"   comment:"size of the cache for page content in MB"`
	AssetCacheMB int      `toml:"asset-cache-mb"  comment:"size of the cache for asset content in MB"`
	LogLevel     string   `toml:"log-level"       comment:"log level: debug, info, warn, error, none"`
	Timeouts     Timeouts `toml:"timeouts"        comment:"timeout values in seconds"`
	UpdateSecret string   `toml:"update-secret"   comment:"secret for update operations"`
}

// Config wraps a TOML config file with safe concurrent access and atomic writes
type Config struct {
	noCopy   noCopy // prevent copying
	mu       sync.RWMutex
	data     Data
	filePath string
}

var ErrNoConfig = errors.New("config not found in context")

var DefaultData = Data{
	Addr:         ":9292",
	PageCacheMB:  1024, // 1GB
	AssetCacheMB: 1024, // 1GB
	LogLevel:     "warn",
	Timeouts: Timeouts{
		Tail:  30,
		Fetch: 30,
		Reset: 30,
		Lfs:   60,
		Lunr:  30,
	},
}

// New loads (or creates) the TOML file at path into memory.
func New(path string) (*Config, error) {
	// create empty file if missing
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// create file with default config
			buf, err := toml.Marshal(DefaultData)
			if err != nil {
				return nil, err
			}
			if err := os.WriteFile(path, buf, 0o644); err != nil {
				return nil, err
			}
			return &Config{data: DefaultData, filePath: path}, nil
		} else {
			return nil, err
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var d Data
	if err := toml.Unmarshal(raw, &d); err != nil {
		return nil, err
	}

	// set default values for missing fields
	if d.Addr == "" {
		d.Addr = DefaultData.Addr
	}
	if d.PageCacheMB == 0 {
		d.PageCacheMB = DefaultData.PageCacheMB
	}
	if d.AssetCacheMB == 0 {
		d.AssetCacheMB = DefaultData.AssetCacheMB
	}
	if d.LogLevel == "" {
		d.LogLevel = DefaultData.LogLevel
	}
	if d.Timeouts.Tail == 0 {
		d.Timeouts.Tail = DefaultData.Timeouts.Tail
	}
	if d.Timeouts.Fetch == 0 {
		d.Timeouts.Fetch = DefaultData.Timeouts.Fetch
	}
	if d.Timeouts.Reset == 0 {
		d.Timeouts.Reset = DefaultData.Timeouts.Reset
	}
	if d.Timeouts.Lfs == 0 {
		d.Timeouts.Lfs = DefaultData.Timeouts.Lfs
	}
	if d.Timeouts.Lunr == 0 {
		d.Timeouts.Lunr = DefaultData.Timeouts.Lunr
	}

	return &Config{data: d, filePath: path}, nil
}

type ctxKey struct{}

func IntoContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, ctxKey{}, cfg)
}

func FromContext(ctx context.Context) *Config {
	if cfg, ok := ctx.Value(ctxKey{}).(*Config); ok {
		return cfg
	}
	return nil
}

func GetData(ctx context.Context) Data {
	cfg := FromContext(ctx)
	if cfg == nil {
		logger.Warnf(ctx, "config not found in context")
		return DefaultData
	}
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	return cfg.data
}

// Txn applies fn to a deep copy of the config,
// then, if fn returns nil, commits and writes atomically.
func Txn(ctx context.Context, fn func(m *Data) error) error {
	cfg := FromContext(ctx)
	if cfg == nil {
		return ErrNoConfig
	}

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	// deep copy via marshal/unmarshal
	buf, err := toml.Marshal(cfg.data)
	if err != nil {
		return err
	}
	var copyData Data
	if err := toml.Unmarshal(buf, &copyData); err != nil {
		return err
	}

	// user mutation
	if err := fn(&copyData); err != nil {
		return err
	}

	// commit
	cfg.data = copyData
	return cfg.save()
}

func (c *Config) save() error {
	buf, err := toml.Marshal(c.data)
	if err != nil {
		return err
	}
	tmp := c.filePath + ".tmp"
	if err := os.WriteFile(tmp, buf, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.filePath)
}
