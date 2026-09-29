package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// bootstrap 首次启动初始化:补齐配置文件、默认路径与默认编辑器/终端。
// 窗口创建前在 main 里调用。
func bootstrap() error {
	dir, err := appConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// projects.json:不存在则落空列表
	pp, err := projectsFile()
	if err != nil {
		return err
	}
	if !fileExists(pp) {
		if err := saveProjects([]Project{}); err != nil {
			return err
		}
	}

	// editors.json:不存在则落盘首次扫描结果;已存在则每次启动自动合并——
	// 新装的编辑器/CLI 工具自动出现,卸载的自动消失;顺序/开关/手动项由 mergeScan 保留。
	ep, err := editorsFile()
	if err != nil {
		return err
	}
	if !fileExists(ep) {
		eds, err := scanEditorsFunc()
		if err != nil {
			return fmt.Errorf("首次扫描编辑器失败: %w", err)
		}
		if err := saveEditors(eds); err != nil {
			return err
		}
	} else {
		// 自动合并失败不阻断启动(扫描器或磁盘的临时问题下次启动再试)
		if found, err := scanEditorsFunc(); err == nil {
			if existing, err := loadEditors(); err == nil {
				_ = saveEditors(mergeScan(existing, found))
			}
		}
	}

	// settings.json:不存在则 DefaultPath=$HOME/Projects 并创建该目录
	sp, err := settingsFile()
	if err != nil {
		return err
	}
	st, err := loadSettings()
	if err != nil {
		return err
	}
	fresh := !fileExists(sp)
	if fresh {
		home, err := userHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户主目录失败: %w", err)
		}
		st.DefaultPath = filepath.Join(home, "Projects")
		if err := os.MkdirAll(st.DefaultPath, 0o755); err != nil {
			return fmt.Errorf("创建默认项目目录失败: %w", err)
		}
	}

	// 仅首次初始化(本次新建 settings.json)时自动挑选默认编辑器/终端;
	// 之后用户可通过 SetDefault*("") 显式清除,重启不再回填。
	if fresh {
		editors, err := loadEditors()
		if err != nil {
			return err
		}
		// 扫描项 ID 取命令名(code/zed/konsole…),VS Code 的命令名是 code
		st.DefaultEditorID = pickDefaultID(editors, KindEditor, "code")
		st.DefaultTerminalID = pickDefaultID(editors, KindTerminal, "konsole")
	}
	return saveSettings(st)
}

// pickDefaultID 选默认项:优先 preferred(且 enabled),否则第一个 enabled。
func pickDefaultID(editors []Editor, kind EditorKind, preferred string) string {
	var first string
	for _, e := range editors {
		if e.Kind != kind || !e.Enabled {
			continue
		}
		if first == "" {
			first = e.ID
		}
		if e.ID == preferred {
			return e.ID
		}
	}
	return first
}
