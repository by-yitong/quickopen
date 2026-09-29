# 任务卡:quickopen 前端(React + TypeScript + Tailwind v4)

## 背景与目标
「项目快开」桌面工具的前端:管理项目列表并一键用编辑器/终端打开;新建项目(含临时项目);设置页管理编辑器与默认路径。后端 Go 服务与 TS 绑定已由主 agent 完成并生成,你只做 `frontend/` 内的 UI。

## 必读(开工前逐个读完)
1. `/home/admins/.zcode/workspace/default/quickopen/design.md` —— 设计系统(tokens/布局/动效/图标/反模式),**所有视觉决策以它为准**。
2. `/home/admins/.agents/skills/react-best-practices/SKILL.md` —— React 性能规则。
3. `/home/admins/.agents/skills/tailwind-v4/SKILL.md` —— Tailwind v4 语法(`@theme`、CSS-first 配置)。
4. 生成的绑定(已生成好,直接用):`frontend/bindings/quickopen/` 下 `projectservice.js`、`editorservice.js`、`settingsservice.js`(函数式导出,返回 `CancellablePromise`)、`models.js`(`Project`/`Editor`/`Settings` class,带 JSDoc 类型)、`index.js`。tsconfig 需加 `"allowJs": true`(tsconfig.json 允许改这一处),这样 JSDoc 类型会流入 .ts;import 时带 `.js` 后缀,如 `import * as Projects from "../../bindings/quickopen/projectservice.js"`(从 src/lib 出发的相对层级自行计算)。

## 数据契约(后端已实现并冻结;字段为 json 名)
```ts
interface Project { id: string; name: string; note: string; path: string; isTemp: boolean;
  createdAt: number; lastOpenedAt: number; openCount: number } // 时间为 unix 毫秒,0=从未打开
interface Editor { id: string; name: string; icon: string; command: string; args: string;
  kind: "editor" | "terminal"; source: "scan" | "manual"; enabled: boolean; inTerminal: boolean }
interface Settings { defaultPath: string; defaultEditorId: string; defaultTerminalId: string }
```
可用方法(生成物在 `frontend/bindings/quickopen/` 下,与下同名):
ProjectService(projectservice.js): ListProjects() / AddProject(path,name,note) / UpdateProject(project) / DeleteProject(id) / OpenProject(id,editorID) / CreateProject(name,note,parentPath) / CreateTempProject(name,note)
EditorService(editorservice.js): ListEditors() / ScanEditors() / AddManualEditor(name,command,args,kind) / UpdateEditor(editor) / DeleteEditor(id) / SetDefaultEditor(id) / SetDefaultTerminal(id)
SettingsService(settingsservice.js): GetSettings() / SaveSettings(settings) / PickDirectory()  // PickDirectory 取消返回 ""
约定:OpenProject/CreateProject/CreateTempProject/UpdateProject 等返回值或错误按生成物实际类型使用;所有调用失败必须 toast 错误消息(中文,后端返回什么显示什么)。

## 仓库与允许改动范围
- 根:`/home/admins/.zcode/workspace/default/quickopen/frontend`
- 可改/可建:`frontend/` 下任意文件;依赖只允许新增 `tailwindcss`、`@tailwindcss/vite`(npm install 已跑过,补装这两个即可)。禁止动 `bindings/` 生成物、`../main.go`、`build/`。
- 初始化:模板自带 `public/style.css`、`public/bg-desktop.jpg`、`bg-mobile.jpg`、`App.tsx` 演示代码 —— 全部清理,`index.html` 只留必要引用。

## 页面与交互(必须完整实现)

### 视图框架
- 三个视图:`projects` / `create` / `settings`,App 级状态切换;顶部大标题随视图变化;底部悬浮 Dock(见 design.md):左「项目」「设置」两项导航 + 右侧圆形 accent FAB「+」进新建视图。
- 初始加载:并行拉 ListProjects/ListEditors/GetSettings(Promise.all);加载中显示 3 张骨架卡(脉冲动画);失败整页错误态 + 重试按钮。
- Toast:顶部居中滑入,success/error 两态,2.4s 自动消失,可叠 1 条(新的替换旧的)。

### 视图 1:项目列表
- 头部:大标题「项目快开」+ 副标题(如 `6 个项目 · 2 个临时`)。
- 筛选行:胶囊 pills「全部」「临时」+ 胶囊搜索框(按名称/备注/路径过滤,`useDeferredValue` 优化)。
- ProjectCard(design.md 卡片规格):状态行(临时项目→accent 淡底徽章「临时」;普通项目→圆点+「打开 N 次」,从未打开显示「未打开过」)+ 右侧相对时间(lastOpenedAt,从未打开用 createdAt,格式:刚刚/N 分钟前/N 小时前/N 天前/超过 7 天显示 YYYY-MM-DD)→ 名称 17px bold → 备注(有才显示)→ 路径 mono 12px truncate(hover 显示 title 完整路径)→ 底部图标行:终端图标(用 settings.defaultTerminalId 对应项,没有则不显示)+ 所有 enabled 且 kind=editor 的编辑器图标(顺序稳定,vscode 类靠前不强制),点击图标=用该编辑器打开;最右「⋯」菜单:编辑信息 / 复制路径 / 删除。
- 点击卡片主体(非图标区)= 用默认编辑器打开;打开中图标转圈,失败 toast。
- 编辑弹层(Mod):名称/备注/路径三个字段,路径带「更改」按钮调 PickDirectory;保存调 UpdateProject;删除需二次确认弹层(danger 按钮,文案「删除后列表不再显示,磁盘目录不受影响」)。
- 列表排序:lastOpenedAt 降序(0 排最后按 createdAt);列表尾固定「没有更多了」(ink-3 居中 13px)。空态:「还没有项目,点下方 + 创建一个」;搜索无结果:「没有匹配的项目」。
- 复制路径:clipboard API + toast「已复制」。

