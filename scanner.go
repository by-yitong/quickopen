package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// 编辑器扫描:PATH 遍历 + .desktop 补扫(Linux),Windows/macOS 尽力而为。
// 全部走 scanEnv 注入路径,测试不依赖真实机器。

// editorSpec 内置表条目:命令名 → 显示名/图标/类别/是否终端内。
type editorSpec struct {
	Name       string
	Icon       string
	Kind       EditorKind
	InTerminal bool
}

// builtinEditors PATH 扫描内置表(编辑器)。
var builtinEditors = map[string]editorSpec{
	"code":           {"VS Code", "vscode", KindEditor, false},
	"code-insiders":  {"VS Code Insiders", "vscode", KindEditor, false},
	"codium":         {"VSCodium", "vscodium", KindEditor, false},
	"cursor":         {"Cursor", "cursor", KindEditor, false},
	"zed":            {"Zed", "zed", KindEditor, false},
	"subl":           {"Sublime Text", "sublime", KindEditor, false},
	"windsurf":       {"Windsurf", "windsurf", KindEditor, false},
	"trae":           {"Trae", "trae", KindEditor, false},
	"idea":           {"IntelliJ IDEA", "jetbrains", KindEditor, false},
	"pycharm":        {"PyCharm", "jetbrains", KindEditor, false},
	"webstorm":       {"WebStorm", "jetbrains", KindEditor, false},
	"goland":         {"GoLand", "jetbrains", KindEditor, false},
	"clion":          {"CLion", "jetbrains", KindEditor, false},
	"rider":          {"Rider", "jetbrains", KindEditor, false},
	"phpstorm":       {"PhpStorm", "jetbrains", KindEditor, false},
	"rubymine":       {"RubyMine", "jetbrains", KindEditor, false},
	"fleet":          {"Fleet", "jetbrains", KindEditor, false},
	"emacs":          {"Emacs", "emacs", KindEditor, false},
	"kate":           {"Kate", "generic", KindEditor, false},
	"nvim":           {"Neovim", "nvim", KindEditor, true},
	"vim":            {"Vim", "vim", KindEditor, true},
	"hx":             {"Helix", "helix", KindEditor, true},
}

// builtinCLI PATH 扫描内置表(CLI 编码工具):命令名 → 显示名/图标 slug。
// 启动 = 打开终端到项目目录直接执行命令;默认不追加参数(工具以 cwd 为工作目录)。
var builtinCLI = map[string][2]string{
	"codex":        {"Codex CLI", "codex"},
	"claude":       {"Claude Code", "claude"},
	"gemini":       {"Gemini CLI", "gemini"},
	"grok":         {"Grok CLI", "grok"},
	"aider":        {"Aider", "aider"},
	"cursor-agent": {"Cursor Agent", "agent"},
	"qwen":         {"Qwen Code", "agent"},
	"opencode":     {"OpenCode", "agent"},
	"crush":        {"Crush", "agent"},
}

// builtinTerminals PATH 扫描内置表(终端):命令名 → 显示名。
var builtinTerminals = map[string]string{
	"gnome-terminal":  "GNOME Terminal",
	"konsole":         "Konsole",
	"alacritty":       "Alacritty",
	"kitty":           "kitty",
	"wezterm":         "WezTerm",
	"tilix":           "Tilix",
	"xterm":           "XTerm",
	"foot":            "foot",
	"ptyxis":          "Ptyxis",
	"terminator":      "Terminator",
	"deepin-terminal": "Deepin Terminal",
	"qterminal":       "QTerminal",
	"xfce4-terminal":  "Xfce4 Terminal",
}

// desktopMatch .desktop 关键词匹配表:命中关键词 → 图标/类别;Name 为空则沿用 desktop 文件的 Name。
// 顺序敏感:更具体的关键词放前面(如 neovim 在 vim 之前)。
type desktopMatch struct {
	Keywords   []string
	Name       string
	Icon       string
	Kind       EditorKind
	InTerminal bool
}

