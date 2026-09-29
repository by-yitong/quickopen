import {
  useCallback,
  useDeferredValue,
  useMemo,
  useState,
  type FormEvent,
} from "react";
import type { Editor, Project } from "../../bindings/quickopen/models.js";
import { PickDirectory } from "../../bindings/quickopen/settingsservice.js";
import { errMsg, validateName, copyText } from "../lib/format";
import { Button, ConfirmDialog, FieldError, Modal, TextInput, cx } from "./ui";
import { ProjectCard } from "./ProjectCard";
import { SearchIcon } from "./icons";
import type { ToastKind } from "./Toast";

/** 视图 1:项目列表(筛选/搜索/卡片/编辑弹层/删除确认)。 */

interface ProjectsViewProps {
  projects: Project[];
  editors: Editor[];
  defaultEditorId: string;
  defaultTerminalId: string;
  openingKey: string | null;
  pushToast: (kind: ToastKind, message: string) => void;
  onOpenWith: (projectId: string, editorId: string) => void;
  onOpenDefault: (projectId: string) => void;
  onUpdateProject: (project: Project) => Promise<void>;
  onDeleteProject: (id: string) => Promise<void>;
}

type Filter = "all" | "temp";

export function ProjectsView({
  projects,
  editors,
  defaultEditorId,
  defaultTerminalId,
  openingKey,
  pushToast,
  onOpenWith,
  onOpenDefault,
  onUpdateProject,
  onDeleteProject,
}: ProjectsViewProps) {
  const [filter, setFilter] = useState<Filter>("all");
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const [editing, setEditing] = useState<Project | null>(null);
  const [deleting, setDeleting] = useState<Project | null>(null);

  const launchers = useMemo(
    () => editors.filter((e) => e.enabled && (e.kind === "editor" || e.kind === "cli")),
    [editors],
  );
  const terminal = useMemo(
    () => (defaultTerminalId ? editors.find((e) => e.id === defaultTerminalId && e.enabled) ?? null : null),
    [editors, defaultTerminalId],
  );

  const visible = useMemo(() => {
    let list = filter === "temp" ? projects.filter((p) => p.isTemp) : projects;
    const q = deferredQuery.trim().toLowerCase();
    if (q) {
      list = list.filter(
        (p) => p.name.toLowerCase().includes(q) || p.note.toLowerCase().includes(q) || p.path.toLowerCase().includes(q),
      );
    }
    return [...list].sort((a, b) => {
      if (a.lastOpenedAt && b.lastOpenedAt) return b.lastOpenedAt - a.lastOpenedAt;
      if (!a.lastOpenedAt && !b.lastOpenedAt) return b.createdAt - a.createdAt;
      return a.lastOpenedAt ? -1 : 1;
    });
  }, [projects, filter, deferredQuery]);

  const closeEdit = useCallback(() => setEditing(null), []);
  const closeDelete = useCallback(() => setDeleting(null), []);
  const handleEdit = useCallback((p: Project) => setEditing(p), []);
  const handleDelete = useCallback((p: Project) => setDeleting(p), []);

  const handleCopy = useCallback(
    (p: Project) => {
      copyText(p.path).then(
        () => pushToast("success", "已复制"),
        () => pushToast("error", "复制失败"),
      );
    },
    [pushToast],
  );

  const confirmDelete = useCallback(() => {
    if (!deleting) return Promise.resolve();
    return onDeleteProject(deleting.id);
  }, [deleting, onDeleteProject]);

  const isEmpty = projects.length === 0;
  const noMatch = visible.length === 0;

  return (
    <div className="space-y-4">
      {/* 筛选行 */}
      <div className="flex items-center gap-2">
        <FilterPill label="全部" active={filter === "all"} onClick={() => setFilter("all")} />
        <FilterPill label="临时" active={filter === "temp"} onClick={() => setFilter("temp")} />
        <div className="relative flex-1">
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="搜索名称、备注、路径"
            aria-label="搜索项目"
            className="w-full rounded-pill border border-line bg-surface px-4 py-2 pr-10 text-[13px] text-ink transition-colors duration-[var(--dur-fast)] outline-offset-2 placeholder:text-ink-3 focus:border-accent focus-visible:outline-2 focus-visible:outline-accent"
          />
          <SearchIcon className="pointer-events-none absolute right-3.5 top-1/2 size-4 -translate-y-1/2 text-ink-3" />
        </div>
      </div>

      {/* 列表 */}
      {isEmpty ? (
        <div className="anim-enter rounded-card border border-dashed border-line py-16 text-center">
          <p className="text-[14px] text-ink-2">还没有项目,点下方 + 创建一个</p>
        </div>
      ) : noMatch ? (
        <div className="anim-enter rounded-card border border-dashed border-line py-16 text-center">
          <p className="text-[14px] text-ink-2">没有匹配的项目</p>
        </div>
      ) : (
        <div className="space-y-2">
          {visible.map((p, i) => {
            const openingEditorId =
              openingKey !== null && openingKey.startsWith(p.id + "|") ? openingKey.slice(p.id.length + 1) : null;
            return (
              <ProjectCard
                key={p.id}
                project={p}
                launchers={launchers}
                terminal={terminal}
                openingEditorId={openingEditorId}
                onOpenWith={onOpenWith}
                onOpenDefault={onOpenDefault}
                onEdit={handleEdit}
                onCopy={handleCopy}
                onDelete={handleDelete}
                delay={i}
              />
            );
          })}
          <p className="pt-1 text-center text-[13px] text-ink-3">没有更多了</p>
        </div>
      )}

      {/* 弹层 */}
      {editing ? (
        <EditProjectModal project={editing} onClose={closeEdit} pushToast={pushToast} onSaved={onUpdateProject} />
      ) : null}
      {deleting ? (
        <ConfirmDialog
          title={"删除 " + deleting.name}
          body="删除后列表不再显示,磁盘目录不受影响"
          confirmLabel="删除"
          onClose={closeDelete}
          onConfirm={confirmDelete}
        />
      ) : null}
    </div>
  );
}

