package config

import (
	"file-storage/internal/errs"
	"file-storage/internal/imgproc"
	"fmt"
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

type App struct {
	Host      string        `json:"host" yaml:"host"`
	Port      int           `json:"port" yaml:"port"`
	Timeout   time.Duration `json:"timeout" yaml:"timeout"`
	SizeLimit int           `json:"size_limit" yaml:"size_limit"`
	Security  Security      `json:"security" yaml:"security"`
}

type Log struct {
	Level string `json:"level" yaml:"level"`
	Type  string `json:"type" yaml:"type"`
}

type Security struct {
	ReadToken  string `json:"read_token" yaml:"read_token"`
	WriteToken string `json:"write_token" yaml:"write_token"`
}

type Image struct {
	Ext          string `json:"ext" yaml:"ext"`
	MaxDimention int    `json:"width" yaml:"max_dimention"`
}

type Config struct {
	App   App   `json:"app" yaml:"app"`
	Log   Log   `json:"log" yaml:"log"`
	Image Image `json:"image" yaml:"image"`
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
			Host:      "127.0.0.1",
			Port:      8080,
			SizeLimit: defaultSizeLimit,
			Timeout:   5 * time.Second,
			Security: Security{
				ReadToken:  "default token",
				WriteToken: "default token",
			},
		},
		Image: Image{
			Ext:          "jpeg",
			MaxDimention: 2000},
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

	sAppPort := os.Getenv("FILE_STORAGE_APP_PORT")
	if sAppPort != "" {
		port, err := strconv.Atoi(sAppPort)
		if err != nil {
			return err
		}
		cfg.App.Port = port
	}

	sAppHost := os.Getenv("FILE_STORAGE_APP_HOST")
	if sAppHost != "" {
		cfg.App.Host = sAppHost
	}

	sSizeLimit := os.Getenv("FILE_STORAGE_SIZE_LIMIT")
	if sSizeLimit != "" {
		sizeLimit, err := strconv.Atoi(sSizeLimit)
		if err != nil {
			return err
		}
		cfg.App.SizeLimit = sizeLimit
	}

	sTimeout := os.Getenv("FILE_STORAGE_TIMEOUT")
	if sTimeout != "" {
		timeout, err := time.ParseDuration(sTimeout)
		if err != nil {
			return err
		}
		cfg.App.Timeout = timeout
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

	sImageMaxDimention := os.Getenv("FILE_STORAGE_IMAGE_MAX_DIMENTION")
	if sImageMaxDimention != "" {
		maxDim, err := strconv.Atoi(sImageMaxDimention)
		if err != nil {
			return err
		}
		cfg.Image.MaxDimention = maxDim
	}

	return nil
}

func applyFlags(cfg *Config) error {

	if !pflag.Parsed() {
		return errs.ErrConfigFlagsNotParsed
	}

	fAppPort := pflag.Lookup("port")
	if fAppPort != nil && fAppPort.Changed {
		raw := fAppPort.Value.String()
		port, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		cfg.App.Port = port
	}

	fHost := pflag.Lookup("readtoken")
	if fHost != nil && fHost.Changed {
		cfg.App.Host = fHost.Value.String()
	}

	fSizeLimit := pflag.Lookup("sizelimit")
	if fSizeLimit != nil && fSizeLimit.Changed {
		raw := fSizeLimit.Value.String()
		sizeLimit, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		cfg.App.SizeLimit = sizeLimit
	}

	fTimeout := pflag.Lookup("timeout")
	if fTimeout != nil && fTimeout.Changed {
		raw := fTimeout.Value.String()
		timeout, err := time.ParseDuration(raw)
		if err != nil {
			return err
		}
		cfg.App.Timeout = timeout
	}

	fReadToken := pflag.Lookup("readtoken")
	if fReadToken != nil && fReadToken.Changed {
		cfg.App.Security.ReadToken = fReadToken.Value.String()
	}

	fWriteToken := pflag.Lookup("writetoken")
	if fWriteToken != nil && fWriteToken.Changed {
		cfg.App.Security.WriteToken = fWriteToken.Value.String()
	}

	fLogLevel := pflag.Lookup("loglevel")
	if fLogLevel != nil && fLogLevel.Changed {
		cfg.Log.Level = fLogLevel.Value.String()
	}

	fLogType := pflag.Lookup("logtype")
	if fLogType != nil && fLogType.Changed {
		cfg.Log.Type = fLogType.Value.String()
	}

	fImageExt := pflag.Lookup("imageext")
	if fImageExt != nil && fImageExt.Changed {
		cfg.Image.Ext = fImageExt.Value.String()
	}

	fImageMaxDimention := pflag.Lookup("imagemaxdimention")
	if fImageMaxDimention != nil && fImageMaxDimention.Changed {
		raw := fImageMaxDimention.Value.String()
		maxDim, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		cfg.Image.MaxDimention = maxDim
	}

	return nil
}

func normalize(cfg *Config) {
	cfg.Log.Type = strings.ToLower(cfg.Log.Type)
	cfg.Log.Level = strings.ToLower(cfg.Log.Level)
}

func validate(cfg *Config) error {
	if cfg.App.Host == "" {
		return errs.ErrConfigHostNotSet
	}

	if cfg.App.Port < 1 || cfg.App.Port > 65535 {
		return errs.ErrConfigPortOutOfRange
	}

	if cfg.App.Timeout <= 0 {
		return errs.ErrConfigInvalidTimeout
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

	if cfg.Image.MaxDimention < 1000 || cfg.Image.MaxDimention > 10000 {
		return errs.ErrConfigImageDimentionOutOfRange
	}

	if cfg.App.Security.ReadToken == "" {
		return fmt.Errorf("read token not set : %w", errs.ErrTokenNotSet)
	}

	if cfg.App.Security.WriteToken == "" {
		return fmt.Errorf("write token not set : %w", errs.ErrTokenNotSet)
	}

	return nil
}
