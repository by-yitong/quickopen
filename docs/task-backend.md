# 任务卡:quickopen 后端(Wails v3 Go 服务)

## 背景与目标
桌面工具「项目快开」:记忆项目路径,一键用终端/VS Code 等已装编辑器打开;支持创建新项目目录(含临时目录)。仓库已由 `wails3 init -n quickopen -t react` 脚手架(wails v3.0.0-beta.4,go.mod module 名 `quickopen`,纯 package main)。你负责 Go 后端全部业务码。

## 仓库与目录
- 仓库根:`/home/admins/.zcode/workspace/default/quickopen`
- 你只在根目录写 Go 文件 + 测试;`frontend/`、`build/` 一律不动。`main.go` 允许改(注册服务、窗口配置),`greetservice.go` 删除。

## 冻结契约(不得增删改名;字段 json tag 照抄)

```go
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

type EditorKind string // "editor" | "terminal" | "cli"【2026-09-29 主 agent 增补:cli=CLI 编码工具(codex/claude/grok…),终端里在项目目录直接执行,默认无参数】
const (
    KindEditor   EditorKind = "editor"
    KindTerminal EditorKind = "terminal"
)

type Editor struct {
    ID         string     `json:"id"`         // 稳定 slug: vscode/zed/konsole/...;手动的 "m-"+8位hex
    Name       string     `json:"name"`
    Icon       string     `json:"icon"`       // 前端图标 slug,见下方扫描表
    Command    string     `json:"command"`    // PATH 中的可执行名或绝对路径;可含空格(flatpak Exec 整行)
    Args       string     `json:"args"`       // 参数模板,含 {path} 占位符,空格分隔,默认 "{path}"
    Kind       EditorKind `json:"kind"`
    Source     string     `json:"source"`     // "scan" | "manual"
    Enabled    bool       `json:"enabled"`
    InTerminal bool       `json:"inTerminal"` // true=终端内编辑器(nvim/vim/hx),启动时包一层终端
}

type Settings struct {
    DefaultPath       string `json:"defaultPath"`       // 新项目默认父目录
    DefaultEditorID   string `json:"defaultEditorId"`   // ""=未设置
    DefaultTerminalID string `json:"defaultTerminalId"` // ""=未设置
}
```

### 服务与公开方法(绑定给前端,签名不得改)

```go
// project_service.go
type ProjectService struct{}
func (s *ProjectService) ListProjects() ([]Project, error)
func (s *ProjectService) AddProject(path, name, note string) (Project, error) // 登记已存在目录;name 为空取 basename
func (s *ProjectService) UpdateProject(project Project) (Project, error)      // 按 ID 全量更新
func (s *ProjectService) DeleteProject(id string, deleteFiles bool) error // 【2026-09-29 主 agent 增补:deleteFiles=true 同时删除磁盘目录,根目录/主目录护栏,失败保留列表项】
func (s *ProjectService) OpenProject(id, editorID string) error               // editorID=="" 用默认编辑器;成功才更新 LastOpenedAt/OpenCount
func (s *ProjectService) CreateProject(name, note, parentPath string) (Project, error) // mkdirAll parent/name;已存在报错;parentPath=="" 用 Settings.DefaultPath
func (s *ProjectService) CreateTempProject(name, note string) (Project, error)  // parent=DefaultPath/tmp;name 空则自动生成;给了 name 则非法字符/已存在报错【2026-09-28 主 agent 修订:加 name 参数,支持自定义临时目录名】

// editor_service.go
type EditorService struct{}
func (s *EditorService) ListEditors() ([]Editor, error)
func (s *EditorService) ScanEditors() ([]Editor, error) // 全量重扫;保留 manual;同名 ID 保留 Enabled/Args 人工改动;消失的 scan 项删除
func (s *EditorService) AddManualEditor(name, command, args string, kind EditorKind) (Editor, error)
func (s *EditorService) UpdateEditor(editor Editor) (Editor, error)
func (s *EditorService) DeleteEditor(id string) error
func (s *EditorService) ReorderEditors(ids []string) error // 【2026-09-29 主 agent 增补:拖拽排序落盘;ids 必须恰好覆盖全部现有 ID;mergeScan 保持既有顺序】
func (s *EditorService) SetDefaultEditor(id string) error   // 需已存在;"" 清除
func (s *EditorService) SetDefaultTerminal(id string) error

// settings_service.go
type SettingsService struct{}
func (s *SettingsService) GetSettings() (Settings, error)
func (s *SettingsService) SaveSettings(settings Settings) (Settings, error) // DefaultPath 不存在则 MkdirAll
func (s *SettingsService) PickDirectory() (string, error) // 原生目录选择框;取消返回 "", nil
```

