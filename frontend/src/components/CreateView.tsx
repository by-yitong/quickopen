import { useCallback, useMemo, useState } from "react";
import type { Editor, Project, Settings } from "../../bindings/quickopen/models.js";
import { CreateProject, CreateTempProject, OpenProject } from "../../bindings/quickopen/projectservice.js";
import { PickDirectory } from "../../bindings/quickopen/settingsservice.js";
import { errMsg, joinPath, validateName } from "../lib/format";
import { Button, FieldError, Spinner, Switch, TextInput } from "./ui";
import { EditorIcon } from "./icons";
import type { ToastKind } from "./Toast";

/**
 * 视图 2:新建项目。
 * 「临时项目」开关打开后免填名称(自动命名,建在 <默认路径>/tmp/ 下);
 * 操作区统一:「仅创建」或点编辑器图标「创建并直接打开」。
 */

interface CreateViewProps {
  settings: Settings;
  editors: Editor[];
  pushToast: (kind: ToastKind, message: string) => void;
  onCreated: (project: Project) => void;
}

export function CreateView({ settings, editors, pushToast, onCreated }: CreateViewProps) {
  const [isTemp, setIsTemp] = useState(false);
  const [name, setName] = useState("");
  const [note, setNote] = useState("");
  const [parentPath, setParentPath] = useState(() => settings.defaultPath);
  const [nameError, setNameError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [picking, setPicking] = useState(false);

  const launchers = useMemo(
    () => editors.filter((e) => e.enabled && (e.kind === "editor" || e.kind === "cli")),
    [editors],
  );

  // 临时项目位置固定为 <默认路径>/tmp,不可自选
  const effectiveParent = isTemp ? joinPath(settings.defaultPath, "tmp") : parentPath.trim();
  const previewPath = isTemp
    ? joinPath(effectiveParent, "tmp-…(自动命名)")
    : joinPath(parentPath.trim(), name.trim());

  const changeName = useCallback((value: string) => {
    setName(value);
    setNameError(null);
  }, []);

  const changeParent = useCallback(() => {
    if (picking) return;
    setPicking(true);
    PickDirectory()
      .then((dir) => {
        if (dir) setParentPath(dir);
      })
      .catch((err: unknown) => pushToast("error", errMsg(err)))
      .finally(() => setPicking(false));
  }, [picking, pushToast]);

  const submit = useCallback(
    (ed: Editor | null) => {
      if (busy) return;
      const err = isTemp ? null : validateName(name, true);
      setNameError(err);
      if (err) return;
      setBusy(ed === null ? "create" : ed.id);
      const create = isTemp
        ? CreateTempProject("", note.trim())
        : CreateProject(name.trim(), note.trim(), parentPath.trim());
      create
        .then((p) => {
          onCreated(p); // 项目已登记:先回列表,再尝试打开
          if (ed === null) {
            pushToast("success", "已创建");
            return undefined;
          }
          return OpenProject(p.id, ed.id).then(
            () => pushToast("success", "已创建并打开"),
            (err: unknown) => pushToast("error", errMsg(err)),
          );
        })
        .catch((err: unknown) => pushToast("error", errMsg(err)))
        .finally(() => setBusy(null));
    },
    [busy, isTemp, name, note, parentPath, pushToast, onCreated],
  );

  return (
    <div className="anim-view rounded-card border border-line bg-surface p-6">
      <form className="space-y-4" onSubmit={(e) => e.preventDefault()}>
        {/* 临时项目开关:打开后免填名称,位置固定 <默认路径>/tmp */}
        <div className="flex items-center gap-3">
          <Switch
            checked={isTemp}
            label="临时项目"
            disabled={busy !== null}
            onChange={() => {
              setIsTemp((v) => !v);
              setNameError(null);
            }}
          />
          <div className="min-w-0">
            <p className="text-[13px] font-medium text-ink">临时项目</p>
            <p className="truncate text-[12px] text-ink-3">
              自动命名,建在 <span className="font-mono">{joinPath(settings.defaultPath, "tmp")}/</span> 下,不会自动删除
            </p>
          </div>
        </div>

        {isTemp ? null : (
          <label className="block">
            <TextInput
              value={name}
              onChange={(e) => changeName(e.target.value)}
              placeholder="项目名称"
              aria-label="项目名称"
            />
            {nameError ? <FieldError message={nameError} /> : null}
          </label>
        )}
        <TextInput
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="备注(可选)"
          aria-label="备注"
        />

        {/* 位置:纯文本展示 + 更改链接;临时项目位置固定不可改 */}
        <div>
          <span className="mb-1 block text-[13px] font-medium text-ink-2">位置</span>
          <p className="truncate font-mono text-[13px] text-ink-2" title={effectiveParent || "未设置默认路径"}>
            {effectiveParent || "未设置默认路径"}
          </p>
          {isTemp ? null : (
            <button
              type="button"
              onClick={changeParent}
              disabled={picking || busy !== null}
              className="mt-0.5 rounded-pill text-[12px] text-accent underline-offset-2 transition-colors duration-[var(--dur-fast)] ease-out outline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-accent disabled:opacity-45"
            >
              {picking ? "选择中…" : "更改"}
            </button>
          )}
          <p className="mt-1.5 truncate font-mono text-[12px] text-ink-3" title={previewPath}>
            {previewPath}
          </p>
        </div>

        {/* 操作区:标题行左侧文字、右侧「仅创建」;下方编辑器图标 = 创建并打开 */}
        <div className="border-t border-line pt-4">
          <div className="flex items-center justify-between gap-2">
            <p className="text-[13px] font-medium text-ink-2">
              创建并直接打开{isTemp ? "(自动命名的临时目录)" : ""}
            </p>
            <Button variant="ghost" onClick={() => submit(null)} loading={busy === "create"} disabled={busy !== null}>
              仅创建
            </Button>
          </div>
          <div className="mt-3 flex flex-wrap items-center gap-2">
            {launchers.length === 0 ? (
              <p className="text-[12px] text-ink-3">没有可用编辑器,请先到设置页扫描或添加</p>
            ) : (
              launchers.map((ed) => (
                <button
                  key={ed.id}
                  type="button"
                  title={"创建并用 " + ed.name + " 打开"}
                  aria-label={"创建并用 " + ed.name + " 打开"}
                  disabled={busy !== null}
                  onClick={() => submit(ed)}
                  className="inline-flex size-11 items-center justify-center rounded-xl bg-surface-2 text-ink transition duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.96] hover:bg-surface disabled:pointer-events-none disabled:opacity-45"
                >
                  {busy === ed.id ? <Spinner className="size-[18px]" /> : <EditorIcon slug={ed.icon} className="size-[18px]" />}
                </button>
              ))
            )}
          </div>
        </div>
      </form>
    </div>
  );
}
