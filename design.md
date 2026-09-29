# design.md — 项目快开 (QuickOpen) 设计系统

```css
/* Hallmark · macrostructure: Workbench list (app-frame) · tone: utilitarian-dark · anchor hue: 250 (cyan-blue)
 * theme: studied-DNA (source: image) · paper oklch(14% 0.01 260) · accent oklch(65% 0.17 250)
 * display: heavy geometric sans (Inter + 系统 CJK) · body: neutral grotesque (同族) · label/数值: mono
 * studied: yes · DNA-source: image (用户参考截图)
 * 注: 参考图主色为绿色,按用户全局规则(禁绿/紫)换成 cyan-blue,其余 DNA 不变
 */
```

## Provenance

- 提取自用户附带的手机 App 截图(2026-09-28)。Tokens 由截图色带估算。
- 保留的 DNA:深色底 + 亮色卡片抬升、全胶囊圆角、底部悬浮 Dock + 圆形主按钮、大标题 + 胶囊筛选 + 圆角卡片列表、列表尾部"没有更多了"。
- 更换的轴:accent 从绿 → cyan-blue(用户规则:主题色禁绿/紫)。

## Tokens(Tailwind v4 `@theme` + CSS 自定义属性,禁止内联裸值)

| token | 值 | 用途 |
| --- | --- | --- |
| `--color-paper` | `#0b0c0f` (oklch 14% 0.01 260) | 窗口底色 |
| `--color-surface` | `#16171d` | 卡片/输入框/弹层 |
| `--color-surface-2` | `#1e2027` | 卡片 hover、开关轨道 |
| `--color-line` | `rgba(255,255,255,0.07)` | 发丝描边 |
| `--color-accent` | `#3e8eff` (oklch 65% 0.17 250) | 主色:激活态、FAB、链接 |
| `--color-accent-dim` | `rgba(62,142,255,0.14)` | 主色淡底(激活胶囊/徽章底) |
| `--color-danger` | `#ff5d5d` | 删除 |
| `--color-ink` | `#f2f3f5` | 一级文本 |
| `--color-ink-2` | `#9ba0aa` | 二级文本 |
| `--color-ink-3` | `#676c75` | 三级文本/占位 |
| `--font-sans` | `Inter, system-ui, "PingFang SC", "Microsoft YaHei", "Noto Sans SC", sans-serif` | 标题+正文(同一族,靠字重分层) |
| `--font-mono` | `"JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, monospace` | 路径、数值、命令 |
| `--radius-card` | `24px` | 卡片(rounded-3xl) |
| `--radius-pill` | `999px` | 胶囊(筛选/Dock/FAB) |
| `--radius-input` | `14px` | 输入框/按钮次级 |
| `--space-unit` | 4pt 刻度(`space-y-1.5/2/3/4/6` 等) | 间距 |
| `--dur-fast` | `140ms` | hover/press |
| `--dur-view` | `200ms` | 视图切换/弹层 |
| `--ease-out` | `cubic-bezier(0.22, 1, 0.36, 1)` | 唯一缓动 |

## 布局骨架

- 窗口 980×640(min 720×560),内容单列居中 `max-w-2xl px-6`,底部为悬浮 Dock 预留 `pb-28`。
- 三个视图:项目列表 / 新建 / 设置,顶部大标题(34px,font-bold,靠左,`pt-8`)。
- Dock:固定底部居中,`rounded-full bg-surface/95 border border-line backdrop-blur px-3 py-2 flex items-center gap-1`;
  两项导航(项目/设置,图标+11px 小字竖排,激活项 `bg-accent-dim text-accent rounded-full`)+ 右侧 48px 圆形 FAB(`bg-accent` 白色 +)。
- 卡片:项目卡 `bg-surface border border-line rounded-3xl px-4 py-2.5`(紧凑横向,尽量矮):左侧信息列(min-w-0,两行:标题行=名称 15px bold + 临时徽章 + 右侧「打开 N 次/未打开过」与相对时间 11px;下行=路径 mono 11px truncate,备注有才显示在中间 12px;点主体=默认编辑器打开),右侧快捷启动图标行(shrink-0:终端+编辑器图标 28px 按钮/16px 图形,超过 6 个收进「+N」弹层,列表带图标+名称;最右 ⋯ 菜单)。列表卡间距 space-y-2。
- 筛选:标题下胶囊 pills(全部 / 临时)+ 搜索框(胶囊,右侧放大镜)。

## 组件状态纪律(所有可交互元素 8 态)

default · hover · focus-visible(2px accent 外环,不动画)· active(scale .98)· disabled(opacity .45)· loading( spinner/骨架)· error(文案 + danger)· success(静默,toast 一次)。

## 动效(仅 transform/opacity;respect prefers-reduced-motion)

- 列表项入场:fade-up 160ms 依次 stagger 30ms(≤12 项)。
- 卡片按压 scale(0.98);Dock 切换背景胶囊 140ms 滑动。
- 弹层:fade + 8px 上移 200ms;toast:顶部滑入 200ms、2.4s 自动消失。
- 禁止:弹跳/过冲、transition-all、hover 放大卡片。

## 文案语气

按钮用动词:立即创建、重新扫描、更改、保存;空态直白:"还没有项目""没有匹配的项目";列表尾:"没有更多了"。

## 编辑器图标(slug → 内联单色 SVG,16–18px,`currentColor`)

vscode · cursor · zed · sublime · jetbrains · vim · nvim · emacs · helix · windsurf · vscodium · trae · terminal(>_ 形)· generic(圆角方块+首字母)。未知 slug 一律 generic。图标按钮 hover `bg-surface-2`,tooltip 显示编辑器名。

## 反模式(禁止)

绿色/紫色任何色相 · 感illin渐变大横幅 · emoji 当图标 · `transition-all` · 双行可点击文案 · 长列表无 end-of-list 提示 · 弹窗放大动画。
