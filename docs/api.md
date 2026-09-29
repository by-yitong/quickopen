# quickopen API 文档

Go 后端通过 Wails v3 服务绑定暴露给前端。前端从 `frontend/bindings/quickopen/*.js` 导入(由 `wails3 generate bindings` 生成,勿手改):

```js
import * as Projects from "../bindings/quickopen/projectservice";
import * as Editors from "../bindings/quickopen/editorservice";
import * as SettingsSrv from "../bindings/quickopen/settingsservice";
import { Project, Editor, Settings } from "../bindings/quickopen/models";
```

所有方法返回 `CancellablePromise`,错误以中文消息 reject。

## 数据模型

```ts
interface Project {
  id: string;           // "p-"+8位hex
  name: string;
  note: string;
  path: string;         // 绝对路径(Clean+Abs)
  isTemp: boolean;      // 临时项目
  createdAt: number;    // unix 毫秒
  lastOpenedAt: number; // unix 毫秒,0=从未打开
  openCount: number;
}

type EditorKind = "editor" | "terminal";

interface Editor {
  id: string;         // 稳定 slug: vscode/zed/konsole/...;手动的 "m-"+8位hex
  name: string;
  icon: string;       // 图标 slug: vscode/vscodium/cursor/zed/sublime/jetbrains/emacs/kate/nvim/vim/helix/windsurf/trae/terminal/generic
  command: string;    // PATH 中的可执行名或绝对路径;可含空格(flatpak Exec 整行)
  args: string;       // 参数模板,含 {path} 占位符,空格分隔,默认 "{path}"
  kind: EditorKind;
  source: "scan" | "manual";
  enabled: boolean;
  inTerminal: boolean; // true=终端内编辑器(nvim/vim/hx),启动时包一层终端
}

interface Settings {
  defaultPath: string;       // 新项目默认父目录
  defaultEditorId: string;   // ""=未设置
  defaultTerminalId: string; // ""=未设置
}
```

## ProjectService(projectservice.js)

| 方法 | 参数 | 返回 | 错误 |
| --- | --- | --- | --- |
| `ListProjects()` | 无 | `Project[]`(按创建顺序) | 存储读写失败 |
| `AddProject(path, name, note)` | path 必须是已存在目录;name 空→取 basename | `Project` | `目录不存在: <abs>`、`该项目已在列表中`、`路径不能为空` |
| `UpdateProject(project)` | 按 ID 更新 Name/Note/Path(空字段沿用旧值,ID/统计字段不变) | `Project` | `项目不存在: <id>`、`该项目已在列表中` |
| `DeleteProject(id, deleteFiles)` | deleteFiles=true 时同时删除磁盘项目目录(**根目录/用户主目录被护栏拒绝**;磁盘删除失败报错并保留列表项) | `void` | `项目不存在: <id>`、`拒绝删除根目录`、`拒绝删除用户主目录: <path>`、`删除目录失败: …` |
| `OpenProject(id, editorID)` | editorID 为空用默认编辑器;成功启动才更新 lastOpenedAt/openCount;**kind=cli 时打开默认终端到项目目录并在其中执行该命令(默认无参数,codex/claude/grok 等 CLI 以 cwd 为工作目录)** | `void` | `项目不存在: <id>`、`目录不存在: <path>`、`未找到可用编辑器,请先在设置中扫描或添加`、`编辑器不存在: <id>`、`未找到可用终端,请在设置中扫描或添加` |
| `CreateProject(name, note, parentPath)` | mkdirAll parent/name;parentPath 空→Settings.defaultPath | `Project` | `名称不能为空`、`名称不能包含路径分隔符`、`未设置默认目录,请先在设置中配置默认路径`、`该项目已在列表中`、`目标目录已存在: <path>` |
| `CreateTempProject(name, note)` | name 空→自动生成 `tmp-YYYYMMDD-HHMMSS`(碰撞追加 `-2`/`-3`…);给了 name 则不允许路径分隔符、已存在报错;在 `<defaultPath>/tmp` 下,IsTemp=true,绝不自动删除 | `Project` | `未设置默认目录,请先在设置中配置默认路径`、`名称不能包含路径分隔符`、`目标目录已存在: <path>` |

