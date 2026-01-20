package config

import (
	"errors"
	"file-storage/internal/errs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

func TestNewConfig(t *testing.T) {
	tmpdir := t.TempDir()
	cfg := defaults()
	cfg.App.Port = 999
	cfg.Log.Level = LogLevelWarn
	cfg.Log.Type = LogTypeText
	cfg.Image.MaxDimention = 2000
	cfg.App.Security.ReadToken = "22"

	path := filepath.Join(tmpdir, "config.yaml")
	bytes, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		t.Fatalf("write file %s error: %s", path, err)
	}

	err = os.Setenv("FILE_STORAGE_LOG_LEVEL", LogLevelError)
	if err != nil {
		t.Fatalf("set FILE_STORAGE_LOG_LEVEL error: %s", err)
	}
	defer os.Unsetenv("FILE_STORAGE_LOG_LEVEL")

	err = os.Setenv("FILE_STORAGE_WRITE_TOKEN", "222")
	if err != nil {
		t.Fatalf("set FILE_STORAGE_WRITE_TOKEN error: %s", err)
	}
	defer os.Unsetenv("FILE_STORAGE_WRITE_TOKEN")

	pflag.CommandLine = pflag.NewFlagSet("test_nc", pflag.ExitOnError)
	pflag.String("loglevel", "info", "log level")
	pflag.String("logtype", "json", "log type")
	pflag.Int("port", 0, "application port")
	pflag.String("readtoken", "123", "read token")
	pflag.String("writetoken", "123", "write token")
	pflag.String("imageext", "", "stored image format")

	pflag.Set("port", "666")
	pflag.Set("imageext", "jpeg")
	pflag.Parse()

	config, err := NewConfig(path)
	if err != nil {
		t.Fatalf("new config error: %s", err)
	}

	if config.Log.Level != LogLevelError {
		t.Errorf("expected log level %s got %s", LogLevelError, config.Log.Level)
	}
	if config.Log.Type != LogTypeText {
		t.Errorf("expected log type %s got %s", LogTypeText, config.Log.Type)
	}
	if config.App.Port != 666 {
		t.Errorf("expected app port 666 got %d", config.App.Port)
	}
}

func TestDefaults(t *testing.T) {
	cfg := defaults()

	if cfg.Log.Level != LogLevelInfo {
		t.Errorf("expected log level %s got %s", LogLevelInfo, cfg.Log.Level)
	}
	if cfg.Log.Type != LogTypeJSON {
		t.Errorf("expected log type %s got %s", LogTypeJSON, cfg.Log.Type)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("expected app port 0 got %d", cfg.App.Port)
	}
}

func TestApplyConfigFile(t *testing.T) {
	tmpdir := t.TempDir()
	cfg := defaults()

	path := filepath.Join(tmpdir, "config.yaml")
	bytes, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		t.Fatalf("write file %s error: %s", path, err)
	}

	cfgt := defaults()
	err = applyConfigFile(&cfgt, path)
	if err != nil {
		t.Fatalf("expected ok got err: %s", err)
	}
	if !reflect.DeepEqual(cfg, cfgt) {
		t.Errorf("expected %#v got %#v", cfg, cfgt)
	}

	path2 := filepath.Join(tmpdir, "config2.yaml")
	err = applyConfigFile(&cfgt, path2)
	if err == nil {
		t.Errorf("expected path error got ok")
	}

	bytes = []byte("текст")
	path3 := filepath.Join(tmpdir, "config3.yaml")
	err = os.WriteFile(path3, bytes, 0644)
	if err != nil {
		t.Fatalf("write file %s error: %s", path3, err)
	}
	err = applyConfigFile(&cfgt, path3)
	if err == nil {
		t.Errorf("expected unmarshal error got ok")
	}
}

func TestApplyEnv(t *testing.T) {
	cfg := Config{App: App{}, Log: Log{}}

	err := os.Setenv("FILE_STORAGE_APP_PORT", "5")
	if err != nil {
		t.Fatalf("set FILE_STORAGE_APP_PORT error: %s", err)
	}
	defer os.Unsetenv("FILE_STORAGE_APP_PORT")

	err = os.Setenv("FILE_STORAGE_LOG_LEVEL", LogLevelWarn)
	if err != nil {
		t.Fatalf("set FILE_STORAGE_LOG_LEVEL error: %s", err)
	}
	defer os.Unsetenv("FILE_STORAGE_LOG_LEVEL")

	err = os.Setenv("FILE_STORAGE_LOG_TYPE", LogTypeText)
	if err != nil {
		t.Fatalf("set FILE_STORAGE_LOG_TYPE error: %s", err)
	}
	defer os.Unsetenv("FILE_STORAGE_LOG_TYPE")

	err = os.Setenv("FILE_STORAGE_TIMEOUT", "5s")
	if err != nil {
		t.Fatalf("set FILE_STORAGE_TIMEOUT error: %s", err)
	}
	defer os.Unsetenv("FILE_STORAGE_TIMEOUT")

	err = applyEnv(&cfg)
	if err != nil {
		t.Fatalf("applyEnv error: %s", err)
	}

	if cfg.App.Port != 5 {
		t.Errorf("expect port 5 got %d", cfg.App.Port)
	}
	if cfg.App.Timeout != 5*time.Second {
		t.Errorf("expect timeout %v got %v", 5*time.Second, cfg.App.Timeout)
	}
	if cfg.Log.Level != LogLevelWarn {
		t.Errorf("expect log level 'warn' got %s", cfg.Log.Level)
	}
	if cfg.Log.Type != LogTypeText {
		t.Errorf("expect log type 'text'  got %s", cfg.Log.Type)
	}
}

