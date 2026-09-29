import {
  useCallback,
  useMemo,
  useRef,
  useState,
  type Dispatch,
  type DragEvent,
  type SetStateAction,
} from "react";
import { EditorKind, type Editor, type Settings } from "../../bindings/quickopen/models.js";
import {
  AddManualEditor,
  DeleteEditor,
  ReorderEditors,
  ScanEditors,
  SetDefaultEditor,
  SetDefaultTerminal,
  UpdateEditor,
} from "../../bindings/quickopen/editorservice.js";
import { PickDirectory, SaveSettings } from "../../bindings/quickopen/settingsservice.js";
import { errMsg } from "../lib/format";
import { Button, ConfirmDialog, FieldError, IconButton, Modal, Switch, TextInput, cx, INPUT_CLS } from "./ui";
import { ChevronDownIcon, EditorIcon, GripIcon, PlusIcon, TrashIcon } from "./icons";
import type { ToastKind } from "./Toast";

/** 视图 3:设置(默认路径 / 默认应用 / 编辑器管理)。 */

interface SettingsViewProps {
  settings: Settings;
  editors: Editor[];
  setSettings: Dispatch<SetStateAction<Settings | null>>;
  setEditors: Dispatch<SetStateAction<Editor[]>>;
  pushToast: (kind: ToastKind, message: string) => void;
}