### 视图 2:新建
- 大标题「新建项目」。
- 表单:名称 input(placeholder「项目名称」;校验:非空用于普通创建、不含 / \ )+ 备注 input(可选)。
- 路径行:mono 显示父目录(settings.defaultPath,可被用户改)+「更改」按钮(PickDirectory)+ 下方实时预览完整路径 `{parent}/{name}`(ink-3 12px)。
- 操作区 A:「仅创建」次级按钮(校验名称非空)→ CreateProject → toast「已创建」→ 跳回列表(新项目在顶部可见)。
- 操作区 B:分隔线 + 标签「创建并直接打开」→ 编辑器图标行(全部 enabled 编辑器),点击图标 → CreateProject → OpenProject(id, editor.id) → toast「已创建并打开」/失败 toast 并回列表(项目已登记)。
- 操作区 C(临时项目):卡片内底部分隔块,说明文案:「临时目录建在 {defaultPath}/tmp/ 下,不会自动删除」;一个 accent 描边按钮「🎲 随机临时目录并打开」:名称空 → 直接 CreateTempProject + OpenProject(默认编辑器);名称已填 → 用该名称作为 tmp 下的目录名。同预览路径。临时项目也可只创建不打开:按住 alt?不要——给个次要文字按钮「仅创建临时目录」。
- 表单校验错误:input 下方 12px danger 文案,不弹窗。

### 视图 3:设置
- 大标题「设置」。
- 区块 1「默认路径」:mono 显示 current defaultPath + 「更改」(PickDirectory→SaveSettings)+ 说明文字。
- 区块 2「默认应用」:两个下拉(原生 select 样式化):默认编辑器(enabled 且 kind=editor 的列表)、默认终端(kind=terminal);改动即 SetDefaultEditor/SetDefaultTerminal + toast。
- 区块 3「编辑器管理」:头部:「重新扫描」按钮(loading 态转圈,ScanEditors 后 toast「扫描完成,共 N 个」)+「手动添加」按钮。列表行:图标 + 名称 + command(mono 12px truncate)+ 徽章(编辑器/终端;InTerminal 加「终端内」;来源「扫描」/「手动」)+ Enabled 开关(乐观更新,失败回滚 + toast)+ 删除按钮(二次确认)。
- 手动添加弹层:名称/命令/参数(默认 `{path}`)/类型单选(编辑器、终端)→ AddManualEditor;命令校验非空。

### EditorIcon 组件
- `icon` slug → 内联 SVG(16-18px,`fill="currentColor"`,单色几何近似即可):vscode、cursor、zed、sublime、jetbrains、vim、nvim、emacs、helix、windsurf、vscodium、trae、terminal、generic;未知 slug → generic(圆角方块 + 首字母)。SVG path 自己画,禁止外链图片;图标写在 `src/components/icons.tsx` 一个文件里。

## React 性能纪律(硬性)
- 组件内不得定义组件(全提文件级);ProjectCard 用 `memo`,handlers 用 `useCallback` + 函数式 setState;effect 依赖只放原始值;条件渲染用三元;过滤/排序 useMemo;不用任何状态库/动画库;不建 barrel 文件。
- 可交互元素 8 态齐全(design.md),`focus-visible` 外环必须有;Esc 关弹层;弹层打开锁 body 滚动。
- 全程无 `transition-all`;动效只用 transform/opacity;`prefers-reduced-motion: reduce` 时入场动画退化为 opacity。

## 验收标准
1. `cd frontend && npm run build`(tsc + vite build)零错误零警告。
2. 无内联裸色值:颜色一律来自 `@theme` token(design.md 表);无绿/紫任意色相。
3. 交互完整覆盖上面三视图所有分支(含空态/错误态/加载态)。
4. 交付报告:文件清单、`npm run build` 输出原文、与 design.md 的任何偏差及原因、遗留疑问。

## 禁止事项
- 不得改 `bindings/`、Go 代码、`build/`;不得引入任务卡外依赖;不得再派生子 agent;卡外问题写报告带回。
- 不得用 emoji 作界面图标(「🎲」按钮文字里的骰子除外,它算文案);不得使用 alert/confirm 原生弹窗(统一自制弹层)。