### 行为细则
- **存储**:`os.UserConfigDir()/quickopen/` 下 `projects.json`/`editors.json`/`settings.json`,RWMutex 保护,写盘用临时文件+rename 原子替换。JSON 存储是主 agent 的既定决策(数据量小、无服务端,不引入 GORM/SQLite)。
- **ID**:`p-`+8位hex(crypto/rand)。
- **临时项目**:父目录固定 `<DefaultPath>/tmp/`;名称用户给了就用(非法字符 `\` `/` 拒绝),没给生成 `tmp-YYYYMMDD-HHMMSS`,已存在则追加 `-2`、`-3`…。IsTemp=true。**绝不自动删除**。
- **CreateProject/CreateTempProject 只创建登记,不负责打开**(前端随后自己调 OpenProject)。
- **OpenProject**:editorID 空→Settings.DefaultEditorID;仍空→错误「未找到可用编辑器,请先在设置中扫描或添加」。编辑器 Kind==terminal 或 InTerminal 时按终端逻辑打开。启动失败返回错误且不更新统计;启动成功(spawn 成功)即更新。
- **错误**:用 `fmt.Errorf` 中文消息,如「目录不存在: %s」「该项目已在列表中」「目标目录已存在: %s」「名称不能包含路径分隔符」。
- **去重**:AddProject/Create 按绝对路径(Clean+Abs)去重。
- **PickDirectory**:`application.Get().Dialog.OpenFile().CanChooseFiles(false).CanChooseDirectories(true).CanCreateDirectories(true).SetTitle("选择目录").PromptForSingleSelection()`(API 位于 wails v3 `pkg/application/dialogs.go`,可再自行 grep 核对)。取消时返回 "" 和 nil(注意 Linux 实现对取消可能返回错误,若错误串含 cancel/取消 则按取消处理)。
- **首次启动初始化**(main.go 里、窗口创建前调用,如 `bootstrap()`):目录不存在→创建三个 json;Settings 为空→DefaultPath=`$HOME/Projects`(MkdirAll);editors.json 不存在→自动跑一次扫描并落盘;然后:DefaultEditorID 为空→优先选 `vscode`,否则第一个 enabled 编辑器;DefaultTerminalID 为空→优先 `konsole`,否则第一个 enabled 终端。

### 扫描(Linux 为主,`runtime.GOOS` 分支,Windows/macOS 尽力而为)
1. PATH 遍历找可执行,命中内置表(命令名→Name/Icon/Args/Kind/InTerminal):
   - 编辑器: `code`→VS Code/vscode,`code-insiders`→VS Code Insiders/vscode,`codium`→VSCodium/vscodium,`cursor`→Cursor/cursor,`zed`→Zed/zed,`subl`→Sublime Text/sublime,`windsurf`→Windsurf/windsurf,`trae`→Trae/trae,`idea`→IntelliJ IDEA/jetbrains,`pycharm`→PyCharm/jetbrains,`webstorm`→WebStorm/jetbrains,`goland`→GoLand/jetbrains,`clion`→CLion/jetbrains,`rider`→Rider/jetbrains,`phpstorm`→PhpStorm/jetbrains,`rubymine`→RubyMine/jetbrains,`fleet`→Fleet/jetbrains,`emacs`→Emacs/emacs,`kate`→Kate/generic;终端内编辑器(InTerminal=true): `nvim`→Neovim/nvim,`vim`→Vim/vim,`hx`→Helix/helix。Args 一律 `{path}`。
   - 终端(Kind=terminal,Icon=terminal): gnome-terminal, konsole, alacritty, kitty, wezterm, tilix, xterm, foot, ptyxis, terminator, deepin-terminal, qterminal, xfce4-terminal;Windows 另查 `wt`,macOS 特殊处理 Terminal.app。
2. .desktop 补扫(PATH 没命中的,如 JetBrains Toolbox/flatpak):目录 `/usr/share/applications`、`~/.local/share/applications`、`/var/lib/flatpak/exports/share/applications`、`~/.local/share/flatpak/exports/share/applications`;解析 `Name=`、`Exec=`(去掉 `%f %F %u %U` 等字段码)、跳过 `NoDisplay=true`/`Type!=Application`;用 Name/Exec 关键词匹配上面同一张编辑器关键词表(如 visual studio code、sublime text、jetbrains、zed…),`Icon` slug 沿用匹配结果否则不收;Exec 含空格(如 `flatpak run ...`)整行存入 Command。ID 用 desktop 文件名去 `.desktop` 后缀。
3. 去重:按最终可执行解析结果;PATH 扫描优先于 desktop。
4. Enabled 默认 true。
- Windows:`exec.LookPath` 同表 + 检查 `%LOCALAPPDATA%\Programs\Microsoft VS Code\bin` 等常见路径;macOS:扫 `/Applications/*.app` 按表映射,启动用 `open -a`。实现可以薄,但**必须编译通过且有基本测试**(用接口/表驱动便于测)。

### 启动器(launcher.go)
- 编辑器:解析 Command(PATH 里 LookPath,或绝对路径);Args 模板 `{path}` 替换为绝对路径,无占位符则追加;含空格的 Command(Linux/mac)用 `sh -c "exec <整行>"` 执行。Unix 上 `SysProcAttr{Setsid: true}` 脱离父进程。
- 终端(内置表按命令名生成 argv;`inner` 为可选的内嵌命令 argv):
  - konsole: `--workdir <path>` + inner 时 `-e <inner...>`
  - gnome-terminal: `--working-directory=<path>` + `-- <inner...>`
  - alacritty: `--working-directory <path>` + `-e <inner...>`
  - kitty: `--directory <path>` + inner 直接追加
  - wezterm: `start --cwd <path>` + `-- <inner...>`
  - tilix: `--working-directory=<path>` + `-e <inner...>`
  - ptyxis: `--working-directory=<path>` + `-- <inner...>`
  - foot: `--working-directory <path>` + inner 追加
  - terminator: `--working-directory=<path>` + `-x <inner...>`
  - xterm: `-e sh -c "cd '<path>' && exec <inner>"`(无 inner 时 `sh -c "cd '<path>' && exec $SHELL"`)
  - 未知/手动终端:用户的 Args 模板替换 {path},inner 原样追加
  - Windows: wt→`-d <path>`;macOS Terminal.app→`open -a Terminal <path>`
- InTerminal 编辑器:inner = [编辑器可执行, path],用 DefaultTerminalID 对应终端打开;无可用终端→错误「未找到可用终端,请在设置中扫描或添加」。
- 表驱动构建 argv(**纯函数便于测试**,如 `buildTerminalArgs(term Editor, path string, inner ...string) []string`),真正 spawn 只在薄薄一层。

## 验收标准
1. `cd /home/admins/.zcode/workspace/default/quickopen && go build . && go vet . && go test .` 全绿(注意:只能用根包 `.`,不能用 `./...`——`build/ios` 等模板目录带平台标签会被误构建)。
2. 测试覆盖(表驱动为主):存储 CRUD+原子写、项目校验(重名/路径分隔符/不存在)、临时名生成与碰撞、终端 argv 表(至少 konsole/gnome-terminal/alacritty/kitty/wezterm/未知)、扫描器(用临时 PATH 目录+假可执行文件+临时 .desktop 目录测,禁止依赖真实机器环境)、bootstrap 默认值。
3. `wails3 generate bindings` 能通过(可自行运行验证,生成物留在 frontend/bindings/)。
4. 交付报告:改了哪些文件、测试结果原文、遗留疑问。

## 禁止事项
- 不得改 `frontend/` 下任何文件(除允许 bindings 生成器写入 `frontend/bindings/`)、不得改 `build/`。
- 不得改契约签名/json tag;不得引入 go.mod 之外的新依赖(wails v3 已在 go.mod;确需标准库以外依赖先停下报告)。
- 不得再派生子 agent;卡外问题写进交付报告带回。
