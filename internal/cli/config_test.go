package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/pelletier/go-toml/v2"
)

// TestConfigCommandShowsEffectiveSettings verifies complete output after loading defaults and overrides.
func TestConfigCommandShowsEffectiveSettings(t *testing.T) {
	for _, overridden := range []bool{false, true} {
		name := "defaults"
		if overridden {
			name = "overrides and normalization"
		}
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			want := config.Default(home)
			if overridden {
				data := []byte("logRetentionDays = 0\nlogTimezone = 'UTC'\nhistoryDir = ''\n")
				if err := os.MkdirAll(filepath.Dir(want.Path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(want.Path, data, 0o644); err != nil {
					t.Fatal(err)
				}
				want.LogTimezone = "UTC"
			}
			cfg, err := config.Load()
			if err != nil {
				t.Fatal(err)
			}
			output := &stubOutput{}
			if err := runCommand("config", stubLogger{}, output, cfg); err != nil {
				t.Fatal(err)
			}
			var got config.Config
			if err := toml.Unmarshal(output.buffer.Bytes(), &got); err != nil {
				t.Fatalf("decode config output: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("effective config = %+v, want %+v", got, want)
			}
			var settings map[string]any
			if err := toml.Unmarshal(output.buffer.Bytes(), &settings); err != nil {
				t.Fatal(err)
			}
			// Check keys separately so omitted zero values cannot pass.
			schema := reflect.TypeOf(cfg)
			for i := 0; i < schema.NumField(); i++ {
				key := schema.Field(i).Tag.Get("toml")
				if _, ok := settings[key]; !ok {
					t.Errorf("config output missing parameter %q", key)
				}
			}
		})
	}
}
