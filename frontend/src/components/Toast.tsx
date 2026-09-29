import { cx } from "./ui";

/** Toast 状态(App 级持有,新的替换旧的)。 */
export interface ToastData {
  id: number;
  kind: ToastKind;
  message: string;
}

export type ToastKind = "success" | "error" | "info";

export function ToastView({ toast }: { toast: ToastData | null }) {
  if (!toast) return null;
  return (
    <div className="pointer-events-none fixed inset-x-0 top-4 z-[60] flex justify-center px-6">
      <div
        key={toast.id}
        role="status"
        className={cx(
          "anim-toast flex items-center gap-2 rounded-pill border bg-surface px-4 py-2.5 text-[13px] font-medium",
          toast.kind === "error" ? "border-danger/40" : "border-line",
        )}
      >
        <span
          className={cx(
            "size-1.5 shrink-0 rounded-pill",
            toast.kind === "error" ? "bg-danger" : toast.kind === "success" ? "bg-accent" : "bg-ink-2",
          )}
          aria-hidden="true"
        />
        <span className="truncate text-ink">{toast.message}</span>
      </div>
    </div>
  );
}
