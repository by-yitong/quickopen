/** 相对时间:刚刚 / N 分钟前 / N 小时前 / N 天前 / 超过 7 天 → YYYY-MM-DD。 */
export function relativeTime(ts: number): string {
  if (!ts) return "";
  const diff = Date.now() - ts;
  const MIN = 60_000;
  const HOUR = 3_600_000;
  const DAY = 86_400_000;
  if (diff < MIN) return "刚刚";
  if (diff < HOUR) return Math.floor(diff / MIN) + " 分钟前";
  if (diff < DAY) return Math.floor(diff / HOUR) + " 小时前";
  if (diff < 7 * DAY) return Math.floor(diff / DAY) + " 天前";
  const d = new Date(ts);
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return d.getFullYear() + "-" + m + "-" + day;
}

/** 目录拼接:去掉尾部分隔符后以 / 连接。dir 为空时返回 name。 */
export function joinPath(dir: string, name: string): string {
  if (!dir) return name;
  return dir.replace(/[\\/]+$/, "") + "/" + name;
}

/** 名称校验:required 时非空;一律不允许 / \。返回错误文案或 null。 */
export function validateName(name: string, required: boolean): string | null {
  if (required && name.trim() === "") return "名称不能为空";
  if (/[\\/]/.test(name)) return "名称不能包含 / 或 \\";
  return null;
}

/** 后端错误统一转成可展示字符串(Wails 可能抛 Error 或字符串)。 */
export function errMsg(err: unknown): string {
  if (err instanceof Error && err.message) return err.message;
  if (typeof err === "string" && err) return err;
  return "操作失败";
}

/** 剪贴板:优先 Clipboard API,失败退回 execCommand。 */
export function copyText(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) return navigator.clipboard.writeText(text);
  return new Promise<void>((resolve, reject) => {
    const ta = document.createElement("textarea");
    ta.value = text;
    ta.style.position = "fixed";
    ta.style.opacity = "0";
    document.body.appendChild(ta);
    ta.select();
    try {
      if (document.execCommand("copy")) resolve();
      else reject(new Error("复制失败"));
    } finally {
      document.body.removeChild(ta);
    }
  });
}