var desktopKeywords = []desktopMatch{
	{[]string{"visual studio code", "visualstudiocode", "com.visualstudio.code"}, "VS Code", "vscode", KindEditor, false},
	{[]string{"code insiders", "code-insiders"}, "VS Code Insiders", "vscode", KindEditor, false},
	{[]string{"vscodium"}, "VSCodium", "vscodium", KindEditor, false},
	{[]string{"sublime text", "sublimetext"}, "Sublime Text", "sublime", KindEditor, false},
	{[]string{"intellij idea"}, "IntelliJ IDEA", "jetbrains", KindEditor, false},
	{[]string{"pycharm"}, "PyCharm", "jetbrains", KindEditor, false},
	{[]string{"webstorm"}, "WebStorm", "jetbrains", KindEditor, false},
	{[]string{"goland"}, "GoLand", "jetbrains", KindEditor, false},
	{[]string{"clion"}, "CLion", "jetbrains", KindEditor, false},
	{[]string{"rider"}, "Rider", "jetbrains", KindEditor, false},
	{[]string{"phpstorm"}, "PhpStorm", "jetbrains", KindEditor, false},
	{[]string{"rubymine"}, "RubyMine", "jetbrains", KindEditor, false},
	{[]string{"fleet"}, "Fleet", "jetbrains", KindEditor, false},
	{[]string{"jetbrains"}, "", "jetbrains", KindEditor, false},
	{[]string{"neovim", "nvim"}, "Neovim", "nvim", KindEditor, true},
	{[]string{"gvim"}, "Vim", "vim", KindEditor, true},
	{[]string{"vim"}, "Vim", "vim", KindEditor, true},
	{[]string{"helix"}, "Helix", "helix", KindEditor, true},
	{[]string{"emacs"}, "Emacs", "emacs", KindEditor, false},
	{[]string{"kate"}, "Kate", "generic", KindEditor, false},
	{[]string{"zed"}, "Zed", "zed", KindEditor, false},
	{[]string{"cursor"}, "Cursor", "cursor", KindEditor, false},
	{[]string{"windsurf"}, "Windsurf", "windsurf", KindEditor, false},
	{[]string{"trae"}, "Trae", "trae", KindEditor, false},
	{[]string{"gnome-terminal", "gnome terminal"}, "GNOME Terminal", "terminal", KindTerminal, false},
	{[]string{"konsole"}, "Konsole", "terminal", KindTerminal, false},
	{[]string{"alacritty"}, "Alacritty", "terminal", KindTerminal, false},
	{[]string{"kitty"}, "kitty", "terminal", KindTerminal, false},
	{[]string{"wezterm"}, "WezTerm", "terminal", KindTerminal, false},
	{[]string{"tilix"}, "Tilix", "terminal", KindTerminal, false},
	{[]string{"xterm"}, "XTerm", "terminal", KindTerminal, false},
	{[]string{"ptyxis"}, "Ptyxis", "terminal", KindTerminal, false},
	{[]string{"terminator"}, "Terminator", "terminal", KindTerminal, false},
	{[]string{"deepin-terminal", "deepin terminal"}, "Deepin Terminal", "terminal", KindTerminal, false},
	{[]string{"qterminal"}, "QTerminal", "terminal", KindTerminal, false},
	{[]string{"xfce4-terminal", "xfce4 terminal"}, "Xfce4 Terminal", "terminal", KindTerminal, false},
	{[]string{"foot"}, "foot", "terminal", KindTerminal, false},
	{[]string{"terminal"}, "Terminal", "terminal", KindTerminal, false},
}

// scanEnv 扫描环境,全部可注入。
type scanEnv struct {
	GOOS        string
	PathDirs    []string // PATH 各目录
	DesktopDirs []string // .desktop 目录
	AppsDir     string   // macOS /Applications
	HomeDir     string
}

