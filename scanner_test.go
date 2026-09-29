package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mkExec 在目录下创建可执行假文件。
func mkExec(t *testing.T, dir, name string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("创建 %s: %v", p, err)
	}
}

func TestScanEditorsByPath(t *testing.T) {
	bin := t.TempDir()
	mkExec(t, bin, "code")
	mkExec(t, bin, "konsole")
	mkExec(t, bin, "nvim")
	mkExec(t, bin, "hx")
	mkExec(t, bin, "notepad")  // 不在内置表
	mkExec(t, bin, "code.txt") // 不可执行
	if err := os.Mkdir(filepath.Join(bin, "vim.d"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	env := scanEnv{GOOS: "linux", PathDirs: []string{bin}}
	got, err := scanEditors(env)
	if err != nil {
		t.Fatalf("scanEditors: %v", err)
	}
	byID := map[string]Editor{}
	for _, e := range got {
		byID[e.ID] = e
	}
	if len(got) != 4 {
		t.Fatalf("应命中 4 个,得到 %d: %+v", len(got), got)
	}
	want := map[string]struct {
		editorName string
		icon       string
		kind       EditorKind
		inTerm     bool
	}{
		"code":    {"VS Code", "vscode", KindEditor, false},
		"konsole": {"Konsole", "terminal", KindTerminal, false},
		"nvim":    {"Neovim", "nvim", KindEditor, true},
		"hx":      {"Helix", "helix", KindEditor, true},
	}
	for id, w := range want {
		e, ok := byID[id]
		if !ok {
			t.Fatalf("缺 %s: %+v", id, byID)
		}
		if e.Name != w.editorName || e.Icon != w.icon || e.Kind != w.kind || e.InTerminal != w.inTerm {
			t.Fatalf("%s 不符: %+v", id, e)
		}
		if e.Command != filepath.Join(bin, id) || e.Args != "{path}" || e.Source != "scan" || !e.Enabled {
			t.Fatalf("%s 基础字段不符: %+v", id, e)
		}
	}
	if _, ok := byID["notepad"]; ok {
		t.Fatalf("内置表外不应收录")
	}
}

func TestScanEditorsDesktop(t *testing.T) {
	bin := t.TempDir()
	desk := t.TempDir()

	writeDesktop := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(desk, name), []byte(content), 0o644); err != nil {
			t.Fatalf("写 %s: %v", name, err)
		}
	}
	writeDesktop("jetbrains-idea.desktop", "[Desktop Entry]\nName=IntelliJ IDEA Ultimate\nExec=/opt/idea/bin/idea.sh %F\nType=Application\n")
	writeDesktop("com.visualstudio.code.desktop", "[Desktop Entry]\nName=Visual Studio Code\nExec=flatpak run --branch=stable --arch=x86_64 com.visualstudio.code %F\nType=Application\nNoDisplay=false\n")
	writeDesktop("hidden.desktop", "[Desktop Entry]\nName=Neovim\nExec=nvim %F\nNoDisplay=true\nType=Application\n")
	writeDesktop("linkitem.desktop", "[Desktop Entry]\nName=Not App\nExec=foo %F\nType=Link\n")
	writeDesktop("nomatch.desktop", "[Desktop Entry]\nName=Foo Bar Baz\nExec=fooarg %F\nType=Application\n")
	writeDesktop("neovim-flatpak.desktop", "[Desktop Entry]\nName=Neovim\nExec=flatpak run io.neovim.nvim %F\nType=Application\n")
	// PATH 已有 konsole,desktop 重复项应跳过
	mkExec(t, bin, "konsole")
	writeDesktop("org.kde.konsole.desktop", "[Desktop Entry]\nName=Konsole\nExec=konsole %F\nType=Application\n")

	env := scanEnv{GOOS: "linux", PathDirs: []string{bin}, DesktopDirs: []string{desk}}
	got, err := scanEditors(env)
	if err != nil {
		t.Fatalf("scanEditors: %v", err)
	}
	byID := map[string]Editor{}
	for _, e := range got {
		byID[e.ID] = e
	}

	idea, ok := byID["jetbrains-idea"]
	if !ok {
		t.Fatalf("缺 jetbrains-idea: %v", ids(got))
	}
	if idea.Icon != "jetbrains" || idea.Kind != KindEditor || idea.Command != "/opt/idea/bin/idea.sh" {
		t.Fatalf("idea 不符: %+v", idea)
	}

	code, ok := byID["com.visualstudio.code"]
	if !ok {
		t.Fatalf("缺 flatpak code: %v", ids(got))
	}
	if code.Name != "VS Code" || code.Icon != "vscode" {
		t.Fatalf("flatpak code 不符: %+v", code)
	}
	if !strings.Contains(code.Command, " ") || !strings.HasPrefix(code.Command, "flatpak run") {
		t.Fatalf("Exec 整行应含空格存入 Command: %q", code.Command)
	}

	nvim, ok := byID["neovim-flatpak"]
	if !ok {
		t.Fatalf("缺 neovim-flatpak: %v", ids(got))
	}
	if !nvim.InTerminal || nvim.Icon != "nvim" {
		t.Fatalf("neovim flatpak 不符: %+v", nvim)
	}

	for _, absent := range []string{"hidden", "linkitem", "nomatch", "org.kde.konsole"} {
		if _, ok := byID[absent]; ok {
			t.Fatalf("%s 不应收录", absent)
		}
	}
	if _, ok := byID["konsole"]; !ok {
		t.Fatalf("PATH 命中的 konsole 应保留")
	}
}

