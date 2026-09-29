package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetSettingsEmpty(t *testing.T) {
	withTempConfig(t)
	s := &SettingsService{}
	got, err := s.GetSettings()
	if err != nil || got != (Settings{}) {
		t.Fatalf("空配置应返回零值: %+v %v", got, err)
	}
}

func TestSaveSettingsMkdirAll(t *testing.T) {
	root := withTempConfig(t)
	s := &SettingsService{}
	nested := filepath.Join(root, "a", "b", "c")
	saved, err := s.SaveSettings(Settings{DefaultPath: nested})
	if err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}
	if info, err := os.Stat(nested); err != nil || !info.IsDir() {
		t.Fatalf("DefaultPath 应被 MkdirAll: %v", err)
	}
	if !filepath.IsAbs(saved.DefaultPath) {
		t.Fatalf("DefaultPath 应存绝对路径: %s", saved.DefaultPath)
	}
	// 回读一致
	got, _ := s.GetSettings()
	if got.DefaultPath != nested {
		t.Fatalf("回读不一致: %+v", got)
	}
}

func TestScanEditorsDefaultFunc(t *testing.T) {
	// 默认扫描器在任意机器上应能跑通(空结果也算通过),不依赖真实环境
	old := userConfigDir
	userConfigDir = func() (string, error) { return t.TempDir(), nil }
	t.Cleanup(func() { userConfigDir = old })
	if _, err := scanEditorsFunc(); err != nil {
		t.Fatalf("系统扫描不应报错: %v", err)
	}
}

// bootstrap 用测试辅助:注入扫描器与主目录。
type bootstrapEnv struct {
	scanned int
}

func TestBootstrapFirstRun(t *testing.T) {
	root := withTempConfig(t)
	home := withTempHome(t)
	be := &bootstrapEnv{}
	oldScan := scanEditorsFunc
	scanEditorsFunc = func() ([]Editor, error) {
		be.scanned++
		return []Editor{
			{ID: "zed", Name: "Zed", Command: "zed", Kind: KindEditor, Source: "scan", Enabled: true},
			{ID: "code", Name: "VS Code", Command: "code", Kind: KindEditor, Source: "scan", Enabled: true},
			{ID: "alacritty", Name: "Alacritty", Command: "alacritty", Kind: KindTerminal, Source: "scan", Enabled: true},
			{ID: "konsole", Name: "Konsole", Command: "konsole", Kind: KindTerminal, Source: "scan", Enabled: true},
		}, nil
	}
	t.Cleanup(func() { scanEditorsFunc = oldScan })

	if err := bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if be.scanned != 1 {
		t.Fatalf("首次启动应扫描一次,实际 %d", be.scanned)
	}
	// 三个 json 均已落盘
	for _, f := range []string{"projects.json", "editors.json", "settings.json"} {
		if !fileExists(filepath.Join(root, f)) {
			t.Fatalf("缺 %s", f)
		}
	}
	// 默认路径 = $HOME/Projects 且已创建
	st, _ := loadSettings()
	if st.DefaultPath != filepath.Join(home, "Projects") {
		t.Fatalf("DefaultPath 应为 $HOME/Projects: %s", st.DefaultPath)
	}
	if info, err := os.Stat(st.DefaultPath); err != nil || !info.IsDir() {
		t.Fatalf("默认目录应已创建: %v", err)
	}
	// 默认编辑器优先 code(VS Code 的命令名),终端优先 konsole
	if st.DefaultEditorID != "code" {
		t.Fatalf("默认编辑器应为 code: %s", st.DefaultEditorID)
	}
	if st.DefaultTerminalID != "konsole" {
		t.Fatalf("默认终端应为 konsole: %s", st.DefaultTerminalID)
	}
}

func TestBootstrapFallbackDefaults(t *testing.T) {
	withTempConfig(t)
	withTempHome(t)
	oldScan := scanEditorsFunc
	scanEditorsFunc = func() ([]Editor, error) {
		return []Editor{
			{ID: "zed", Name: "Zed", Command: "zed", Kind: KindEditor, Source: "scan", Enabled: true},
			{ID: "alacritty", Name: "Alacritty", Command: "alacritty", Kind: KindTerminal, Source: "scan", Enabled: true},
			{ID: "code", Name: "VS Code", Command: "code", Kind: KindEditor, Source: "scan", Enabled: false}, // 禁用不参与
		}, nil
	}
	t.Cleanup(func() { scanEditorsFunc = oldScan })

	if err := bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	st, _ := loadSettings()
	if st.DefaultEditorID != "zed" {
		t.Fatalf("无 vscode 时应取第一个 enabled 编辑器: %s", st.DefaultEditorID)
	}
	if st.DefaultTerminalID != "alacritty" {
		t.Fatalf("无 konsole 时应取第一个 enabled 终端: %s", st.DefaultTerminalID)
	}
}