## EditorService(editorservice.js)

> kind 取值:`editor`(GUI 编辑器)/ `terminal`(终端)/ `cli`(CLI 编码工具:codex、claude、gemini、grok、aider、cursor-agent、qwen、opencode、crush,扫描时在 PATH 里命中即收录,默认无参数)。

| 方法 | 参数 | 返回 | 错误 |
| --- | --- | --- | --- |
| `ListEditors()` | 无 | `Editor[]` | 存储读写失败 |
| `ScanEditors()` | 无;全量重扫:保留 manual;同名 ID 保留 Enabled/Args 人工改动;消失的 scan 项删除;**已有项保持既有相对顺序(拖拽排序不被重扫打乱),新扫描项追加在已有项之后** | `Editor[]` | `扫描编辑器失败: …` |
| `AddManualEditor(name, command, args, kind)` | kind ∈ editor/terminal/cli;args 空则默认 `{path}`(**cli 类型例外:args 空保持空,工具以终端 cwd 即项目目录为工作目录**) | `Editor`(ID 为 `m-`+8hex) | `名称不能为空`、`命令不能为空`、`类别无效: <kind>` |
| `UpdateEditor(editor)` | 按 ID 更新(空字段沿用旧值) | `Editor` | `编辑器不存在: <id>` |
| `ReorderEditors(ids)` | 拖拽排序落盘;ids 必须恰好覆盖全部现有 ID 且不重复 | `void` | `ID 数量与编辑器数量不符: <n> != <m>`、`编辑器不存在: <id>`、`ID 重复: <id>` |
| `DeleteEditor(id)` | 编辑器 ID | `void` | `编辑器不存在: <id>` |
| `SetDefaultEditor(id)` | `""` 清除;需已存在 | `void` | `编辑器不存在: <id>` |
| `SetDefaultTerminal(id)` | `""` 清除;需已存在 | `void` | `编辑器不存在: <id>` |

## SettingsService(settingsservice.js)

| 方法 | 参数 | 返回 | 错误 |
| --- | --- | --- | --- |
| `GetSettings()` | 无 | `Settings` | 存储读写失败 |
| `SaveSettings(settings)` | defaultPath 不存在则 MkdirAll,存绝对路径 | `Settings` | `创建目录失败: …` |
| `PickDirectory()` | 无;原生目录选择框 | `string`,取消返回 `""`(不报错) | `打开目录选择框失败: …` |

## 调用示例

```js
// 列出项目
const projects = await Projects.ListProjects();

// 打开项目(用默认编辑器)
await Projects.OpenProject(projects[0].id, "");

// 创建临时项目(自动命名)
const tmp = await Projects.CreateTempProject("", "试验用");
// 创建临时项目(自定义名称)
const tmp2 = await Projects.CreateTempProject("my-sandbox", "");

// 设置默认编辑器
await Editors.SetDefaultEditor("vscode");
```

## 启动行为(main.go)

- 窗口「项目快开」980×640,min 720×560,背景 RGB(11,12,15)。
- `bootstrap()` 在窗口创建前执行:补齐 `projects.json`/`editors.json`/`settings.json`;默认路径 `$HOME/Projects`(自动创建);**仅首次初始化**时自动挑默认:编辑器优先 `code`(VS Code 的命令名)否则第一个 enabled 编辑器,终端优先 `konsole` 否则第一个 enabled 终端;用户显式清除后重启不回填。
- **每次启动自动重新扫描并合并**:新装的编辑器/CLI 工具自动出现在列表,卸载的自动消失;已有项的相对顺序、启用开关、手动项全部保留(合并失败不阻断启动,下次启动再试)。