func ids(es []Editor) string {
	var parts []string
	for _, e := range es {
		parts = append(parts, e.ID)
	}
	return strings.Join(parts, ",")
}

func TestParseDesktopContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantName  string
		wantExec  string
		wantHide  bool
		wantType  string
	}{
		{
			name:     "标准条目",
			content:  "[Desktop Action x]\nName=Wrong\n\n[Desktop Entry]\nName=My App\nExec=app %U\nType=Application\nName[zh_CN]=错名\n",
			wantName: "My App",
			wantExec: "app %U",
			wantType: "Application",
		},
		{
			name:     "NoDisplay",
			content:  "[Desktop Entry]\nName=A\nExec=a\nNoDisplay=true\nType=Application\n",
			wantName: "A",
			wantExec: "a",
			wantHide: true,
			wantType: "Application",
		},
		{
			name:     "原样保留Exec",
			content:  "[Desktop Entry]\nName=A\nExec=flatpak run --branch=stable org.zed.Zed %f\nType=Application\n",
			wantName: "A",
			wantExec: "flatpak run --branch=stable org.zed.Zed %f",
			wantType: "Application",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			df := parseDesktopContent(tt.content)
			if df.Name != tt.wantName || df.Exec != tt.wantExec || df.NoDisplay != tt.wantHide || df.Type != tt.wantType {
				t.Fatalf("got %+v", df)
			}
		})
	}
}

func TestStripFieldCodes(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"code %f", "code"},
		{"code --flag %F %U", "code --flag"},
		{"flatpak run org.x.Y", "flatpak run org.x.Y"},
		{"code 100%%", "code 100%%"},
	}
	for _, tt := range tests {
		if got := stripFieldCodes(tt.in); got != tt.want {
			t.Fatalf("stripFieldCodes(%q)=%q,期望 %q", tt.in, got, tt.want)
		}
	}
}

