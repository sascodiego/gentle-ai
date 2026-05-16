package pi

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
	"github.com/gentleman-programming/gentle-ai/internal/versions"
)

func TestAdapterIdentityAndCapabilities(t *testing.T) {
	a := NewAdapter()

	if got := a.Agent(); got != model.AgentPi {
		t.Fatalf("Agent() = %q, want %q", got, model.AgentPi)
	}
	if got := a.Tier(); got != model.TierFull {
		t.Fatalf("Tier() = %q, want %q", got, model.TierFull)
	}

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{"SupportsAutoInstall", a.SupportsAutoInstall(), true},
		{"SupportsSkills", a.SupportsSkills(), false},
		{"SupportsMCP", a.SupportsMCP(), true},
		{"SupportsSystemPrompt", a.SupportsSystemPrompt(), false},
		{"SupportsSlashCommands", a.SupportsSlashCommands(), false},
		{"SupportsOutputStyles", a.SupportsOutputStyles(), false},
		{"SupportsSubAgents", a.SupportsSubAgents(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestAdapterPaths(t *testing.T) {
	a := NewAdapter()
	homeDir := t.TempDir()
	piDir := filepath.Join(homeDir, ".pi")
	piAgentDir := filepath.Join(piDir, "agent")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"GlobalConfigDir", a.GlobalConfigDir(homeDir), piDir},
		{"SystemPromptDir", a.SystemPromptDir(homeDir), ""},
		{"SystemPromptFile", a.SystemPromptFile(homeDir), ""},
		{"SkillsDir", a.SkillsDir(homeDir), ""},
		{"SettingsPath", a.SettingsPath(homeDir), filepath.Join(piAgentDir, "settings.json")},
		{"CommandsDir", a.CommandsDir(homeDir), ""},
		{"MCPConfigPath", a.MCPConfigPath(homeDir, "context7"), filepath.Join(piAgentDir, "mcp.json")},
		{"OutputStyleDir", a.OutputStyleDir(homeDir), ""},
		{"SubAgentsDir", a.SubAgentsDir(homeDir), ""},
		{"EmbeddedSubAgentsDir", a.EmbeddedSubAgentsDir(), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestAdapterDetectUsesPiBinaryAndConfigPath(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".pi")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	a := &Adapter{
		lookPath: func(file string) (string, error) {
			if file != "pi" {
				t.Fatalf("lookPath called with %q, want pi", file)
			}
			return "/usr/local/bin/pi", nil
		},
		statPath: defaultStat,
	}

	installed, binaryPath, configPath, configFound, err := a.Detect(context.Background(), homeDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if !installed {
		t.Fatalf("Detect() installed = false, want true")
	}
	if binaryPath != "/usr/local/bin/pi" {
		t.Fatalf("Detect() binaryPath = %q, want /usr/local/bin/pi", binaryPath)
	}
	if configPath != configDir {
		t.Fatalf("Detect() configPath = %q, want %q", configPath, configDir)
	}
	if !configFound {
		t.Fatalf("Detect() configFound = false, want true")
	}
}

func TestAdapterDetectMissingPiBinary(t *testing.T) {
	homeDir := t.TempDir()
	a := &Adapter{
		lookPath: func(file string) (string, error) {
			return "", os.ErrNotExist
		},
		statPath: defaultStat,
	}

	installed, binaryPath, configPath, configFound, err := a.Detect(context.Background(), homeDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if installed {
		t.Fatalf("Detect() installed = true, want false")
	}
	if binaryPath != "" {
		t.Fatalf("Detect() binaryPath = %q, want empty", binaryPath)
	}
	if configPath != filepath.Join(homeDir, ".pi") {
		t.Fatalf("Detect() configPath = %q, want ~/.pi under home", configPath)
	}
	if configFound {
		t.Fatalf("Detect() configFound = true, want false")
	}
}

func TestInstallCommand(t *testing.T) {
	pnpmDlxSuffix := []string{"--package", "gentle-engram@" + versions.GentleEngram, "--", "pi-engram", "init"}

	// Bare pi commands with npm: prefix — used when no sudo needed.
	barePICommands := func() [][]string {
		return [][]string{
			{"pi", "install", "npm:gentle-pi"},
			{"pi", "install", "npm:gentle-engram"},
			{"pi", "install", "npm:pi-mcp-adapter"},
			{"pnpm", "dlx"},
			{"pi", "install", "npm:pi-subagents"},
			{"pi", "install", "npm:pi-intercom"},
			{"pi", "install", "npm:@juicesharp/rpiv-ask-user-question"},
			{"pi", "install", "npm:pi-web-access"},
			{"pi", "install", "npm:pi-lens"},
			{"pi", "install", "npm:@juicesharp/rpiv-todo"},
			{"pi", "install", "npm:pi-btw"},
		}
	}

	tests := []struct {
		name          string
		profile       system.PlatformProfile
		lookPath      func(string) (string, error)
		wantErr       bool
		wantCommands  [][]string
	}{
		{
			name:    "linux non-writable uses sudo with full path and npm prefix",
			profile: system.PlatformProfile{OS: "linux", PnpmWritable: false},
			lookPath: func(file string) (string, error) {
				if file == "pi" {
					return "/opt/nodejs/bin/pi", nil
				}
				return "", os.ErrNotExist
			},
			wantErr: false,
			wantCommands: func() [][]string {
				return [][]string{
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:gentle-pi"},
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:gentle-engram"},
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:pi-mcp-adapter"},
					append([]string{"sudo", "pnpm", "dlx"}, pnpmDlxSuffix...),
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:pi-subagents"},
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:pi-intercom"},
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:@juicesharp/rpiv-ask-user-question"},
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:pi-web-access"},
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:pi-lens"},
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:@juicesharp/rpiv-todo"},
					{"sudo", "/opt/nodejs/bin/pi", "install", "npm:pi-btw"},
				}
			}(),
		},
		{
			name:    "linux writable uses bare pi with npm prefix",
			profile: system.PlatformProfile{OS: "linux", PnpmWritable: true},
			lookPath: func(file string) (string, error) {
				return "/usr/local/bin/pi", nil
			},
			wantErr: false,
			wantCommands: func() [][]string {
				cmds := barePICommands()
				cmds[3] = append([]string{"pnpm", "dlx"}, pnpmDlxSuffix...)
				return cmds
			}(),
		},
		{
			name:    "darwin uses bare pi with npm prefix",
			profile: system.PlatformProfile{OS: "darwin"},
			lookPath: func(file string) (string, error) {
				return "/usr/local/bin/pi", nil
			},
			wantErr: false,
			wantCommands: func() [][]string {
				cmds := barePICommands()
				cmds[3] = append([]string{"pnpm", "dlx"}, pnpmDlxSuffix...)
				return cmds
			}(),
		},
		{
			name:    "linux non-writable lookPath failure returns error",
			profile: system.PlatformProfile{OS: "linux", PnpmWritable: false},
			lookPath: func(file string) (string, error) {
				return "", os.ErrNotExist
			},
			wantErr:      true,
			wantCommands: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				lookPath: tt.lookPath,
				statPath: defaultStat,
			}
			commands, err := a.InstallCommand(tt.profile)
			if (err != nil) != tt.wantErr {
				t.Fatalf("InstallCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(commands, tt.wantCommands) {
				t.Fatalf("InstallCommand(%+v)\ngot  = %v\nwant = %v", tt.profile, commands, tt.wantCommands)
			}
		})
	}
}
