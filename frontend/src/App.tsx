import { useCallback, useEffect, useRef, useState } from "react";
import type { Editor, Project, Settings } from "../bindings/quickopen/models.js";
import {
  AddProject,
  DeleteProject,
  ListProjects,
  OpenProject,
  UpdateProject,
} from "../bindings/quickopen/projectservice.js";
import { ListEditors } from "../bindings/quickopen/editorservice.js";
import { GetSettings, PickDirectory } from "../bindings/quickopen/settingsservice.js";
import { errMsg } from "./lib/format";
import { Button, SkeletonCard } from "./components/ui";
import { Dock } from "./components/Dock";
import { ToastView, type ToastData, type ToastKind } from "./components/Toast";
import { ProjectsView } from "./components/ProjectsView";
import { CreateView } from "./components/CreateView";
import { SettingsView } from "./components/SettingsView";
import type { View } from "./types";

/**
 * App:三视图框架 + 初始加载(Promise.all 并行)+ Toast + Dock。
 * 性能纪律:全部 handler useCallback;视图切换三元;effect 依赖只用原始值。
 */

type LoadState = "loading" | "error" | "ready";

const VIEW_TITLES: Record<View, string> = {
  projects: "项目快开",
  create: "新建项目",
  settings: "设置",
};

export default function App() {
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [loadError, setLoadError] = useState("");
  const [projects, setProjects] = useState<Project[]>([]);
  const [editors, setEditors] = useState<Editor[]>([]);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [view, setView] = useState<View>("projects");
  const [toast, setToast] = useState<ToastData | null>(null);
  const [openingKey, setOpeningKey] = useState<string | null>(null);
  const openingRef = useRef(false);

  /* ---- 初始加载:并行拉取,失败整页错误态 ---- */
  const load = useCallback(() => {
    setLoadState("loading");
    setLoadError("");
    Promise.all([ListProjects(), ListEditors(), GetSettings()])
      .then(([ps, es, st]) => {
        setProjects(ps ?? []);
        setEditors(es ?? []);
        setSettings(st);
        setLoadState("ready");
      })
      .catch((err: unknown) => {
        setLoadError(errMsg(err));
        setLoadState("error");
      });
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  /* ---- Toast:2.4s 自动消失,新替旧(effect 依赖只用原始值) ---- */
  const pushToast = useCallback((kind: ToastKind, message: string) => {
    setToast({ id: Date.now(), kind, message });
  }, []);
  const toastId = toast?.id ?? 0;
  useEffect(() => {
    if (toastId === 0) return;
    const t = window.setTimeout(() => {
      setToast((cur) => (cur !== null && cur.id === toastId ? null : cur));
    }, 2400);
    return () => window.clearTimeout(t);
  }, [toastId]);

  /* ---- 打开项目(默认编辑器 / 指定编辑器;key = projectId|editorId) ---- */
  const openProject = useCallback(
    (projectId: string, editorId: string) => {
      if (openingRef.current) {
        pushToast("info", "正在打开,请稍候");
        return;
      }
      openingRef.current = true;
      const key = projectId + "|" + editorId;
      setOpeningKey(key);
      OpenProject(projectId, editorId)
        .then(() => {
          const now = Date.now();
          setProjects((prev) =>
            prev.map((p) => (p.id === projectId ? { ...p, lastOpenedAt: now, openCount: p.openCount + 1 } : p)),
          );
        })
        .catch((err: unknown) => pushToast("error", errMsg(err)))
        .finally(() => {
          openingRef.current = false;
          setOpeningKey((cur) => (cur === key ? null : cur));
        });
    },
    [pushToast],
  );

  const openDefault = useCallback(
    (projectId: string) => {
      openProject(projectId, settings?.defaultEditorId ?? "");
    },
    [openProject, settings],
  );

  /* ---- 项目变更 ---- */
  const updateProject = useCallback(
    (p: Project) =>
      UpdateProject(p).then((saved) => {
        setProjects((prev) => prev.map((x) => (x.id === saved.id ? saved : x)));
        pushToast("success", "已保存");
      }),
    [pushToast],
  );

  const deleteProject = useCallback(
    (id: string) =>
      DeleteProject(id).then(
        () => {
          setProjects((prev) => prev.filter((p) => p.id !== id));
          pushToast("success", "已删除");
        },
        (err: unknown) => {
          pushToast("error", errMsg(err));
          throw err; // 弹层据此外抛保持打开
        },
      ),
    [pushToast],
  );

  /* ---- 登记已有目录:原生选目录 → AddProject → 追加列表 ---- */
  const addExisting = useCallback(() => {
    PickDirectory()
      .then((dir) => {
        if (!dir) return undefined;
        return AddProject(dir, "", "").then((p) => {
          setProjects((prev) => [...prev, p]);
          pushToast("success", "已添加 " + p.name);
        });
      })
      .catch((err: unknown) => pushToast("error", errMsg(err)));
  }, [pushToast]);

  /* ---- 新建完成:登记 + 跳回列表 ---- */
  const handleCreated = useCallback((p: Project) => {
    setProjects((prev) => [p, ...prev]);
    setView("projects");
  }, []);

  /* ---- Dock 导航 ---- */
  const handleNav = useCallback((next: View) => setView(next), []);
  const handleCreate = useCallback(() => setView("create"), []);

  const ready = loadState === "ready" && settings !== null;
  const tempCount = ready ? projects.reduce((n, p) => (p.isTemp ? n + 1 : n), 0) : 0;
  const subtitle =
    tempCount > 0 ? projects.length + " 个项目 · " + tempCount + " 个临时" : projects.length + " 个项目";

  return (
    <div className="mx-auto min-h-dvh w-full max-w-2xl px-6 pb-28">
      <header className="pt-8">
        <h1 className="text-[34px] font-bold leading-tight text-ink">{VIEW_TITLES[view]}</h1>
        {view === "projects" && ready ? (
          <div className="mt-1 flex items-center justify-between gap-2">
            <p className="text-[13px] text-ink-3">{subtitle}</p>
            <button
              type="button"
              onClick={addExisting}
              className="shrink-0 rounded-pill px-2.5 py-1.5 text-[13px] text-ink-2 transition-colors duration-[var(--dur-fast)] ease-out outline-offset-2 hover:bg-surface hover:text-ink focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.98]"
            >
              + 添加已有目录
            </button>
          </div>
        ) : null}
      </header>

      <main className="mt-6">
        {loadState === "loading" ? (
          <div className="anim-fade space-y-3" aria-busy="true" aria-label="加载中">
            <SkeletonCard />
            <SkeletonCard />
            <SkeletonCard />
          </div>
        ) : loadState === "error" ? (
          <div className="anim-fade flex flex-col items-center gap-3 py-20 text-center">
            <p className="text-[15px] font-semibold text-ink">加载失败</p>
            <p className="max-w-md break-all text-[13px] text-ink-2">{loadError}</p>
            <Button onClick={load}>重试</Button>
          </div>
        ) : ready && settings !== null ? (
          view === "projects" ? (
            <ProjectsView
              projects={projects}
              editors={editors}
              defaultEditorId={settings.defaultEditorId}
              defaultTerminalId={settings.defaultTerminalId}
              openingKey={openingKey}
              pushToast={pushToast}
              onOpenWith={openProject}
              onOpenDefault={openDefault}
              onUpdateProject={updateProject}
              onDeleteProject={deleteProject}
            />
          ) : view === "create" ? (
            <CreateView settings={settings} editors={editors} pushToast={pushToast} onCreated={handleCreated} />
          ) : (
            <SettingsView
              settings={settings}
              editors={editors}
              setSettings={setSettings}
              setEditors={setEditors}
              pushToast={pushToast}
            />
          )
        ) : null}
      </main>

      <Dock view={view} enabled={loadState === "ready"} onNav={handleNav} onCreate={handleCreate} />
      <ToastView toast={toast} />
    </div>
  );
}
