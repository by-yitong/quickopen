import {
  useCallback,
  useEffect,
  useState,
  type ButtonHTMLAttributes,
  type InputHTMLAttributes,
  type ReactNode,
} from "react";
import { CloseIcon } from "./icons";

/**
 * 基础 UI 件(文件级组件,全站复用)。
 * 8 态纪律:default / hover / focus-visible / active / disabled / loading / error(由调用方文案)/ success(toast)。
 * 颜色一律引用 @theme token;动效只走 transform/opacity。
 */

export function cx(...parts: Array<string | false | null | undefined>): string {
  return parts.filter(Boolean).join(" ");
}

/* ---- Spinner ---- */

export function Spinner({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" className={cx("animate-spin", className ?? "size-4")} aria-hidden="true">
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeOpacity="0.25" strokeWidth="3" />
      <path d="M21 12a9 9 0 0 0-9-9" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
    </svg>
  );
}

/* ---- Button ---- */

type ButtonVariant = "secondary" | "accent-outline" | "danger" | "ghost";

const BTN_BASE =
  "inline-flex items-center justify-center gap-1.5 rounded-input px-4 py-2 text-[14px] font-medium transition duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.98] disabled:pointer-events-none disabled:opacity-45";

const BTN_VARIANTS: Record<ButtonVariant, string> = {
  secondary: "border border-line bg-surface text-ink hover:bg-surface-2",
  "accent-outline": "border border-accent/60 bg-transparent text-accent hover:bg-accent-dim",
  danger: "border border-danger/60 bg-danger/15 text-danger hover:bg-danger/25",
  ghost: "bg-transparent text-ink-2 hover:bg-surface-2 hover:text-ink",
};

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  loading?: boolean;
}

export function Button({ variant = "secondary", loading = false, className, children, disabled, ...rest }: ButtonProps) {
  return (
    <button
      type="button"
      {...rest}
      disabled={disabled || loading}
      className={cx(BTN_BASE, BTN_VARIANTS[variant], className)}
    >
      {loading ? <Spinner className="size-4" /> : null}
      {children}
    </button>
  );
}

/* ---- IconButton(卡片/列表行内图标按钮) ---- */

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  label: string;
  spinning?: boolean;
  variant?: "ghost" | "tile" | "danger-ghost";
}

const ICON_BTN_BASE =
  "inline-flex shrink-0 items-center justify-center text-ink-2 transition duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.96] disabled:pointer-events-none disabled:opacity-45";

const ICON_BTN_VARIANTS = {
  ghost: "size-8 rounded-lg bg-transparent hover:bg-surface-2 hover:text-ink",
  tile: "size-11 rounded-xl bg-surface-2 text-ink hover:bg-surface",
  "danger-ghost": "size-8 rounded-lg bg-transparent hover:bg-danger/15 hover:text-danger",
} as const;

export function IconButton({ label, spinning = false, variant = "ghost", className, children, ...rest }: IconButtonProps) {
  return (
    <button type="button" {...rest} title={label} aria-label={label} className={cx(ICON_BTN_BASE, ICON_BTN_VARIANTS[variant], className)}>
      {spinning ? <Spinner className={variant === "tile" ? "size-5" : "size-4"} /> : children}
    </button>
  );
}

/* ---- Switch(Enabled 开关) ---- */

interface SwitchProps {
  checked: boolean;
  label: string;
  disabled?: boolean;
  onChange: () => void;
}

export function Switch({ checked, label, disabled, onChange }: SwitchProps) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      onClick={onChange}
      className={cx(
        "relative h-6 w-10 shrink-0 rounded-pill transition-colors duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent disabled:opacity-45",
        checked ? "bg-accent" : "bg-surface-2",
      )}
    >
      <span
        className={cx(
          "absolute left-0.5 top-0.5 size-5 rounded-pill bg-ink transition-transform duration-[var(--dur-fast)] ease-out",
          checked ? "translate-x-4" : "translate-x-0",
        )}
      />
    </button>
  );
}

/* ---- 输入框 ---- */

export const INPUT_CLS =
  "w-full rounded-input border border-line bg-surface-2 px-3.5 py-2.5 text-[14px] text-ink transition-colors duration-[var(--dur-fast)] outline-offset-2 placeholder:text-ink-3 focus:border-accent focus-visible:outline-2 focus-visible:outline-accent";

export function TextInput({ className, ...rest }: InputHTMLAttributes<HTMLInputElement>) {
  return <input {...rest} className={cx(INPUT_CLS, className)} />;
}

export function FieldError({ message }: { message: string }) {
  return <p className="mt-1.5 text-[12px] text-danger">{message}</p>;
}

/* ---- Modal(Esc 关闭 + 锁 body 滚动;fade + 8px 上移) ---- */

interface ModalProps {
  title: string;
  onClose: () => void;
  children: ReactNode;
}

export function Modal({ title, onClose, children }: ModalProps) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      window.removeEventListener("keydown", onKey);
      document.body.style.overflow = prev;
    };
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="anim-fade absolute inset-0 bg-paper/80" onClick={onClose} aria-hidden="true" />
      <div
        role="dialog"
        aria-modal="true"
        aria-label={title}
        className="anim-pop relative w-full max-w-md rounded-card border border-line bg-surface p-6"
      >
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="text-[17px] font-bold text-ink">{title}</h2>
          <IconButton label="关闭" onClick={onClose}>
            <CloseIcon className="size-4" />
          </IconButton>
        </div>
        {children}
      </div>
    </div>
  );
}

/* ---- ConfirmDialog(危险操作二次确认) ---- */

interface ConfirmDialogProps {
  title: string;
  body: string;
  confirmLabel: string;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}

export function ConfirmDialog({ title, body, confirmLabel, onClose, onConfirm }: ConfirmDialogProps) {
  const [busy, setBusy] = useState(false);

  const handleConfirm = useCallback(() => {
    if (busy) return;
    setBusy(true);
    onConfirm()
      .then(onClose)
      .catch(() => setBusy(false));
  }, [busy, onConfirm, onClose, setBusy]);

  return (
    <Modal title={title} onClose={onClose}>
      <p className="text-[13px] leading-relaxed text-ink-2">{body}</p>
      <div className="mt-5 flex justify-end gap-2">
        <Button onClick={onClose}>取消</Button>
        <Button variant="danger" loading={busy} onClick={handleConfirm}>
          {confirmLabel}
        </Button>
      </div>
    </Modal>
  );
}

/* ---- 骨架卡(加载态) ---- */

export function SkeletonCard() {
  return (
    <div className="rounded-card border border-line bg-surface p-5" aria-hidden="true">
      <div className="flex items-center justify-between">
        <div className="h-5 w-14 animate-pulse rounded-pill bg-surface-2" />
        <div className="h-4 w-16 animate-pulse rounded-pill bg-surface-2" />
      </div>
      <div className="mt-4 h-4 w-2/5 animate-pulse rounded-pill bg-surface-2" />
      <div className="mt-3 h-3.5 w-3/5 animate-pulse rounded-pill bg-surface-2" />
      <div className="mt-5 flex gap-2">
        <div className="size-8 animate-pulse rounded-lg bg-surface-2" />
        <div className="size-8 animate-pulse rounded-lg bg-surface-2" />
      </div>
    </div>
  );
}