func TestBootstrapSecondRunKeepsManual(t *testing.T) {
	root := withTempConfig(t)
	withTempHome(t)
	oldScan := scanEditorsFunc
	scanEditorsFunc = func() ([]Editor, error) {
		return []Editor{{ID: "vscode", Name: "VS Code", Command: "code", Kind: KindEditor, Source: "scan", Enabled: true}}, nil
	}
	t.Cleanup(func() { scanEditorsFunc = oldScan })

	if err := bootstrap(); err != nil {
		t.Fatalf("第一次: %v", err)
	}
	// 用户改动:清空默认编辑器 + 删掉扫描项
	if err := saveSettings(Settings{DefaultPath: root, DefaultEditorID: "m-manual", DefaultTerminalID: ""}); err != nil {
		t.Fatalf("saveSettings: %v", err)
	}
	if err := saveEditors([]Editor{}); err != nil {
		t.Fatalf("saveEditors: %v", err)
	}
	if err := bootstrap(); err != nil {
		t.Fatalf("第二次: %v", err)
	}
	st, _ := loadSettings()
	if st.DefaultEditorID != "m-manual" {
		t.Fatalf("已有编辑器列表时不应改默认编辑器: %+v", st)
	}
	// 自动合并语义:扫描到的 scan 项会被合并回来,但用户设置的默认项不受影响
	es, _ := loadEditors()
	if len(es) != 1 || es[0].ID != "vscode" {
		t.Fatalf("二次启动应自动合并扫描项(vscode 回来): %+v", es)
	}
}

// 每次启动自动合并扫描:新装工具自动出现、卸载的自动消失,用户顺序/开关/手动项保留。
func TestBootstrapAutoMergesOnStartup(t *testing.T) {
	withTempConfig(t)
	withTempHome(t)
	old := scanEditorsFunc
	scanEditorsFunc = func() ([]Editor, error) {
		return []Editor{{ID: "zed", Name: "Zed", Command: "zed", Kind: KindEditor, Source: "scan", Enabled: true}}, nil
	}
	t.Cleanup(func() { scanEditorsFunc = old })

	if err := bootstrap(); err != nil {
		t.Fatalf("第一次: %v", err)
	}
	// 用户改动:禁用 zed,手动加了一个工具
	if err := saveEditors([]Editor{
		{ID: "zed", Name: "Zed", Command: "zed", Kind: KindEditor, Source: "scan", Enabled: false},
		{ID: "m-abcd1234", Name: "手动", Command: "myed", Kind: KindEditor, Source: "manual", Enabled: true},
	}); err != nil {
		t.Fatalf("saveEditors: %v", err)
	}
	// 第二次启动:扫描器发现了新装的 codex
	scanEditorsFunc = func() ([]Editor, error) {
		return []Editor{
			{ID: "zed", Name: "Zed", Command: "zed", Kind: KindEditor, Source: "scan", Enabled: true},
			{ID: "codex", Name: "Codex CLI", Command: "codex", Kind: KindCLI, Source: "scan", Enabled: true},
		}, nil
	}
	if err := bootstrap(); err != nil {
		t.Fatalf("第二次: %v", err)
	}
	es, _ := loadEditors()
	if len(es) != 3 {
		t.Fatalf("应为 zed+codex+手动 3 项: %+v", es)
	}
	if es[0].ID != "zed" || es[0].Enabled {
		t.Fatalf("zed 应保持原顺序且保持禁用: %+v", es[0])
	}
	if es[1].ID != "codex" || es[1].Kind != KindCLI {
		t.Fatalf("新装的 codex 应追加在已有项之后: %+v", es[1])
	}
	if es[2].ID != "m-abcd1234" {
		t.Fatalf("手动项应保持在末尾: %+v", es[2])
	}
}

// 用户显式清除默认项(SetDefault*("")),重启后不得被自动挑选回填。
func TestBootstrapKeepsClearedDefaults(t *testing.T) {
	withTempConfig(t)
	withTempHome(t)
	oldScan := scanEditorsFunc
	scanEditorsFunc = func() ([]Editor, error) {
		return []Editor{
			{ID: "vscode", Name: "VS Code", Command: "code", Kind: KindEditor, Source: "scan", Enabled: true},
			{ID: "konsole", Name: "Konsole", Command: "konsole", Kind: KindTerminal, Source: "scan", Enabled: true},
		}, nil
	}
	t.Cleanup(func() { scanEditorsFunc = oldScan })

	if err := bootstrap(); err != nil {
		t.Fatalf("第一次: %v", err)
	}
	if err := saveSettings(Settings{DefaultPath: "/tmp", DefaultEditorID: "", DefaultTerminalID: ""}); err != nil {
		t.Fatalf("saveSettings: %v", err)
	}
	if err := bootstrap(); err != nil {
		t.Fatalf("第二次: %v", err)
	}
	st, _ := loadSettings()
	if st.DefaultEditorID != "" || st.DefaultTerminalID != "" {
		t.Fatalf("显式清除后重启不应回填: %+v", st)
	}
}

func TestPickEditorPreference(t *testing.T) {
	eds := []Editor{codeEditor, nvimEditor}
	st := Settings{DefaultEditorID: "vscode"}
	// 显式 ID 优先
	ed, err := pickEditor("nvim", eds, st)
	if err != nil || ed.ID != "nvim" {
		t.Fatalf("显式 ID 应优先: %+v %v", ed, err)
	}
	// 空则用默认
	ed, err = pickEditor("", eds, st)
	if err != nil || ed.ID != "vscode" {
		t.Fatalf("应回退默认编辑器: %+v %v", ed, err)
	}
	// 全空报错
	_, err = pickEditor("", eds, Settings{})
	if err == nil || !strings.Contains(err.Error(), "未找到可用编辑器") {
		t.Fatalf("应报未找到可用编辑器: %v", err)
	}
}
