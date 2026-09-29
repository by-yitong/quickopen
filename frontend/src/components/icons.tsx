import type { JSX } from "react";

/**
 * 编辑器/终端图标:icon slug → 内联单色 SVG(几何近似,fill=currentColor)。
 * 未知 slug 一律走 generic(圆角方块 + 首字母)。全部手绘 path,无外链资源。
 */

interface IconProps {
  className?: string;
}

function VscodeIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" fillRule="evenodd" className={className} aria-hidden="true">
      <path d="M17.9 2.6 21 4.2v15.6l-3.1 1.6-9.2-7.8L5 16.5 3 15.4V8.6l2-1.1 3.7 2.9 9.2-7.8ZM17.9 7 11 12l6.9 5V7Z" />
    </svg>
  );
}

function CursorIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" fillRule="evenodd" className={className} aria-hidden="true">
      <path d="M12 2.5 21 7.5v9l-9 5-9-5v-9l9-5Zm0 2.3L5.6 8.3 12 11.8l6.4-3.5L12 4.8Zm-7 5.2v5.9l6 3.3v-5.9l-6-3.3Zm14 0-6 3.3v5.9l6-3.3V10Z" />
    </svg>
  );
}

function ZedIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M4 4h16v4.2L11.8 16H20v4H4v-4.2L12.2 8H4V4Z" />
    </svg>
  );
}

function SublimeIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M4 2.5 20 1v5.5L4 8V2.5Zm0 8L20 9v5.5L4 16v-5.5Zm0 8L20 17v5.5L4 24v-5.5Z" />
    </svg>
  );
}

function JetbrainsIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" fillRule="evenodd" className={className} aria-hidden="true">
      <path d="M3 3h18v18H3V3Zm4 12.5h10v2H7v-2Zm0-8.5h7v2H7V7Z" />
    </svg>
  );
}

function VimIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M12 1.5l2.2 2.2L12 5.9 9.8 3.7 12 1.5ZM6 7h4l2 7.5L14 7h4l-4.5 15h-3L6 7Z" />
    </svg>
  );
}

function NvimIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" fillRule="evenodd" className={className} aria-hidden="true">
      <path d="M2.5 5.5 10 2v20l-7.5-3.5v-13Zm19 0L14 2v20l7.5-3.5v-13Z" />
    </svg>
  );
}

function EmacsIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" fillRule="evenodd" className={className} aria-hidden="true">
      <path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20Zm0 3.2a6.8 6.8 0 1 1 0 13.6 6.8 6.8 0 0 1 0-13.6Zm-6.5 5.6h13v2.4h-13v-2.4Z" />
    </svg>
  );
}

function HelixIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M3.5 3h5L21 16.5V21h-5L3.5 7.5V3Zm17 0h-5L3 16.5V21h5L20.5 7.5V3Z" />
    </svg>
  );
}

function WindsurfIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M2.5 13.5C6 8.5 10.5 7.5 15 9.5c3 1.3 5 1.5 6.5.8-1 4.7-5.5 7.7-10.5 7-3.6-.5-6.6-1.7-8.5-3.8Z" />
    </svg>
  );
}

function VscodiumIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" fillRule="evenodd" className={className} aria-hidden="true">
      <path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20Zm0 3a7 7 0 1 1 0 14 7 7 0 0 1 0-14ZM9.5 8.5 16 12l-6.5 3.5v-7Z" />
    </svg>
  );
}

function TraeIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M3.5 4h17v5.5H15V20H9V9.5H3.5V4Z" />
    </svg>
  );
}

function TerminalIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M4 4.5h3.2l7 7.5-7 7.5H4l7-7.5-7-7.5Zm10.5 10.5H21v3h-6.5v-3Z" />
    </svg>
  );
}

function GenericIcon({ className, letter }: IconProps & { letter: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} aria-hidden="true">
      <rect x="3.5" y="3.5" width="17" height="17" rx="4.5" fill="none" stroke="currentColor" strokeWidth="2" />
      <text
        x="12"
        y="16.2"
        textAnchor="middle"
        fontSize="10.5"
        fontWeight="700"
        fill="currentColor"
        style={{ fontFamily: "var(--font-sans)" }}
      >
        {letter}
      </text>
    </svg>
  );
}

/* ---- 界面通用图标(非编辑器) ---- */

export function PlusIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M11 4h2v7h7v2h-7v7h-2v-7H4v-2h7V4Z" />
    </svg>
  );
}

export function SearchIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" className={className} aria-hidden="true">
      <circle cx="11" cy="11" r="6.5" />
      <path d="m16.2 16.2 4.3 4.3" />
    </svg>
  );
}

