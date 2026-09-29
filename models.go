package main

// 契约模型,字段与 json tag 与任务卡逐字一致,不得改动。

// Project 项目条目。
type Project struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Note         string `json:"note"`
	Path         string `json:"path"`
	IsTemp       bool   `json:"isTemp"`
	CreatedAt    int64  `json:"createdAt"`    // unix 毫秒
	LastOpenedAt int64  `json:"lastOpenedAt"` // unix 毫秒,0=从未打开
	OpenCount    int    `json:"openCount"`
}

// EditorKind 编辑器类别:"editor" | "terminal" | "cli"。
type EditorKind string

const (
	KindEditor   EditorKind = "editor"
	KindTerminal EditorKind = "terminal"
	KindCLI      EditorKind = "cli" // CLI 编码工具(codex/claude/grok…):终端里在项目目录直接执行
)

// Editor 编辑器/终端条目。
type Editor struct {
	ID         string     `json:"id"`         // 稳定 slug: vscode/zed/konsole/...;手动的 "m-"+8位hex
	Name       string     `json:"name"`
	Icon       string     `json:"icon"`    // 前端图标 slug
	Command    string     `json:"command"` // PATH 中的可执行名或绝对路径;可含空格(flatpak Exec 整行)
	Args       string     `json:"args"`    // 参数模板,含 {path} 占位符,空格分隔,默认 "{path}"
	Kind       EditorKind `json:"kind"`
	Source     string     `json:"source"` // "scan" | "manual"
	Enabled    bool       `json:"enabled"`
	InTerminal bool       `json:"inTerminal"` // true=终端内编辑器(nvim/vim/hx),启动时包一层终端
}

// Settings 应用设置。
type Settings struct {
	DefaultPath       string `json:"defaultPath"`       // 新项目默认父目录
	DefaultEditorID   string `json:"defaultEditorId"`   // ""=未设置
	DefaultTerminalID string `json:"defaultTerminalId"` // ""=未设置
}