export function SettingsView({ settings, editors, setSettings, setEditors, pushToast }: SettingsViewProps) {
  const [pathBusy, setPathBusy] = useState(false);
  const [scanning, setScanning] = useState(false);
  const [showAdd, setShowAdd] = useState(false);
  const [deleting, setDeleting] = useState<Editor | null>(null);

  const editorOptions = useMemo(() => editors.filter((e) => e.enabled && e.kind === "editor"), [editors]);
  const terminalOptions = useMemo(() => editors.filter((e) => e.kind === "terminal"), [editors]);

  const changePath = useCallback(() => {
    if (pathBusy) return;
    setPathBusy(true);
    PickDirectory()
      .then((dir) => {
        if (!dir) return; // 用户取消
        return SaveSettings({ ...settings, defaultPath: dir }).then((saved) => {
          setSettings(saved);
          pushToast("success", "默认路径已更新");
        });
      })
      .catch((err: unknown) => pushToast("error", errMsg(err)))
      .finally(() => setPathBusy(false));
  }, [pathBusy, settings, setSettings, pushToast]);

  const handleDefaultEditor = useCallback(
    (id: string) => {
      SetDefaultEditor(id)
        .then(() => {
          setSettings((prev) => (prev ? { ...prev, defaultEditorId: id } : prev));
          pushToast("success", id ? "已设为默认编辑器" : "已清除默认编辑器");
        })
        .catch((err: unknown) => pushToast("error", errMsg(err)));
    },
    [setSettings, pushToast],
  );

  const handleDefaultTerminal = useCallback(
    (id: string) => {
      SetDefaultTerminal(id)
        .then(() => {
          setSettings((prev) => (prev ? { ...prev, defaultTerminalId: id } : prev));
          pushToast("success", id ? "已设为默认终端" : "已清除默认终端");
        })
        .catch((err: unknown) => pushToast("error", errMsg(err)));
    },
    [setSettings, pushToast],
  );

  const rescan = useCallback(() => {
    if (scanning) return;
    setScanning(true);
    ScanEditors()
      .then((list) => {
        const found = list ?? [];
        setEditors(found);
        pushToast("success", "扫描完成,共 " + found.length + " 个");
      })
      .catch((err: unknown) => pushToast("error", errMsg(err)))
      .finally(() => setScanning(false));
  }, [scanning, setEditors, pushToast]);

  const toggleEnabled = useCallback(
    (ed: Editor) => {
      const next = { ...ed, enabled: !ed.enabled };
      setEditors((prev) => prev.map((e) => (e.id === ed.id ? next : e))); // 乐观更新
      UpdateEditor(next).then(
        (saved) => setEditors((prev) => prev.map((e) => (e.id === ed.id ? saved : e))),
        (err: unknown) => {
          setEditors((prev) => prev.map((e) => (e.id === ed.id ? ed : e))); // 失败回滚
          pushToast("error", errMsg(err));
        },
      );
    },
    [setEditors, pushToast],
  );

  const closeAdd = useCallback(() => setShowAdd(false), []);
  const closeDelete = useCallback(() => setDeleting(null), []);

  /* ---- 拖拽排序(原生 HTML5 DnD,无第三方库) ---- */
  const dragIdRef = useRef<string | null>(null);
  const [dragId, setDragId] = useState<string | null>(null);
  const [overId, setOverId] = useState<string | null>(null);

  const handleDragStart = useCallback((ed: Editor) => (e: DragEvent) => {
    dragIdRef.current = ed.id;
    setDragId(ed.id);
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", ed.id);
  }, []);

  const handleDragOver = useCallback((ed: Editor) => (e: DragEvent) => {
    if (dragIdRef.current === null || dragIdRef.current === ed.id) return;
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    setOverId(ed.id);
  }, []);

  const handleDrop = useCallback(
    (ed: Editor) => (e: DragEvent) => {
      e.preventDefault();
      const fromId = dragIdRef.current;
      if (fromId === null || fromId === ed.id) return;
      const fromIdx = editors.findIndex((x) => x.id === fromId);
      const toIdx = editors.findIndex((x) => x.id === ed.id);
      if (fromIdx < 0 || toIdx < 0) return;
      const next = editors.slice();
      const [moved] = next.splice(fromIdx, 1);
      next.splice(toIdx, 0, moved);
      setEditors(next); // 乐观更新
      ReorderEditors(next.map((x) => x.id)).catch((err: unknown) => {
        setEditors(editors); // 失败回滚到拖拽前
        pushToast("error", errMsg(err));
      });
    },
    [editors, setEditors, pushToast],
  );

  const handleDragEnd = useCallback(() => {
    dragIdRef.current = null;
    setDragId(null);
    setOverId(null);
  }, []);

  const confirmDelete = useCallback(() => {
    if (!deleting) return Promise.resolve();
    return DeleteEditor(deleting.id).then(
      () => {
        setEditors((prev) => prev.filter((e) => e.id !== deleting.id));
        pushToast("success", "已删除");
      },
      (err: unknown) => {
        pushToast("error", errMsg(err));
        throw err; // 弹层据此外抛保持打开
      },
    );
  }, [deleting, setEditors, pushToast]);

  return (
    <div className="anim-view space-y-4">
      {/* 区块 1:默认路径 */}
      <section className="rounded-card border border-line bg-surface p-5">
        <h2 className="text-[15px] font-semibold text-ink">默认路径</h2>
        <div className="mt-3 flex items-center gap-2">
          <code
            className="min-w-0 flex-1 truncate rounded-input border border-line bg-surface-2 px-3.5 py-2.5 font-mono text-[13px] text-ink-2"
            title={settings.defaultPath}
          >
            {settings.defaultPath || "未设置"}
          </code>
          <Button onClick={changePath} loading={pathBusy} className="shrink-0">
            更改
          </Button>
        </div>
        <p className="mt-2 text-[12px] text-ink-3">新项目默认创建在该目录下;临时项目建在其 tmp/ 子目录。</p>
      </section>

      {/* 区块 2:默认应用 */}
      <section className="rounded-card border border-line bg-surface p-5">
        <h2 className="text-[15px] font-semibold text-ink">默认应用</h2>
        <div className="mt-3 grid gap-3">
          <label className="block">
            <span className="mb-1.5 block text-[13px] font-medium text-ink-2">默认编辑器</span>
            <Select
              value={settings.defaultEditorId}
              onChange={handleDefaultEditor}
              options={editorOptions}
              currentId={settings.defaultEditorId}
              all={editors}
            />
          </label>
          <label className="block">
            <span className="mb-1.5 block text-[13px] font-medium text-ink-2">默认终端</span>
            <Select
              value={settings.defaultTerminalId}
              onChange={handleDefaultTerminal}
              options={terminalOptions}
              currentId={settings.defaultTerminalId}
              all={editors}
            />
          </label>
        </div>
      </section>

      {/* 区块 3:编辑器管理 */}
      <section className="rounded-card border border-line bg-surface p-5">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-[15px] font-semibold text-ink">编辑器管理</h2>
          <div className="flex items-center gap-2">
            <Button onClick={rescan} loading={scanning} disabled={showAdd || deleting !== null}>
              重新扫描
            </Button>
            <Button onClick={() => setShowAdd(true)} disabled={scanning || deleting !== null}>
              <PlusIcon className="size-4" />
              手动添加
            </Button>
          </div>
        </div>
        {editors.length === 0 ? (
          <p className="mt-4 text-[13px] text-ink-3">未发现编辑器,点「重新扫描」或「手动添加」</p>
        ) : (
          <>
            <p className="mt-3 text-[12px] text-ink-3">拖动 ⋮⋮ 调整顺序,顺序即项目卡片上图标行的顺序。</p>
            <ul className="mt-1 divide-y divide-line">
              {editors.map((ed) => (
                <EditorRow
                  key={ed.id}
                  editor={ed}
                  dragging={dragId === ed.id}
                  dropHint={overId === ed.id && dragId !== null && dragId !== ed.id}
                  onToggle={toggleEnabled}
                  onDelete={setDeleting}
                  onDragStart={handleDragStart(ed)}
                  onDragOver={handleDragOver(ed)}
                  onDrop={handleDrop(ed)}
                  onDragEnd={handleDragEnd}
                />
              ))}
            </ul>
          </>
        )}
      </section>

      {showAdd ? <AddEditorModal onClose={closeAdd} setEditors={setEditors} pushToast={pushToast} /> : null}
      {deleting ? (
        <ConfirmDialog
          title={"删除 " + deleting.name}
          body="删除后不再出现在打开图标与默认应用中,该操作不可撤销。"
          confirmLabel="删除"
          onClose={closeDelete}
          onConfirm={confirmDelete}
        />
      ) : null}
    </div>
  );
}

