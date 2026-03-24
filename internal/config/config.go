package config

import (
	"file-storage/internal/errs"
	"file-storage/internal/imgproc"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

const (
	LogTypeJSON = "json"
	LogTypeText = "text"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

const (
	StorageFileSystem = "filesystem"
	StorageInmemory   = "inmemory"
)

type App struct {
	Server   Server   `json:"server" yaml:"server"`
	Timeouts Timeouts `json:"timeouts" yaml:"timeouts"`
	Limits   Limits   `json:"limits" yaml:"limits"`
	Security Security `json:"security" yaml:"security"`
	Storage  string   `json:"storage" yaml:"storage"`
}

type Server struct {
	Host           string `json:"host" yaml:"host"`
	Port           int    `json:"port" yaml:"port"`
	MaxHeaderBytes int    `json:"max_header_bytes" yaml:"max_header_bytes"`
}

type Timeouts struct {
	HandlerTimeout    time.Duration `json:"handler_timeout" yaml:"handler_timeout"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout" yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout       time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
}

type Limits struct {
	RateLimiter      RateLimiter `json:"rate_limiter" yaml:"rate_limiter"`
	ConcurrencyLimit int         `json:"concurrency_limit" yaml:"concurrency_limit"`
	SizeLimit        int         `json:"size_limit" yaml:"size_limit"`
}

type Log struct {
	Level string `json:"level" yaml:"level"`
	Type  string `json:"type" yaml:"type"`
}

type Security struct {
	ReadToken  string `json:"read_token" yaml:"read_token"`
	WriteToken string `json:"write_token" yaml:"write_token"`
}

func (s *Security) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("read_token", "***"),
		slog.String("write_token", "***"),
	)
}

type RateLimiter struct {
	Capacity   int `json:"capacity" yaml:"capacity"`
	RefillRate int `json:"refill_rate" yaml:"refill_rate"`
}

type Image struct {
	Ext          string `json:"ext" yaml:"ext"`
	MaxDimension int    `json:"width" yaml:"max_dimension"`
}

type FileSystem struct {
	Path string `json:"path" yaml:"path"`
}

type Storage struct {
	FileSystem FileSystem `json:"filesystem" yaml:"filesystem"`
}

type Config struct {
	App     App     `json:"app" yaml:"app"`
	Log     Log     `json:"log" yaml:"log"`
	Image   Image   `json:"image" yaml:"image"`
	Storage Storage `json:"storage" yaml:"storage"`
}

// NewConfig loads, merges and validates app configuration
func NewConfig(configPath string) (*Config, error) {

	cfg := defaults()

	err := applyConfigFile(&cfg, configPath)
	if err != nil {
		return nil, err
	}

	err = applyEnv(&cfg)
	if err != nil {
		return nil, err
	}

	err = applyFlags(&cfg)
	if err != nil {
		return nil, err
	}

	normalize(&cfg)
	err = validate(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func defaults() Config {
	const defaultSizeLimit = 10 * 1024 * 1024
	cfg := Config{
		Log: Log{
			Level: LogLevelInfo,
			Type:  LogTypeJSON,
		},
		App: App{
			Server: Server{
				Host:           "127.0.0.1",
				Port:           8080,
				MaxHeaderBytes: 65536,
			},
			Timeouts: Timeouts{
				HandlerTimeout:    5 * time.Second,
				ReadHeaderTimeout: 1 * time.Second,
				WriteTimeout:      5 * time.Second,
				IdleTimeout:       10 * time.Second,
			},
			Limits: Limits{
				SizeLimit: defaultSizeLimit,
				RateLimiter: RateLimiter{
					Capacity:   0,
					RefillRate: 0,
				},
			},
			Security: Security{
				ReadToken:  "default token",
				WriteToken: "default token",
			},
		},
		Image: Image{
			Ext:          "jpeg",
			MaxDimension: 2000},
	}
	return cfg
}

func applyConfigFile(cfg *Config, configPath string) error {
	if configPath == "" {
		return nil
	}

	_, err := os.Stat(configPath)
	if err != nil {
		return err
	}

	b, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return err
	}

	return nil
}

func applyEnv(cfg *Config) error {

	v, ok, err := readIntEnv("FILE_STORAGE_APP_PORT")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Server.Port = v
	}

	sAppHost := os.Getenv("FILE_STORAGE_APP_HOST")
	if sAppHost != "" {
		cfg.App.Server.Host = sAppHost
	}

	v, ok, err = readIntEnv("FILE_STORAGE_MAX_HEADER_BYTES")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Server.MaxHeaderBytes = v
	}

	d, ok, err := readDurationEnv("FILE_STORAGE_HANDLER_TIMEOUT")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Timeouts.HandlerTimeout = d
	}

	d, ok, err = readDurationEnv("FILE_STORAGE_READ_HEADER_TIMEOUT")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Timeouts.ReadHeaderTimeout = d
	}

	d, ok, err = readDurationEnv("FILE_STORAGE_WRITE_TIMEOUT")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Timeouts.WriteTimeout = d
	}

	d, ok, err = readDurationEnv("FILE_STORAGE_IDLE_TIMEOUT")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Timeouts.IdleTimeout = d
	}

	v, ok, err = readIntEnv("FILE_STORAGE_SIZE_LIMIT")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Limits.SizeLimit = v
	}

	v, ok, err = readIntEnv("FILE_STORAGE_RATE_LIMITER_CAPACITY")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Limits.RateLimiter.Capacity = v
	}

	v, ok, err = readIntEnv("FILE_STORAGE_RATE_LIMITER_REFILL_RATE")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Limits.RateLimiter.RefillRate = v
	}

	v, ok, err = readIntEnv("FILE_STORAGE_CONCURRENCY_LIMIT")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Limits.ConcurrencyLimit = v
	}

	sReadToken := os.Getenv("FILE_STORAGE_READ_TOKEN")
	if sReadToken != "" {
		cfg.App.Security.ReadToken = sReadToken
	}

	sWriteToken := os.Getenv("FILE_STORAGE_WRITE_TOKEN")
	if sWriteToken != "" {
		cfg.App.Security.WriteToken = sWriteToken
	}

	sLogLevel := os.Getenv("FILE_STORAGE_LOG_LEVEL")
	if sLogLevel != "" {
		cfg.Log.Level = sLogLevel
	}

	sLogType := os.Getenv("FILE_STORAGE_LOG_TYPE")
	if sLogType != "" {
		cfg.Log.Type = sLogType
	}

	sImageExt := os.Getenv("FILE_STORAGE_IMAGE_EXT")
	if sImageExt != "" {
		cfg.Image.Ext = sImageExt
	}

	v, ok, err = readIntEnv("FILE_STORAGE_IMAGE_MAX_DIMENSION")
	if err != nil {
		return err
	}
	if ok {
		cfg.Image.MaxDimension = v
	}

	sStorage := os.Getenv("FILE_STORAGE_STORAGE")
	if sStorage != "" {
		cfg.App.Storage = sStorage
	}

	sFsPath := os.Getenv("FILE_STORAGE_FS_PATH")
	if sFsPath != "" {
		cfg.Storage.FileSystem.Path = sFsPath
	}

	return nil
}

func readIntEnv(name string) (int, bool, error) {
	sValue := os.Getenv(name)
	if sValue != "" {
		v, err := strconv.Atoi(sValue)
		if err != nil {
			return 0, false, fmt.Errorf("invalid %s=%q: %w", name, sValue, err)
		}
		return v, true, nil
	}
	return 0, false, nil
}

func readDurationEnv(name string) (time.Duration, bool, error) {
	sValue := os.Getenv(name)
	if sValue != "" {
		v, err := time.ParseDuration(sValue)
		if err != nil {
			return 0, false, fmt.Errorf("invalid %s=%q: %w", name, sValue, err)
		}
		return v, true, nil
	}
	return 0, false, nil
}

func applyFlags(cfg *Config) error {

	if !pflag.Parsed() {
		return errs.ErrConfigFlagsNotParsed
	}

	v, ok, err := readIntFlag("port")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Server.Port = v
	}

	fHost := pflag.Lookup("host")
	if fHost != nil && fHost.Changed {
		cfg.App.Server.Host = fHost.Value.String()
	}

	v, ok, err = readIntFlag("max-header-bytes")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Server.MaxHeaderBytes = v
	}

	d, ok, err := readDurationFlag("handler-timeout")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Timeouts.HandlerTimeout = d
	}

	d, ok, err = readDurationFlag("read-header-timeout")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Timeouts.ReadHeaderTimeout = d
	}

	d, ok, err = readDurationFlag("write-timeout")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Timeouts.WriteTimeout = d
	}

	d, ok, err = readDurationFlag("idle-timeout")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Timeouts.IdleTimeout = d
	}

	v, ok, err = readIntFlag("size-limit")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Limits.SizeLimit = v
	}

	v, ok, err = readIntFlag("rate-capacity")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Limits.RateLimiter.Capacity = v
	}

	v, ok, err = readIntFlag("rate-refill")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Limits.RateLimiter.RefillRate = v
	}

	v, ok, err = readIntFlag("concurrency-limit")
	if err != nil {
		return err
	}
	if ok {
		cfg.App.Limits.ConcurrencyLimit = v
	}

	fReadToken := pflag.Lookup("read-token")
	if fReadToken != nil && fReadToken.Changed {
		cfg.App.Security.ReadToken = fReadToken.Value.String()
	}

	fWriteToken := pflag.Lookup("write-token")
	if fWriteToken != nil && fWriteToken.Changed {
		cfg.App.Security.WriteToken = fWriteToken.Value.String()
	}

	fLogLevel := pflag.Lookup("log-level")
	if fLogLevel != nil && fLogLevel.Changed {
		cfg.Log.Level = fLogLevel.Value.String()
	}

	fLogType := pflag.Lookup("log-type")
	if fLogType != nil && fLogType.Changed {
		cfg.Log.Type = fLogType.Value.String()
	}

	fImageExt := pflag.Lookup("image-ext")
	if fImageExt != nil && fImageExt.Changed {
		cfg.Image.Ext = fImageExt.Value.String()
	}

	v, ok, err = readIntFlag("image-max-dimension")
	if err != nil {
		return err
	}
	if ok {
		cfg.Image.MaxDimension = v
	}

	fStorage := pflag.Lookup("storage")
	if fStorage != nil && fStorage.Changed {
		cfg.App.Storage = fStorage.Value.String()
	}

	fFsstoragepath := pflag.Lookup("fs-storage-path")
	if fFsstoragepath != nil && fFsstoragepath.Changed {
		cfg.Storage.FileSystem.Path = fFsstoragepath.Value.String()
	}

	return nil
}

func readIntFlag(name string) (int, bool, error) {
	fValue := pflag.Lookup(name)
	if fValue != nil && fValue.Changed {
		raw := fValue.Value.String()
		v, err := strconv.Atoi(raw)
		if err != nil {
			return 0, false, fmt.Errorf("invalid flag %s=%q: %w", name, raw, err)
		}
		return v, true, nil
	}
	return 0, false, nil
}

func readDurationFlag(name string) (time.Duration, bool, error) {
	fValue := pflag.Lookup(name)
	if fValue != nil && fValue.Changed {
		raw := fValue.Value.String()
		v, err := time.ParseDuration(raw)
		if err != nil {
			return 0, false, fmt.Errorf("invalid flag %s=%q: %w", name, raw, err)
		}
		return v, true, nil
	}
	return 0, false, nil
}

func normalize(cfg *Config) {
	cfg.Log.Type = strings.ToLower(cfg.Log.Type)
	cfg.Log.Level = strings.ToLower(cfg.Log.Level)
}

func validate(cfg *Config) error {
	if cfg.App.Server.Host == "" {
		return errs.ErrConfigHostNotSet
	}

	if cfg.App.Server.Port < 1 || cfg.App.Server.Port > 65535 {
		return errs.ErrConfigPortOutOfRange
	}

	if cfg.App.Server.MaxHeaderBytes < 0 || cfg.App.Server.MaxHeaderBytes > 1024*1024 {
		return errs.ErrConfigMaxHeaderBytesOutOfRange
	}

	if cfg.App.Timeouts.HandlerTimeout <= 0 {
		return fmt.Errorf("handler timeout: %w", errs.ErrConfigInvalidTimeout)
	}
	if cfg.App.Timeouts.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("read header timeout: %w", errs.ErrConfigInvalidTimeout)
	}
	if cfg.App.Timeouts.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout: %w", errs.ErrConfigInvalidTimeout)
	}
	if cfg.App.Timeouts.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout: %w", errs.ErrConfigInvalidTimeout)
	}

	capacityEnabled := cfg.App.Limits.RateLimiter.Capacity > 0
	refillRateEnabled := cfg.App.Limits.RateLimiter.RefillRate > 0
	if capacityEnabled != refillRateEnabled {
		return errs.ErrConfigInvalidRateLimiter
	}

	if cfg.Log.Type != LogTypeJSON && cfg.Log.Type != LogTypeText {
		return errs.ErrConfigWrongLogType
	}

	if cfg.Log.Level != LogLevelDebug && cfg.Log.Level != LogLevelInfo && cfg.Log.Level != LogLevelWarn && cfg.Log.Level != LogLevelError {
		return errs.ErrConfigWrongLogLevel
	}

	if _, ok := imgproc.SupportedOutputFormat(cfg.Image.Ext); !ok {
		return errs.ErrConfigInvalidImageFormat
	}

	if cfg.Image.MaxDimension < 1000 || cfg.Image.MaxDimension > 10000 {
		return errs.ErrConfigImageDimensionOutOfRange
	}

	if cfg.App.Security.ReadToken == "" {
		return fmt.Errorf("read token not set : %w", errs.ErrTokenNotSet)
	}

	if cfg.App.Security.WriteToken == "" {
		return fmt.Errorf("write token not set : %w", errs.ErrTokenNotSet)
	}

	if cfg.App.Storage != StorageFileSystem && cfg.App.Storage != StorageInmemory {
		return errs.ErrConfigInvalidStorage
	}

	if cfg.App.Storage == StorageFileSystem {
		if cfg.Storage.FileSystem.Path == "" {
			return fmt.Errorf("%w: filesystem.path is required", errs.ErrConfigInvalidStorage)
		}
	}

	return nil
}