export function GripIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <circle cx="9" cy="6" r="1.6" />
      <circle cx="15" cy="6" r="1.6" />
      <circle cx="9" cy="12" r="1.6" />
      <circle cx="15" cy="12" r="1.6" />
      <circle cx="9" cy="18" r="1.6" />
      <circle cx="15" cy="18" r="1.6" />
    </svg>
  );
}

export function DotsIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <circle cx="5" cy="12" r="1.9" />
      <circle cx="12" cy="12" r="1.9" />
      <circle cx="19" cy="12" r="1.9" />
    </svg>
  );
}

export function CloseIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" className={className} aria-hidden="true">
      <path d="M6 6l12 12M18 6 6 18" />
    </svg>
  );
}

export function ChevronDownIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className} aria-hidden="true">
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}

export function TrashIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M9.5 3h5l.8 2H20v2H4V5h4.7l.8-2ZM6 8.5h12l-.9 12.5H6.9L6 8.5Z" />
    </svg>
  );
}

export function ProjectsNavIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M3 5h2v2H3V5Zm4 0h14v2H7V5ZM3 11h2v2H3v-2Zm4 0h14v2H7v-2ZM3 17h2v2H3v-2Zm4 0h14v2H7v-2Z" />
    </svg>
  );
}

export function SettingsNavIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" fillRule="evenodd" className={className} aria-hidden="true">
      <rect x="11" y="2" width="2" height="3.6" rx="1" />
      <rect x="11" y="2" width="2" height="3.6" rx="1" transform="rotate(45 12 12)" />
      <rect x="11" y="2" width="2" height="3.6" rx="1" transform="rotate(90 12 12)" />
      <rect x="11" y="2" width="2" height="3.6" rx="1" transform="rotate(135 12 12)" />
      <path d="M12 8.2a3.8 3.8 0 1 0 0 7.6 3.8 3.8 0 0 0 0-7.6Zm0 2.2a1.6 1.6 0 1 1 0 3.2 1.6 1.6 0 0 1 0-3.2Z" />
    </svg>
  );
}

/* ---- slug 注册表 ---- */

function CodexIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" fillRule="evenodd" className={className} aria-hidden="true">
      <path d="M12 2.2 20.3 7v10L12 21.8 3.7 17V7L12 2.2Zm0 2.6L6 8.3v7.4l6 3.5 6-3.5V8.3l-6-3.5Zm0 3 3.7 2.1v4.2L12 16.2l-3.7-2.1V9.9L12 7.8Z" />
    </svg>
  );
}

function ClaudeIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M11 2h2v7l4.95-4.95 1.41 1.41L14.41 10.4H21v2h-6.59l4.95 4.95-1.41 1.41L13 13.81V21h-2v-7.19l-4.95 4.95-1.41-1.41 4.95-4.95H3v-2h6.59L4.64 5.46l1.41-1.41L11 9V2Z" />
    </svg>
  );
}

function GeminiIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M12 2c.62 4.9 4.08 8.36 10 10-5.92 1.64-9.38 5.1-10 10-.62-4.9-4.08-8.36-10-10 5.92-1.64 9.38-5.1 10-10Z" />
    </svg>
  );
}

function GrokIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M3.6 3h4.2L12 9.4 16.2 3h4.2l-6.3 9 6.3 9h-4.2L12 14.6 7.8 21H3.6l6.3-9-6.3-9Z" />
    </svg>
  );
}

function AiderIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M5.5 4.2 14.2 12l-8.7 7.8V4.2Z" />
      <path d="M14 18.2h6v2.6h-6v-2.6Z" />
    </svg>
  );
}

function AgentIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" fillRule="evenodd" className={className} aria-hidden="true">
      <path d="M11 2h2v4h-2V2Zm-7 6h16v9a3 3 0 0 1-3 3H7a3 3 0 0 1-3-3V8Zm5 3.2v3h2.4v-3H9Zm6 0v3h2.4v-3H15Z" />
    </svg>
  );
}

const SLUG_ICONS: Record<string, (props: IconProps) => JSX.Element> = {
  vscode: VscodeIcon,
  cursor: CursorIcon,
  zed: ZedIcon,
  sublime: SublimeIcon,
  jetbrains: JetbrainsIcon,
  vim: VimIcon,
  nvim: NvimIcon,
  emacs: EmacsIcon,
  helix: HelixIcon,
  windsurf: WindsurfIcon,
  vscodium: VscodiumIcon,
  trae: TraeIcon,
  terminal: TerminalIcon,
  codex: CodexIcon,
  claude: ClaudeIcon,
  gemini: GeminiIcon,
  grok: GrokIcon,
  aider: AiderIcon,
  agent: AgentIcon,
};

export function EditorIcon({ slug, className }: IconProps & { slug: string }) {
  const Found = SLUG_ICONS[slug];
  if (Found) return <Found className={className} />;
  return <GenericIcon className={className} letter={(slug || "?").charAt(0).toUpperCase()} />;
}
