package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/gliderlabs/sigil"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Main  Main  `toml:"main" validate:"required"`
	Slack Slack `toml:"slack" validate:"required"`
	Toggl Toggl `toml:"toggl" validate:"required"`
}

type Main struct {
	LogLevel string `toml:"log_level" validate:"required,oneof=debug info warn error"`
}

type Slack struct {
	Token     string `toml:"token" validate:"required"`
	ChannelID string `toml:"channel_id" validate:"required"`
}

type Toggl struct {
	APIToken    string `toml:"api_token" validate:"required"`
	WorkspaceID int64  `toml:"workspace_id" validate:"required"`
	ProjectID   int64  `toml:"project_id"`
}

func New(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}

	// ${VAR:-default} のようなPOSIX形式のデフォルト値構文を有効にする
	sigil.PosixPreprocess = true
	tmpl, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}

	// sigil がテンプレート内で別ファイルをincludeする際の基準ディレクトリを設定する
	// 今回はincludeを使わないが、sigil.Execute が内部で参照するため必要
	sigil.PushPath(filepath.Dir(path))
	render, err := sigil.Execute(tmpl, map[string]interface{}{}, filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("sigil.Execute: %w", err)
	}

	cfg := &Config{}
	if _, err := toml.Decode(render.String(), cfg); err != nil {
		return nil, fmt.Errorf("toml.Decode: %w", err)
	}

	v := validator.New()
	if err := v.Struct(cfg); err != nil {
		return nil, fmt.Errorf("validator.Struct: %w", err)
	}

	return cfg, nil
}
