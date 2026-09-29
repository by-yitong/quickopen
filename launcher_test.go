package main

import (
	"strings"
	"testing"
)

func TestApplyArgsTemplate(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		path string
		want []string
	}{
		{"默认占位", "{path}", "/a b/proj", []string{"/a b/proj"}},
		{"带旗标", "-d {path} --new-window", "/p", []string{"-d", "/p", "--new-window"}},
		{"无占位则追加", "--new-window", "/p", []string{"--new-window", "/p"}},
		{"空模板", "", "/p", []string{"/p"}},
		{"token 内嵌", "--cwd={path}", "/p", []string{"--cwd=/p"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyArgsTemplate(tt.tmpl, tt.path)
			if strings.Join(got, "\x00") != strings.Join(tt.want, "\x00") {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// term 构造终端条目,Command 为命令名,Args 默认 {path}。
func term(command string) Editor {
	return Editor{ID: command, Name: command, Command: command, Args: "{path}", Kind: KindTerminal}
}

func TestBuildTerminalArgsKnown(t *testing.T) {
	p := "/home/u/my proj"
	tests := []struct {
		name  string
		cmd   string
		inner []string
		want  []string
	}{
		{"konsole 无inner", "konsole", nil, []string{"konsole", "--workdir", p}},
		{"konsole 有inner", "konsole", []string{"nvim", p}, []string{"konsole", "--workdir", p, "-e", "nvim", p}},
		{"gnome-terminal 无inner", "gnome-terminal", nil, []string{"gnome-terminal", "--working-directory=" + p}},
		{"gnome-terminal 有inner", "gnome-terminal", []string{"nvim", p}, []string{"gnome-terminal", "--working-directory=" + p, "--", "nvim", p}},
		{"alacritty 无inner", "alacritty", nil, []string{"alacritty", "--working-directory", p}},
		{"alacritty 有inner", "alacritty", []string{"nvim", p}, []string{"alacritty", "--working-directory", p, "-e", "nvim", p}},
		{"kitty 无inner", "kitty", nil, []string{"kitty", "--directory", p}},
		{"kitty 有inner", "kitty", []string{"nvim", p}, []string{"kitty", "--directory", p, "nvim", p}},
		{"wezterm 无inner", "wezterm", nil, []string{"wezterm", "start", "--cwd", p}},
		{"wezterm 有inner", "wezterm", []string{"nvim", p}, []string{"wezterm", "start", "--cwd", p, "--", "nvim", p}},
		{"tilix", "tilix", []string{"nvim", p}, []string{"tilix", "--working-directory=" + p, "-e", "nvim", p}},
		{"ptyxis", "ptyxis", []string{"nvim", p}, []string{"ptyxis", "--working-directory=" + p, "--", "nvim", p}},
		{"foot", "foot", []string{"nvim", p}, []string{"foot", "--working-directory", p, "nvim", p}},
		{"terminator", "terminator", []string{"nvim", p}, []string{"terminator", "--working-directory=" + p, "-x", "nvim", p}},
		{"xterm 无inner", "xterm", nil, []string{"xterm", "-e", "sh", "-c", "cd '/home/u/my proj' && exec $SHELL"}},
		{"xterm 有inner", "xterm", []string{"nvim", p}, []string{"xterm", "-e", "sh", "-c", "cd '/home/u/my proj' && exec 'nvim' '/home/u/my proj'"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTerminalArgs(term(tt.cmd), p, tt.inner...)
			if strings.Join(got, "\x00") != strings.Join(tt.want, "\x00") {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildTerminalArgsUnknown(t *testing.T) {
	p := "/p"
	// 未知终端:模板替换 {path},inner 追加
	unknown := Editor{ID: "myterm", Name: "手动终端", Command: "myterm", Args: "-d {path}", Kind: KindTerminal}
	got := buildTerminalArgs(unknown, p, "nvim", p)
	want := []string{"myterm", "-d", "/p", "nvim", "/p"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("got %v, want %v", got, want)
	}
	// 无占位则追加路径
	unknown.Args = "--custom"
	got = buildTerminalArgs(unknown, p)
	want = []string{"myterm", "--custom", "/p"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("got %v, want %v", got, want)
	}
	// 命令含空格 → sh -c exec 整行
	spaced := Editor{ID: "flatpak-term", Command: "flatpak run org.x.Term", Args: "{path}", Kind: KindTerminal}
	got = buildTerminalArgs(spaced, p, "inner1")
	if got[0] != "sh" || got[1] != "-c" {
		t.Fatalf("含空格命令应包 sh -c: %v", got)
	}
	if !strings.HasPrefix(got[2], "exec flatpak run org.x.Term /p") {
		t.Fatalf("sh 脚本不符: %q", got[2])
	}
}

func TestBuildEditorArgv(t *testing.T) {
	// 普通命令
	ed := Editor{ID: "vscode", Command: "code", Args: "{path}"}
	got, err := buildEditorArgv(ed, "/p")
	if err != nil || got[0] != "code" || got[1] != "/p" {
		t.Fatalf("got %v err %v", got, err)
	}
	// 含空格命令 → sh -c exec
	ed = Editor{ID: "idea", Command: "/opt/idea/bin/idea.sh", Args: "{path}"}
	got, err = buildEditorArgv(ed, "/p")
	if err != nil || got[0] != "/opt/idea/bin/idea.sh" {
		t.Fatalf("got %v err %v", got, err)
	}
	spaced := Editor{ID: "flat-ed", Command: "flatpak run com.x.Editor", Args: "{path}"}
	got, err = buildEditorArgv(spaced, "/p")
	if err != nil || got[0] != "sh" || got[1] != "-c" || got[2] != "exec flatpak run com.x.Editor /p" {
		t.Fatalf("got %v err %v", got, err)
	}
	// 命令为空
	if _, err := buildEditorArgv(Editor{Command: " "}, "/p"); err == nil {
		t.Fatalf("空命令应报错")
	}
}

func TestShellQuote(t *testing.T) {
	if got := shellQuote("a'b"); got != `'a'\''b'` {
		t.Fatalf("got %q", got)
	}
	if got := shellQuote("/plain/path"); got != "'/plain/path'" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildCliInner(t *testing.T) {
	// 默认无参数:只有命令本身(工具以终端 cwd 即项目目录为工作目录)
	ed := Editor{ID: "codex", Command: "codex", Kind: KindCLI}
	if got := buildCliInner(ed, "/tmp/p"); strings.Join(got, "\x1f") != "codex" {
		t.Fatalf("无参数应为 [codex]: %v", got)
	}
	// 自定义参数原样带上
	ed.Args = "--model o4-mini"
	if got := buildCliInner(ed, "/tmp/p"); strings.Join(got, "\x1f") != "codex\x1f--model\x1fo4-mini" {
		t.Fatalf("自定义参数不符: %v", got)
	}
	// 模板显式写 {path} 才替换
	ed.Args = "{path}"
	if got := buildCliInner(ed, "/tmp/p"); strings.Join(got, "\x1f") != "codex\x1f/tmp/p" {
		t.Fatalf("{path} 应替换: %v", got)
	}
}

func TestLaunchProjectCLI(t *testing.T) {
	var recorded [][]string
	old := startProcessFunc
	startProcessFunc = func(argv []string) error {
		recorded = append(recorded, argv)
		return nil
	}
	t.Cleanup(func() { startProcessFunc = old })

	cli := Editor{ID: "codex", Name: "Codex CLI", Command: "codex", Kind: KindCLI, Source: "scan", Enabled: true}
	term := Editor{ID: "konsole", Name: "Konsole", Command: "konsole", Kind: KindTerminal, Source: "scan", Enabled: true}
	editors := []Editor{cli, term}
	settings := Settings{DefaultTerminalID: "konsole"}

	if err := launchProject(cli, "/tmp/proj", editors, settings); err != nil {
		t.Fatalf("launchProject: %v", err)
	}
	want := "konsole\x1f--workdir\x1f/tmp/proj\x1f-e\x1fcodex"
	if len(recorded) != 1 || strings.Join(recorded[0], "\x1f") != want {
		t.Fatalf("CLI 应以终端打开并执行命令: %v", recorded)
	}
}