// scanSystemEditors 用真实系统环境扫描。
func scanSystemEditors() ([]Editor, error) {
	home, _ := os.UserHomeDir()
	env := scanEnv{
		GOOS:    runtime.GOOS,
		HomeDir: home,
	}
	if env.GOOS == "windows" {
		env.AppsDir = "" // Windows 不扫 .app
	} else {
		env.AppsDir = "/Applications"
	}
	// 桌面入口启动时 GUI 进程的 PATH 很窄(缺 nvm/volta 等用户目录),
	// 叠加登录 shell 的 PATH 和常见用户目录,保证 CLI 工具(codex/claude…)能被扫到。
	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	pathDirs = append(pathDirs, userLoginPathDirs()...)
	pathDirs = append(pathDirs, commonUserBinDirs(home)...)
	for _, dir := range pathDirs {
		if dir != "" {
			env.PathDirs = append(env.PathDirs, dir)
		}
	}
	env.DesktopDirs = []string{
		"/usr/share/applications",
		filepath.Join(home, ".local/share/applications"),
		"/var/lib/flatpak/exports/share/applications",
		filepath.Join(home, ".local/share/flatpak/exports/share/applications"),
	}
	return scanEditors(env)
}

// userLoginPathDirs 借登录 shell 求得用户完整 PATH(nvm/rustup 等在 rc 文件里追加的目录)。
// 只在类 Unix 上尝试;失败或超时返回 nil,调用方退回进程 PATH。
func userLoginPathDirs() []string {
	if runtime.GOOS == "windows" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	shell := "bash"
	if _, err := exec.LookPath(shell); err != nil {
		shell = "sh"
	}
	out, err := exec.CommandContext(ctx, shell, "-lc", "echo $PATH").Output()
	if err != nil {
		return nil
	}
	return filepath.SplitList(strings.TrimSpace(string(out)))
}

// commonUserBinDirs 常见的用户级可执行目录(nvm 之外的主流安装位置)。
func commonUserBinDirs(home string) []string {
	if home == "" {
		return nil
	}
	return []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, "bin"),
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, ".volta", "bin"),
		filepath.Join(home, ".local", "share", "pnpm"),
	}
}

// scanEditors 执行扫描:PATH 优先,desktop 补扫;按 ID 与最终可执行解析结果去重。
func scanEditors(env scanEnv) ([]Editor, error) {
	var out []Editor
	seenID := map[string]bool{}
	seenCmd := map[string]bool{}

	add := func(e Editor) bool {
		if e.ID == "" || seenID[e.ID] {
			return false
		}
		seenID[e.ID] = true
		if !strings.Contains(e.Command, " ") {
			// 记录单命令名,供 desktop 补扫按可执行去重(PATH 优先)
			seenCmd[strings.ToLower(filepath.Base(e.Command))] = true
		}
		out = append(out, e)
		return true
	}

	switch env.GOOS {
	case "windows":
		scanWindows(env, add)
	case "darwin":
		scanDarwinApps(env, add)
		scanPathDirs(env, add)
	default:
		scanPathDirs(env, add)
		scanDesktopDirs(env, add, seenCmd)
	}
	return out, nil
}

// scanPathDirs 遍历 PATH 目录,命中内置表的普通可执行文件收为编辑器/终端。
func scanPathDirs(env scanEnv, add func(Editor) bool) {
	for _, dir := range env.PathDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // 目录不可读很常见,跳过
		}
		for _, entry := range entries {
			name := entry.Name()
			key := execKey(env.GOOS, name)
			if key == "" {
				continue
			}
			info, err := entry.Info()
			if err != nil || info.IsDir() {
				continue
			}
			if !isExecutable(env.GOOS, info) {
				continue
			}
			if spec, ok := builtinEditors[key]; ok {
				add(Editor{
					ID: key, Name: spec.Name, Icon: spec.Icon,
					// 存绝对路径:GUI 进程 PATH 窄,启动时不依赖运行时 PATH 解析
					Command: filepath.Join(dir, name),
					Args:    "{path}", Kind: spec.Kind, Source: "scan",
					Enabled: true, InTerminal: spec.InTerminal,
				})
				continue
			}
			if cli, ok := builtinCLI[key]; ok {
				add(Editor{
					ID: key, Name: cli[0], Icon: cli[1],
					Command: filepath.Join(dir, name),
					Args:    "", Kind: KindCLI, Source: "scan", Enabled: true,
				})
				continue
			}
			if disp, ok := builtinTerminals[key]; ok {
				add(Editor{
					ID: key, Name: disp, Icon: "terminal",
					Command: filepath.Join(dir, name),
					Args:    "{path}", Kind: KindTerminal, Source: "scan", Enabled: true,
				})
			}
		}
	}
}

