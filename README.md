# QuickOpen · 项目快开

一个跨平台桌面小工具：把常用项目收进列表，一键用你喜欢的编辑器或终端打开。基于 [Wails v3](https://v3.wails.io/)（Go + React）构建，支持 Windows / macOS / Linux。

## 功能

- **项目管理**：手动添加项目，或扫描目录自动发现项目；支持备注、临时项目标记、打开次数与最近打开时间统计。
- **编辑器/终端自动发现**：扫描 PATH 自动识别已安装的工具——VS Code、Cursor、Zed、Windsurf、Trae、JetBrains 全家桶（IDEA / GoLand / PyCharm…）、Sublime、Vim / Neovim、Helix、Emacs 等，以及各平台常见终端。
- **一键启动**：点击项目即用默认编辑器打开；每个项目可快速切换编辑器或终端启动，超过 6 个收进「+N」弹层。
- **CLI 编码工具**：支持 codex / claude / grok 等 CLI 工具，在终端中直接进入项目目录执行。
- **自定义命令**：手动添加任意编辑器/终端，参数模板支持 `{path}` 占位符。
- **设置**：默认项目目录、默认编辑器、默认终端；配置数据存放在系统标准目录（XDG）。

## 安装

### Windows

从源码构建 NSIS 安装包（需要 [Go](https://go.dev/) 1.25+、[Node.js](https://nodejs.org/)、[NSIS](https://nsis.sourceforge.io/)、[wails3 CLI](https://v3.wails.io/getting-started/installation/)）：

```bash
wails3 task build                      # 构建 bin/quickopen.exe
wails3 task package INSTALL_SCOPE=user # 打包用户级 NSIS 安装包（免管理员）
```

安装包输出在 `bin/quickopen-amd64-installer.exe`，支持静默安装：`installer.exe /S /D=<目录>`。也可以直接运行 `bin/quickopen.exe`（绿色版）。

### 其他平台

```bash
wails3 task build        # 当前平台
wails3 task build GOOS=darwin   # 交叉编译（见 Taskfile.yml 各平台任务）
```

> 依赖 WebView2（Windows 10/11 一般自带）、WKWebView（macOS）或 WebKitGTK（Linux）。

## 开发

```bash
wails3 dev      # 开发模式，前后端热重载
go test ./...   # 后端测试
```

项目结构：

```
├── main.go / bootstrap.go   # 应用入口与首次启动初始化
├── project_service.go       # 项目增删改查、目录扫描
├── editor_service.go        # 编辑器/终端发现与管理
├── scanner.go               # PATH 扫描（编辑器/终端/CLI）
├── launcher.go              # 进程启动（编辑器/终端包装）
├── settings_service.go      # 应用设置
├── storage.go               # 配置持久化（XDG 目录）
└── frontend/                # React + TypeScript + Tailwind v4 前端
```

## 技术栈

- **后端**：Go · Wails v3
- **前端**：React 18 · TypeScript · Vite · Tailwind CSS v4
- **打包**：NSIS（Windows）
