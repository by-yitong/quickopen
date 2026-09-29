#!/usr/bin/env bash
# 从系统应用菜单卸载「项目快开」(不动 ~/.config/quickopen 里的数据)。
set -euo pipefail
rm -f "$HOME/.local/bin/quickopen"
rm -f "$HOME/.local/share/applications/quickopen.desktop"
rm -f "$HOME/.local/share/icons/hicolor/512x512/apps/quickopen.png"
command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$HOME/.local/share/applications" || true
echo "已从系统应用菜单移除。"