func TestApplyFlags(t *testing.T) {
	cfg := Config{App: App{}, Log: Log{}}

	pflag.CommandLine = pflag.NewFlagSet("test_af", pflag.ExitOnError)
	pflag.String("loglevel", "info", "log level")
	pflag.String("logtype", "json", "log type")
	pflag.Int("port", 0, "application port")
	pflag.Int("sizelimit", 0, "max file size")

	pflag.Set("config", "C:/config.txt")
	pflag.Set("loglevel", LogLevelWarn)
	pflag.Set("logtype", LogTypeText)
	pflag.Set("port", "5")
	pflag.Set("sizelimit", "1000")

	err := applyFlags(&cfg)
	if err != errs.ErrConfigFlagsNotParsed {
		t.Fatalf("expect %s got %v", errs.ErrConfigFlagsNotParsed, err)
	}

	pflag.Parse()

	err = applyFlags(&cfg)
	if err != nil {
		t.Fatalf("applyFlags error: %s", err)
	}

	if cfg.App.Port != 5 {
		t.Errorf("expect port 5 got %d", cfg.App.Port)
	}
	if cfg.App.SizeLimit != 1000 {
		t.Errorf("expect port 1000 got %d", cfg.App.SizeLimit)
	}
	if cfg.Log.Level != LogLevelWarn {
		t.Errorf("expect log level 'warn' got %s", cfg.Log.Level)
	}
	if cfg.Log.Type != LogTypeText {
		t.Errorf("expect log type 'text'  got %s", cfg.Log.Type)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want error
	}{
		{
			name: "host not set",
			cfg: Config{
				App: App{
					Timeout:  5 * time.Second,
					Security: Security{ReadToken: "1", WriteToken: "2"}},
				Log:   Log{Level: LogLevelDebug, Type: LogTypeJSON},
				Image: Image{Ext: "jpeg", MaxDimention: 2000},
			},
			want: errs.ErrConfigHostNotSet,
		},
		{
			name: "invalid timeout",
			cfg: Config{
				App: App{
					Timeout:  0 * time.Second,
					Port:     1,
					Host:     "127.0.0.1",
					Security: Security{ReadToken: "1", WriteToken: "2"}},
				Log:   Log{Level: LogLevelDebug, Type: LogTypeJSON},
				Image: Image{Ext: "jpeg", MaxDimention: 2000},
			},
			want: errs.ErrConfigInvalidTimeout,
		},
		{
			name: "app port < 0",
			cfg: Config{
				App: App{
					Timeout:  5 * time.Second,
					Host:     "127.0.0.1",
					Port:     -2,
					Security: Security{ReadToken: "1", WriteToken: "2"}},
				Log:   Log{Level: LogLevelDebug, Type: LogTypeJSON},
				Image: Image{Ext: "jpeg", MaxDimention: 2000},
			},
			want: errs.ErrConfigPortOutOfRange,
		},
		{
			name: "app port > 65535",
			cfg: Config{
				App: App{
					Timeout:  5 * time.Second,
					Host:     "127.0.0.1",
					Port:     65536,
					Security: Security{ReadToken: "1", WriteToken: "2"}},
				Log:   Log{Level: LogLevelDebug, Type: LogTypeJSON},
				Image: Image{Ext: "jpeg", MaxDimention: 2000},
			},
			want: errs.ErrConfigPortOutOfRange,
		},
		{
			name: "log level incorrect",
			cfg: Config{
				App: App{
					Timeout:  5 * time.Second,
					Host:     "127.0.0.1",
					Port:     2,
					Security: Security{ReadToken: "1", WriteToken: "2"}},
				Log:   Log{Level: "asd", Type: LogTypeJSON},
				Image: Image{Ext: "jpeg", MaxDimention: 2000},
			},
			want: errs.ErrConfigWrongLogLevel,
		},
		{
			name: "log type incorrect",
			cfg: Config{
				App: App{
					Timeout:  5 * time.Second,
					Host:     "127.0.0.1",
					Port:     2,
					Security: Security{ReadToken: "1", WriteToken: "2"}},
				Log:   Log{Level: LogLevelDebug, Type: "jjson"},
				Image: Image{Ext: "jpeg", MaxDimention: 2000},
			},
			want: errs.ErrConfigWrongLogType,
		},
		{
			name: "invalid image format",
			cfg: Config{
				App: App{
					Timeout:  5 * time.Second,
					Host:     "127.0.0.1",
					Port:     2,
					Security: Security{ReadToken: "1", WriteToken: "2"}},
				Log:   Log{Level: LogLevelDebug, Type: LogTypeJSON},
				Image: Image{Ext: "jpegd", MaxDimention: 2000},
			},
			want: errs.ErrConfigInvalidImageFormat,
		},
		{
			name: "token not set",
			cfg: Config{
				App: App{
					Timeout:  5 * time.Second,
					Host:     "127.0.0.1",
					Port:     2,
					Security: Security{ReadToken: "", WriteToken: ""}},
				Log:   Log{Level: LogLevelDebug, Type: LogTypeJSON},
				Image: Image{Ext: "jpeg", MaxDimention: 2000},
			},
			want: errs.ErrTokenNotSet,
		},
		{
			name: "ok",
			cfg: Config{
				App: App{
					Timeout:  5 * time.Second,
					Host:     "127.0.0.1",
					Port:     2,
					Security: Security{ReadToken: "1", WriteToken: "2"}},
				Log:   Log{Level: LogLevelDebug, Type: LogTypeJSON},
				Image: Image{Ext: "jpeg", MaxDimention: 2000},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		if err := validate(&tt.cfg); !errors.Is(err, tt.want) {
			t.Errorf("%s: expected %v got %v", tt.name, tt.want, err)
		}

	}
}