// execKey 从文件名得到内置表查找键;Windows 下去掉 .exe/.cmd/.bat 后缀。
func execKey(goos, name string) string {
	lower := strings.ToLower(name)
	if goos == "windows" {
		for _, ext := range []string{".exe", ".cmd", ".bat"} {
			lower = strings.TrimSuffix(lower, ext)
		}
	}
	return lower
}

func isExecutable(goos string, info os.FileInfo) bool {
	if goos == "windows" {
		return true
	}
	return info.Mode()&0o111 != 0
}

// scanDesktopDirs 补扫 .desktop 文件;PATH 已覆盖的同名 ID 或同可执行项跳过。
func scanDesktopDirs(env scanEnv, add func(Editor) bool, seenCmd map[string]bool) {
	for _, dir := range env.DesktopDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".desktop") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			df := parseDesktopContent(string(data))
			if df.NoDisplay || df.Type != "Application" || df.Exec == "" {
				continue
			}
			cleanedExec := stripFieldCodes(df.Exec)
			if cleanedExec == "" {
				continue
			}
			// 按最终可执行解析结果去重:可执行名已被 PATH 扫描或之前的 desktop 项收编则跳过
			key := desktopDedupKey(cleanedExec)
			if seenCmd[key] {
				continue
			}
			m := matchDesktopKeywords(df.Name + " " + cleanedExec)
			if m == nil {
				continue // 图标 slug 沿用匹配结果,匹配不上不收
			}
			dispName := m.Name
			if dispName == "" {
				dispName = df.Name
			}
			id := strings.TrimSuffix(name, ".desktop")
			ed := Editor{
				ID: id, Name: dispName, Icon: m.Icon, Command: cleanedExec,
				Args: "{path}", Kind: m.Kind, Source: "scan",
				Enabled: true, InTerminal: m.InTerminal,
			}
			if add(ed) {
				seenCmd[key] = true
			}
		}
	}
}

// desktopDedupKey 可执行去重键:普通命令取首词;flatpak 取应用 ID(最后一个非选项 token)。
func desktopDedupKey(cleanedExec string) string {
	fields := strings.Fields(cleanedExec)
	if len(fields) == 0 {
		return ""
	}
	first := strings.ToLower(filepath.Base(fields[0]))
	if strings.EqualFold(first, "flatpak") {
		for i := len(fields) - 1; i > 0; i-- {
			tok := fields[i]
			if !strings.HasPrefix(tok, "-") && !strings.Contains(tok, "=") {
				return strings.ToLower(tok)
			}
		}
	}
	return first
}

// matchDesktopKeywords 大小写不敏感地按关键词表匹配,返回首个命中项。
func matchDesktopKeywords(haystack string) *desktopMatch {
	lower := strings.ToLower(haystack)
	for i := range desktopKeywords {
		m := &desktopKeywords[i]
		for _, kw := range m.Keywords {
			if strings.Contains(lower, kw) {
				return m
			}
		}
	}
	return nil
}

// desktopFile .desktop 文件解析结果(只取关心字段)。
type desktopFile struct {
	Name      string
	Exec      string
	NoDisplay bool
	Type      string
}