function FilterPill({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      aria-pressed={active}
      onClick={onClick}
      className={cx(
        "shrink-0 rounded-pill px-3.5 py-2 text-[13px] font-medium transition-colors duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.98]",
        active ? "bg-accent-dim text-accent" : "text-ink-2 hover:bg-surface hover:text-ink",
      )}
    >
      {label}
    </button>
  );
}

/* ---- 编辑弹层 ---- */

interface EditProjectModalProps {
  project: Project;
  onClose: () => void;
  pushToast: (kind: ToastKind, message: string) => void;
  onSaved: (edited: Project) => Promise<void>;
}

function EditProjectModal({ project, onClose, pushToast, onSaved }: EditProjectModalProps) {
  const [name, setName] = useState(project.name);
  const [note, setNote] = useState(project.note);
  const [path, setPath] = useState(project.path);
  const [nameError, setNameError] = useState<string | null>(null);
  const [pathError, setPathError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [picking, setPicking] = useState(false);

  const submit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      if (saving) return;
      const nErr = validateName(name, true);
      const pErr = path.trim() === "" ? "路径不能为空" : null;
      setNameError(nErr);
      setPathError(pErr);
      if (nErr || pErr) return;
      setSaving(true);
      onSaved({ ...project, name: name.trim(), note: note.trim(), path: path.trim() })
        .then(onClose)
        .catch((err: unknown) => {
          pushToast("error", errMsg(err));
          setSaving(false);
        });
    },
    [saving, name, note, path, project, onSaved, onClose, pushToast],
  );

  const changePath = useCallback(() => {
    if (picking) return;
    setPicking(true);
    PickDirectory()
      .then((dir) => {
        if (dir) setPath(dir);
      })
      .catch((err: unknown) => pushToast("error", errMsg(err)))
      .finally(() => setPicking(false));
  }, [picking, pushToast]);

  return (
    <Modal title="编辑信息" onClose={onClose}>
      <form className="space-y-4" onSubmit={submit}>
        <label className="block">
          <span className="mb-1.5 block text-[13px] font-medium text-ink-2">名称</span>
          <TextInput value={name} onChange={(e) => setName(e.target.value)} placeholder="项目名称" />
          {nameError ? <FieldError message={nameError} /> : null}
        </label>
        <label className="block">
          <span className="mb-1.5 block text-[13px] font-medium text-ink-2">备注</span>
          <TextInput value={note} onChange={(e) => setNote(e.target.value)} placeholder="可选" />
        </label>
        <div>
          <span className="mb-1.5 block text-[13px] font-medium text-ink-2">路径</span>
          <div className="flex items-center gap-2">
            <code
              className="min-w-0 flex-1 truncate rounded-input border border-line bg-surface-2 px-3.5 py-2.5 font-mono text-[13px] text-ink-2"
              title={path}
            >
              {path}
            </code>
            <Button onClick={changePath} loading={picking} className="shrink-0">
              更改
            </Button>
          </div>
          {pathError ? <FieldError message={pathError} /> : null}
        </div>
        <div className="flex justify-end gap-2 pt-1">
          <Button onClick={onClose}>取消</Button>
          <Button type="submit" variant="accent-outline" loading={saving}>
            保存
          </Button>
        </div>
      </form>
    </Modal>
  );
}
