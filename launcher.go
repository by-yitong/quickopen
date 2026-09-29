package main

import (
	"fmt"
	"runtime"
	"strings"
)

// 启动器:argv 构建为纯函数,便于表驱动测试;真正 spawn 只在 startProcess 一层。

// startProcessFunc 可注入,测试里替换成记录器,不真正起进程。
var startProcessFunc = startProcess

// applyArgsTemplate 把参数模板按空格拆分,含 {path} 的 token 替换为实际路径;
// 无占位符则把路径追加到末尾;模板为空视作 "{path}"。
func applyArgsTemplate(tmpl, path string) []string {
	tokens := strings.Fields(tmpl)
	if len(tokens) == 0 {
		return []string{path}
	}
	found := false
	out := make([]string, 0, len(tokens)+1)
	for _, t := range tokens {
		if strings.Contains(t, "{path}") {
			t = strings.ReplaceAll(t, "{path}", path)
			found = true
		}
		out = append(out, t)
	}
	if !found {
		out = append(out, path)
	}
	return out
}

// shellQuote 单引号包裹,内部单引号转义为 '\''。
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// buildEditorArgv 构建普通编辑器启动 argv。
// 含空格的 Command(Linux/mac)用 sh -c "exec <整行>" 执行;Windows 直接 argv。
func buildEditorArgv(ed Editor, path string) ([]string, error) {
	if strings.TrimSpace(ed.Command) == "" {
		return nil, fmt.Errorf("编辑器命令为空: %s", ed.Name)
	}
	tokens := applyArgsTemplate(ed.Args, path)
	// macOS .app 包统一 open -a 启动
	if runtime.GOOS == "darwin" && strings.HasSuffix(ed.Command, ".app") {
		return []string{"open", "-a", ed.Command, path}, nil
	}
	if runtime.GOOS != "windows" && strings.Contains(ed.Command, " ") {
		line := ed.Command
		if len(tokens) > 0 {
			line += " " + strings.Join(tokens, " ")
		}
		return []string{"sh", "-c", "exec " + line}, nil
	}
	return append([]string{ed.Command}, tokens...), nil
}

// buildTerminalArgs 构建终端启动 argv(含命令名),inner 为可选的内嵌命令 argv。
// 已知终端按内置表生成;未知/手动终端用 Args 模板替换 {path},inner 原样追加。
func buildTerminalArgs(term Editor, path string, inner ...string) []string {
	cmd := term.Command
	switch cmd {
	case "konsole":
		args := []string{cmd, "--workdir", path}
		if len(inner) > 0 {
			args = append(args, "-e")
			args = append(args, inner...)
		}
		return args
	case "gnome-terminal":
		args := []string{cmd, "--working-directory=" + path}
		if len(inner) > 0 {
			args = append(args, "--")
			args = append(args, inner...)
		}
		return args
	case "alacritty":
		args := []string{cmd, "--working-directory", path}
		if len(inner) > 0 {
			args = append(args, "-e")
			args = append(args, inner...)
		}
		return args
	case "kitty":
		args := []string{cmd, "--directory", path}
		return append(args, inner...)
	case "wezterm":
		args := []string{cmd, "start", "--cwd", path}
		if len(inner) > 0 {
			args = append(args, "--")
			args = append(args, inner...)
		}
		return args
	case "tilix":
		args := []string{cmd, "--working-directory=" + path}
		if len(inner) > 0 {
			args = append(args, "-e")
			args = append(args, inner...)
		}
		return args
	case "ptyxis":
		args := []string{cmd, "--working-directory=" + path}
		if len(inner) > 0 {
			args = append(args, "--")
			args = append(args, inner...)
		}
		return args
	case "foot":
		args := []string{cmd, "--working-directory", path}
		return append(args, inner...)
	case "terminator":
		args := []string{cmd, "--working-directory=" + path}
		if len(inner) > 0 {
			args = append(args, "-x")
			args = append(args, inner...)
		}
		return args
	case "xterm":
		innerCmd := "$SHELL"
		if len(inner) > 0 {
			quoted := make([]string, len(inner))
			for i, w := range inner {
				quoted[i] = shellQuote(w)
			}
			innerCmd = strings.Join(quoted, " ")
		}
		script := "cd " + shellQuote(path) + " && exec " + innerCmd
		return []string{cmd, "-e", "sh", "-c", script}
	case "wt":
		if runtime.GOOS == "windows" {
			return append([]string{cmd, "-d", path}, inner...)
		}
	case "Terminal", "terminal":
		if runtime.GOOS == "darwin" {
			// macOS Terminal.app:open -a Terminal <path>
			return []string{"open", "-a", "Terminal", path}
		}
	}
	// 未知/手动终端:Args 模板替换 {path},inner 各词加引号(路径可能含空格)
	tokens := applyArgsTemplate(term.Args, path)
	if runtime.GOOS != "windows" && strings.Contains(cmd, " ") {
		line := cmd
		if len(tokens) > 0 {
			line += " " + strings.Join(tokens, " ")
		}
		if len(inner) > 0 {
			quoted := make([]string, len(inner))
			for i, w := range inner {
				quoted[i] = shellQuote(w)
			}
			line += " " + strings.Join(quoted, " ")
		}
		return []string{"sh", "-c", "exec " + line}
	}
	args := append([]string{cmd}, tokens...)
	return append(args, inner...)
}

// launchProject 启动项目:按编辑器类别分派;InTerminal 编辑器与 CLI 工具都包一层默认终端。
func launchProject(ed Editor, path string, editors []Editor, settings Settings) error {
	switch {
	case ed.Kind == KindTerminal:
		err := startProcessFunc(buildTerminalArgs(ed, path))
		return err
	case ed.Kind == KindCLI:
		term, err := pickTerminal(editors, settings)
		if err != nil {
			return err
		}
		return startProcessFunc(buildTerminalArgs(term, path, buildCliInner(ed, path)...))
	case ed.InTerminal:
		term, err := pickTerminal(editors, settings)
		if err != nil {
			return err
		}
		inner := append(strings.Fields(ed.Command), path)
		err = startProcessFunc(buildTerminalArgs(term, path, inner...))
		return err
	default:
		argv, err := buildEditorArgv(ed, path)
		if err != nil {
			return err
		}
		err = startProcessFunc(argv)
		return err
	}
}

// buildCliInner CLI 工具的终端内命令:命令本身 + 参数模板;
// 不自动追加 {path}(codex/claude 等以终端 cwd 即项目目录为工作目录),模板里显式写 {path} 才会替换。
func buildCliInner(ed Editor, path string) []string {
	inner := strings.Fields(ed.Command)
	for _, t := range strings.Fields(ed.Args) {
		inner = append(inner, strings.ReplaceAll(t, "{path}", path))
	}
	return inner
}

// pickTerminal 选终端:默认 DefaultTerminalID;没有则报错。
func pickTerminal(editors []Editor, settings Settings) (Editor, error) {
	if settings.DefaultTerminalID != "" {
		for _, e := range editors {
			if e.ID == settings.DefaultTerminalID && e.Kind == KindTerminal {
				return e, nil
			}
		}
	}
	return Editor{}, fmt.Errorf("未找到可用终端,请在设置中扫描或添加")
}
