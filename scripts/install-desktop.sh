#!/usr/bin/env bash
# 把「项目快开」安装进系统应用菜单(Linux 桌面入口)。
# 用法:先 wails3 task build,再 scripts/install-desktop.sh;卸载用 scripts/uninstall-desktop.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN_SRC="$ROOT/bin/quickopen"
ICON_SRC="$ROOT/build/appicon.png"

[ -x "$BIN_SRC" ] || { echo "错误: 未找到 $BIN_SRC,请先运行 wails3 task build"; exit 1; }
[ -f "$ICON_SRC" ] || { echo "错误: 未找到 $ICON_SRC,请先运行 go run scripts/genicon.go"; exit 1; }

install -Dm755 "$BIN_SRC" "$HOME/.local/bin/quickopen"
install -Dm644 "$ICON_SRC" "$HOME/.local/share/icons/hicolor/512x512/apps/quickopen.png"
mkdir -p "$HOME/.local/share/applications"
cat > "$HOME/.local/share/applications/quickopen.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=项目快开
GenericName=项目启动器
Comment=记忆项目路径,一键用编辑器/终端打开
Exec=$HOME/.local/bin/quickopen
Icon=quickopen
Terminal=false
Categories=Development;
Keywords=project;launcher;vscode;项目;启动;
StartupWMClass=quickopen
EOF

command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$HOME/.local/share/applications" || true
command -v gtk-update-icon-cache >/dev/null 2>&1 && gtk-update-icon-cache -f -t "$HOME/.local/share/icons/hicolor" 2>/dev/null || true
echo "已安装:应用菜单(Activities/应用网格)里搜「项目快开」即可启动。"