func TestScanEditorsDarwinApps(t *testing.T) {
	apps := t.TempDir()
	for _, name := range []string{"Visual Studio Code.app", "Zed.app", "Terminal.app", "Firefox.app"} {
		if err := os.Mkdir(filepath.Join(apps, name), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	env := scanEnv{GOOS: "darwin", AppsDir: apps}
	got, err := scanEditors(env)
	if err != nil {
		t.Fatalf("scanEditors: %v", err)
	}
	byID := map[string]Editor{}
	for _, e := range got {
		byID[e.ID] = e
	}
	if vs, ok := byID["visual-studio-code"]; !ok || vs.Kind != KindEditor || vs.Icon != "vscode" {
		t.Fatalf("VS Code.app 应映射为 vscode: %+v %+v", vs, ids(got))
	}
	if term, ok := byID["terminal"]; !ok || term.Kind != KindTerminal || term.Command != "Terminal" {
		t.Fatalf("Terminal.app 应映射: %+v", term)
	}
	if _, ok := byID["firefox"]; ok {
		t.Fatalf("Firefox 不应收录")
	}
	if _, ok := byID["zed"]; !ok {
		t.Fatalf("Zed.app 应收录: %v", ids(got))
	}
}

func TestScanEditorsWindows(t *testing.T) {
	dir := t.TempDir()
	mkExec(t, dir, "code.exe")
	env := scanEnv{GOOS: "windows", PathDirs: []string{dir}, HomeDir: t.TempDir()}
	got, err := scanEditors(env)
	if err != nil {
		t.Fatalf("scanEditors: %v", err)
	}
	if len(got) != 1 || got[0].ID != "code" {
		t.Fatalf("应命中 code.exe: %+v", got)
	}
}

func TestMatchDesktopKeywords(t *testing.T) {
	tests := []struct {
		hay      string
		wantIcon string
		wantKind EditorKind
	}{
		{"Visual Studio Code flatpak run com.visualstudio.code", "vscode", KindEditor},
		{"sublime text /opt/subl", "sublime", KindEditor},
		{"PyCharm Professional /opt/pycharm", "jetbrains", KindEditor},
		{"Neovim nvim", "nvim", KindEditor},
		{"Konsole", "terminal", KindTerminal},
		{" totally unrelated ", "", ""},
	}
	for i, tt := range tests {
		m := matchDesktopKeywords(tt.hay)
		if tt.wantIcon == "" {
			if m != nil {
				t.Fatalf("case %d: 不应命中: %+v", i, m)
			}
			continue
		}
		if m == nil || m.Icon != tt.wantIcon || m.Kind != tt.wantKind {
			t.Fatalf("case %d: got %+v", i, m)
		}
	}
	// neovim 优先于 vim
	if m := matchDesktopKeywords("neovim"); m.Icon != "nvim" {
		t.Fatalf("neovim 应先于 vim 命中: %+v", m)
	}
}

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Visual Studio Code": "visual-studio-code",
		"  Zed  ":            "zed",
		"Go_Land 2":          "go-land-2",
	}
	for in, want := range tests {
		if got := slugify(in); got != want {
			t.Fatalf("slugify(%q)=%q,期望 %q", in, got, want)
		}
	}
}

func TestScanCLITools(t *testing.T) {
	bin := t.TempDir()
	mkExec(t, bin, "codex")
	mkExec(t, bin, "claude")
	mkExec(t, bin, "cursor-agent")

	env := scanEnv{GOOS: "linux", PathDirs: []string{bin}}
	got, err := scanEditors(env)
	if err != nil {
		t.Fatalf("scanEditors: %v", err)
	}
	byID := map[string]Editor{}
	for _, e := range got {
		byID[e.ID] = e
	}
	if len(got) != 3 {
		t.Fatalf("应命中 3 个 CLI 工具: %+v", got)
	}
	codex := byID["codex"]
	if codex.Name != "Codex CLI" || codex.Kind != KindCLI || codex.Args != "" || codex.Icon != "codex" {
		t.Fatalf("codex 字段不符: %+v", codex)
	}
	claude := byID["claude"]
	if claude.Name != "Claude Code" || claude.Kind != KindCLI || claude.Icon != "claude" {
		t.Fatalf("claude 字段不符: %+v", claude)
	}
	if byID["cursor-agent"].Icon != "agent" {
		t.Fatalf("cursor-agent 图标应为 agent: %+v", byID["cursor-agent"])
	}
}