/* ---- 原生 select(样式化) ---- */

interface SelectProps {
  value: string;
  options: Editor[];
  currentId: string;
  all: Editor[];
  onChange: (id: string) => void;
}

function Select({ value, options, currentId, all, onChange }: SelectProps) {
  const currentMissing = currentId !== "" && !options.some((o) => o.id === currentId);
  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={cx(INPUT_CLS, "appearance-none pr-10")}
      >
        <option value="">未设置</option>
        {options.map((o) => (
          <option key={o.id} value={o.id}>
            {o.name}
          </option>
        ))}
        {currentMissing ? (
          <option value={currentId}>{all.find((o) => o.id === currentId)?.name ?? currentId}(已禁用)</option>
        ) : null}
      </select>
      <ChevronDownIcon className="pointer-events-none absolute right-3.5 top-1/2 size-4 -translate-y-1/2 text-ink-3" />
    </div>
  );
}

/* ---- 编辑器列表行 ---- */

function EditorRow({
  editor,
  dragging,
  dropHint,
  onToggle,
  onDelete,
  onDragStart,
  onDragOver,
  onDrop,
  onDragEnd,
}: {
  editor: Editor;
  dragging: boolean;
  dropHint: boolean;
  onToggle: (ed: Editor) => void;
  onDelete: (ed: Editor) => void;
  onDragStart: (e: DragEvent) => void;
  onDragOver: (e: DragEvent) => void;
  onDrop: (e: DragEvent) => void;
  onDragEnd: () => void;
}) {
  return (
    <li
      draggable
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDrop={onDrop}
      onDragEnd={onDragEnd}
      className={cx(
        "relative flex items-center gap-3 py-2.5 transition-opacity duration-[var(--dur-fast)] ease-out",
        dragging && "opacity-40",
        dropHint && "shadow-[0_-2px_0_0_var(--color-accent)]",
      )}
    >
      <span className="cursor-grab text-ink-3 active:cursor-grabbing" title="拖拽排序">
        <GripIcon className="size-4" />
      </span>
      <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-surface-2 text-ink-2">
        <EditorIcon slug={editor.icon} className="size-[18px]" />
      </span>
      <div className="min-w-0 flex-1">
        <div className="flex min-w-0 items-center gap-2">
          <span className={cx("truncate text-[14px] font-medium", editor.enabled ? "text-ink" : "text-ink-3")}>
            {editor.name}
          </span>
          <Badge>{editor.kind === "editor" ? "编辑器" : editor.kind === "cli" ? "CLI" : "终端"}</Badge>
          {editor.inTerminal ? <Badge>终端内</Badge> : null}
          <Badge>{editor.source === "scan" ? "扫描" : "手动"}</Badge>
        </div>
        <p className="truncate font-mono text-[12px] text-ink-3" title={editor.command + " " + editor.args}>
          {editor.command} {editor.args}
        </p>
      </div>
      <Switch checked={editor.enabled} label={"启用 " + editor.name} onChange={() => onToggle(editor)} />
      <IconButton label={"删除 " + editor.name} variant="danger-ghost" onClick={() => onDelete(editor)}>
        <TrashIcon className="size-4" />
      </IconButton>
    </li>
  );
}

