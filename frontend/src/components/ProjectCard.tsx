import { memo, useState, useCallback, useEffect, useMemo } from "react";
import type { Editor, Project } from "../../bindings/quickopen/models.js";
import { relativeTime } from "../lib/format";
import { cx, Spinner } from "./ui";
import { DotsIcon, EditorIcon } from "./icons";

/**
 * 项目卡:左侧项目信息,右侧快捷启动图标(超出的收进「+N」弹层)。
 * memo + 全部 handler 由父级 useCallback 注入;openingEditorId 仅命中中的卡变化。
 */

/** 图标行最多直接展示的图标数(终端+编辑器合计),超出收进 +N。 */
const MAX_VISIBLE = 6;

interface ProjectCardProps {
  project: Project;
  launchers: Editor[];
  terminal: Editor | null;
  openingEditorId: string | null;
  onOpenWith: (projectId: string, editorId: string) => void;
  onOpenDefault: (projectId: string) => void;
  onEdit: (project: Project) => void;
  onCopy: (project: Project) => void;
  onDelete: (project: Project) => void;
  delay: number;
}

export const ProjectCard = memo(function ProjectCard({
  project,
  launchers,
  terminal,
  openingEditorId,
  onOpenWith,
  onOpenDefault,
  onEdit,
  onCopy,
  onDelete,
  delay,
}: ProjectCardProps) {
  const [popover, setPopover] = useState<"more" | "hidden" | null>(null);
  const closePopover = useCallback(() => setPopover(null), []);

  // 弹层打开时 Esc 关闭(依赖为原始字符串)
  useEffect(() => {
    if (popover === null) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setPopover(null);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [popover]);

  const slots = useMemo(
    () => (terminal ? [terminal, ...launchers] : launchers),
    [terminal, launchers],
  );
  const visible = useMemo(() => slots.slice(0, MAX_VISIBLE), [slots]);
  const hidden = useMemo(() => slots.slice(MAX_VISIBLE), [slots]);

  return (
    <article
      className="anim-enter relative flex items-center gap-3 rounded-card border border-line bg-surface px-4 py-2.5 transition-colors duration-[var(--dur-fast)] ease-out hover:bg-surface-2"
      style={{ animationDelay: Math.min(delay, 12) * 30 + "ms" }}
    >
      {/* 左:项目信息(点卡片主体 = 默认编辑器打开) */}
      <div
        className="min-w-0 flex-1 cursor-pointer py-0.5"
        onClick={() => onOpenDefault(project.id)}
        title={"用默认编辑器打开 " + project.name}
      >
        <div className="flex items-center gap-2">
          <h3 className="min-w-0 truncate text-[15px] font-bold text-ink">{project.name}</h3>
          {project.isTemp ? (
            <span className="shrink-0 rounded-pill bg-accent-dim px-1.5 py-0.5 text-[10px] font-medium text-accent">
              临时
            </span>
          ) : null}
          <span className="flex-1" aria-hidden="true" />
          <span className="shrink-0 text-[11px] text-ink-3">
            {project.isTemp ? null : project.openCount > 0 ? "打开 " + project.openCount + " 次" : "未打开过"}
          </span>
          <span className="shrink-0 text-[11px] text-ink-3">
            {relativeTime(project.lastOpenedAt || project.createdAt)}
          </span>
        </div>
        {project.note ? <p className="mt-0.5 truncate text-[12px] text-ink-2">{project.note}</p> : null}
        <p className="mt-0.5 truncate font-mono text-[11px] text-ink-3" title={project.path}>
          {project.path}
        </p>
      </div>

      {/* 右:快捷启动图标(阻断冒泡,点击不触发左侧默认打开) */}
      <div className="flex shrink-0 items-center gap-1" onClick={(e) => e.stopPropagation()}>
        {visible.map((ed) => (
          <IconLaunch
            key={ed.id}
            editor={ed}
            spinning={openingEditorId === ed.id}
            onClick={() => onOpenWith(project.id, ed.id)}
          />
        ))}
        {hidden.length > 0 ? (
          <div className="relative">
            <button
              type="button"
              aria-label={"还有 " + hidden.length + " 个编辑器"}
              aria-expanded={popover === "hidden"}
              title={"还有 " + hidden.length + " 个编辑器"}
              onClick={() => setPopover((v) => (v === "hidden" ? null : "hidden"))}
              className={cx(
                "inline-flex size-7 items-center justify-center rounded-lg font-mono text-[11px] text-ink-2 transition duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.96] hover:bg-surface-2 hover:text-ink",
                popover === "hidden" && "bg-surface-2 text-ink",
              )}
            >
              +{hidden.length}
            </button>
            {popover === "hidden" ? (
              <>
                <div className="fixed inset-0 z-10" onClick={closePopover} aria-hidden="true" />
                <div
                  role="menu"
                  className="absolute bottom-full right-0 z-20 mb-2 w-44 rounded-input border border-line bg-surface-2 p-1"
                >
                  {hidden.map((ed) => (
                    <button
                      key={ed.id}
                      type="button"
                      role="menuitem"
                      onClick={() => {
                        closePopover();
                        onOpenWith(project.id, ed.id);
                      }}
                      className="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-[13px] text-ink transition-colors duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent hover:bg-surface"
                    >
                      <EditorIcon slug={ed.icon} className="size-4 shrink-0 text-ink-2" />
                      <span className="truncate">{ed.name}</span>
                    </button>
                  ))}
                </div>
              </>
            ) : null}
          </div>
        ) : null}
        <div className="relative">
          <button
            type="button"
            aria-label="更多操作"
            aria-expanded={popover === "more"}
            title="更多操作"
            onClick={() => setPopover((v) => (v === "more" ? null : "more"))}
            className={cx(
              "inline-flex size-7 items-center justify-center rounded-lg text-ink-2 transition duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.96] hover:bg-surface-2 hover:text-ink",
              popover === "more" && "bg-surface-2 text-ink",
            )}
          >
            <DotsIcon className="size-4" />
          </button>
          {popover === "more" ? (
            <>
              <div className="fixed inset-0 z-10" onClick={closePopover} aria-hidden="true" />
              <div
                role="menu"
                className="absolute bottom-full right-0 z-20 mb-2 w-36 rounded-input border border-line bg-surface-2 p-1"
              >
                <MenuItem label="编辑信息" onClick={() => { closePopover(); onEdit(project); }} />
                <MenuItem label="复制路径" onClick={() => { closePopover(); onCopy(project); }} />
                <MenuItem label="删除" danger onClick={() => { closePopover(); onDelete(project); }} />
              </div>
            </>
          ) : null}
        </div>
      </div>
    </article>
  );
});

function IconLaunch({ editor, spinning, onClick }: { editor: Editor; spinning: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      title={editor.name}
      aria-label={"用 " + editor.name + " 打开"}
      onClick={onClick}
      className="inline-flex size-7 items-center justify-center rounded-lg text-ink-2 transition duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.96] hover:bg-surface-2 hover:text-ink"
    >
      {spinning ? <Spinner className="size-4" /> : <EditorIcon slug={editor.icon} className="size-4" />}
    </button>
  );
}

function MenuItem({ label, danger = false, onClick }: { label: string; danger?: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      role="menuitem"
      onClick={onClick}
      className={cx(
        "block w-full rounded-lg px-3 py-2 text-left text-[13px] transition-colors duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent",
        danger ? "text-danger hover:bg-danger/10" : "text-ink hover:bg-surface",
      )}
    >
      {label}
    </button>
  );
}
