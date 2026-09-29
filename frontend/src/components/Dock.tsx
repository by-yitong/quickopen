import { cx } from "./ui";
import { PlusIcon, ProjectsNavIcon, SettingsNavIcon } from "./icons";
import type { View } from "../types";

/** 底部悬浮 Dock:左「项目」「设置」导航 + 右侧圆形 accent FAB。 */

interface DockProps {
  view: View;
  enabled: boolean;
  onNav: (view: View) => void;
  onCreate: () => void;
}

const NAV_ITEMS: Array<{ view: View; label: string }> = [
  { view: "projects", label: "项目" },
  { view: "settings", label: "设置" },
];

export function Dock({ view, enabled, onNav, onCreate }: DockProps) {
  return (
    <nav
      aria-label="主导航"
      className="fixed bottom-5 left-1/2 z-40 flex -translate-x-1/2 items-center gap-1 rounded-pill border border-line bg-surface/95 px-3 py-2 backdrop-blur"
    >
      {NAV_ITEMS.map((item) => {
        const active = view === item.view;
        return (
          <button
            key={item.view}
            type="button"
            disabled={!enabled}
            aria-current={active ? "page" : undefined}
            onClick={() => onNav(item.view)}
            className={cx(
              "flex h-11 flex-col items-center justify-center gap-0.5 rounded-pill px-4 text-[11px] font-medium transition-colors duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent active:scale-[0.98] disabled:pointer-events-none disabled:opacity-45",
              active ? "bg-accent-dim text-accent" : "text-ink-2 hover:bg-surface-2 hover:text-ink",
            )}
          >
            {item.view === "projects" ? (
              <ProjectsNavIcon className="size-[18px]" />
            ) : (
              <SettingsNavIcon className="size-[18px]" />
            )}
            {item.label}
          </button>
        );
      })}
      <span className="mx-1 h-8 w-px bg-line" aria-hidden="true" />
      <button
        type="button"
        disabled={!enabled}
        aria-label="新建项目"
        title="新建项目"
        onClick={onCreate}
        className="flex size-12 items-center justify-center rounded-pill bg-accent text-ink transition duration-[var(--dur-fast)] ease-out outline-offset-2 focus-visible:outline-2 focus-visible:outline-ink active:scale-95 hover:opacity-90 disabled:pointer-events-none disabled:opacity-45"
      >
        <PlusIcon className="size-[22px]" />
      </button>
    </nav>
  );
}