// parseDesktopContent 解析 [Desktop Entry] 段的 Name/Exec/NoDisplay/Type。
func parseDesktopContent(content string) desktopFile {
	var df desktopFile
	inEntry := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inEntry = line == "[Desktop Entry]"
			continue
		}
		if !inEntry {
			continue
		}
		switch {
		case strings.HasPrefix(line, "Name="):
			if df.Name == "" {
				df.Name = strings.TrimPrefix(line, "Name=")
			}
		case strings.HasPrefix(line, "Exec="):
			if df.Exec == "" {
				df.Exec = strings.TrimPrefix(line, "Exec=")
			}
		case strings.HasPrefix(line, "NoDisplay="):
			df.NoDisplay = df.NoDisplay || strings.TrimSpace(strings.TrimPrefix(line, "NoDisplay=")) == "true"
		case strings.HasPrefix(line, "Type="):
			if df.Type == "" {
				df.Type = strings.TrimSpace(strings.TrimPrefix(line, "Type="))
			}
		}
	}
	return df
}

// stripFieldCodes 去掉 Exec 里的 %f %F %u %U 等字段码,余下部分以空格连接。
func stripFieldCodes(exec string) string {
	var kept []string
	for _, tok := range strings.Fields(exec) {
		if isFieldCode(tok) {
			continue
		}
		kept = append(kept, tok)
	}
	return strings.Join(kept, " ")
}

func isFieldCode(tok string) bool {
	// 形如 %f、%F、%U、%k 等;%% 表示字面百分号,保留
	if !strings.HasPrefix(tok, "%") {
		return false
	}
	body := tok[1:]
	if body == "%" {
		return false
	}
	for _, r := range body {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		default:
			return false
		}
	}
	return true
}

// scanWindows Windows 下额外检查常见安装目录(尽力而为)。
func scanWindows(env scanEnv, add func(Editor) bool) {
	extra := []string{
		filepath.Join(env.HomeDir, "AppData", "Local", "Programs", "Microsoft VS Code", "bin"),
		filepath.Join(env.HomeDir, "AppData", "Local", "Programs", "Microsoft VS Code Insiders", "bin"),
		filepath.Join(env.HomeDir, "scoop", "shims"),
	}
	env.PathDirs = append(append([]string{}, env.PathDirs...), extra...)
	scanPathDirs(env, add)
	// Windows Terminal
	if lookupInDirs("wt", env.PathDirs) {
		add(Editor{ID: "wt", Name: "Windows Terminal", Icon: "terminal", Command: "wt",
			Args: "{path}", Kind: KindTerminal, Source: "scan", Enabled: true})
	}
}

func lookupInDirs(name string, dirs []string) bool {
	for _, dir := range dirs {
		for _, ext := range []string{"", ".exe", ".cmd", ".bat"} {
			if fileExists(filepath.Join(dir, name+ext)) {
				return true
			}
		}
	}
	return false
}

// scanDarwinApps 扫描 /Applications 下的 .app 包,按关键词表映射(尽力而为)。
func scanDarwinApps(env scanEnv, add func(Editor) bool) {
	if env.AppsDir == "" {
		return
	}
	entries, err := os.ReadDir(env.AppsDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".app") {
			continue // .app 包本身是目录
		}
		base := strings.TrimSuffix(name, ".app")
		m := matchDesktopKeywords(base)
		if m == nil {
			continue
		}
		dispName := m.Name
		if dispName == "" {
			dispName = base
		}
		kind := m.Kind
		cmd := filepath.Join(env.AppsDir, name)
		if kind == KindTerminal && strings.EqualFold(base, "terminal") {
			// Terminal.app 启动统一走 open -a Terminal
			cmd = "Terminal"
		}
		add(Editor{
			ID: slugify(base), Name: dispName, Icon: m.Icon, Command: cmd,
			Args: "{path}", Kind: kind, Source: "scan", Enabled: true, InTerminal: m.InTerminal,
		})
	}
}

// slugify 简易 slug:小写、空格与下划线转连字符。
func slugify(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