function Badge({ children }: { children: string }) {
  return (
    <span className="shrink-0 rounded-pill border border-line px-2 py-px text-[11px] text-ink-2">{children}</span>
  );
}

/* ---- 手动添加弹层 ---- */

interface AddEditorModalProps {
  onClose: () => void;
  setEditors: Dispatch<SetStateAction<Editor[]>>;
  pushToast: (kind: ToastKind, message: string) => void;
}

function AddEditorModal({ onClose, setEditors, pushToast }: AddEditorModalProps) {
  const [name, setName] = useState("");
  const [command, setCommand] = useState("");
  const [args, setArgs] = useState("{path}");
  const [kind, setKind] = useState<"editor" | "terminal" | "cli">("editor");
  const [nameError, setNameError] = useState<string | null>(null);
  const [commandError, setCommandError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const changeKind = useCallback(
    (next: "editor" | "terminal" | "cli") => {
      setKind(next);
      // CLI 工具默认不追加参数(以终端 cwd 即项目目录为工作目录)
      if (next === "cli" && args.trim() === "{path}") setArgs("");
      if (next !== "cli" && args.trim() === "") setArgs("{path}");
    },
    [args],
  );

  const submit = useCallback(() => {
    if (saving) return;
    const nErr = name.trim() === "" ? "名称不能为空" : null;
    const cErr = command.trim() === "" ? "命令不能为空" : null;
    setNameError(nErr);
    setCommandError(cErr);
    if (nErr || cErr) return;
    setSaving(true);
    const kindValue =
      kind === "editor" ? EditorKind.KindEditor : kind === "terminal" ? EditorKind.KindTerminal : EditorKind.KindCLI;
    AddManualEditor(name.trim(), command.trim(), args.trim(), kindValue)
      .then((created) => {
        setEditors((prev) => [...prev, created]);
        pushToast("success", "已添加");
        onClose();
      })
      .catch((err: unknown) => {
        pushToast("error", errMsg(err));
        setSaving(false);
      });
  }, [saving, name, command, args, kind, setEditors, pushToast, onClose]);

  return (
    <Modal title="手动添加" onClose={onClose}>
      <form
        className="space-y-4"
        onSubmit={(e) => {
          e.preventDefault();
          submit();
        }}
      >
        <label className="block">
          <span className="mb-1.5 block text-[13px] font-medium text-ink-2">名称</span>
          <TextInput value={name} onChange={(e) => setName(e.target.value)} placeholder="如 Fleet" />
          {nameError ? <FieldError message={nameError} /> : null}
        </label>
        <label className="block">
          <span className="mb-1.5 block text-[13px] font-medium text-ink-2">命令</span>
          <TextInput value={command} onChange={(e) => setCommand(e.target.value)} placeholder="PATH 中的可执行名或绝对路径" />
          {commandError ? <FieldError message={commandError} /> : null}
        </label>
        <label className="block">
          <span className="mb-1.5 block text-[13px] font-medium text-ink-2">参数</span>
          <TextInput
            value={args}
            onChange={(e) => setArgs(e.target.value)}
            placeholder={kind === "cli" ? "留空 = 不追加参数(工具用项目目录)" : "{path}"}
          />
        </label>
        <div>
          <span className="mb-1.5 block text-[13px] font-medium text-ink-2">类型</span>
          <div className="flex gap-2" role="radiogroup" aria-label="类型">
            <KindOption label="编辑器" selected={kind === "editor"} onClick={() => changeKind("editor")} />
            <KindOption label="终端" selected={kind === "terminal"} onClick={() => changeKind("terminal")} />
            <KindOption label="CLI 工具" selected={kind === "cli"} onClick={() => changeKind("cli")} />
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-1">
          <Button onClick={onClose}>取消</Button>
          <Button variant="accent-outline" loading={saving} onClick={submit}>
            添加
          </Button>
        </div>
      </form>
    </Modal>
  );
}

function KindOption({ label, selected, onClick }: { label: string; selected: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      role="radio"
      aria-checked={selected}
      onClick={onClick}
      className={cx(
        "flex-1 rounded-input border px-3.5 py-2.5 text-[13px] font-medium transition-colors duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.98]",
        selected ? "border-accent/60 bg-accent-dim text-accent" : "border-line bg-surface-2 text-ink-2 hover:text-ink",
      )}
    >
      {label}
    </button>
  );
}
