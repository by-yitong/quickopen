package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// SettingsService 应用设置,绑定给前端。
type SettingsService struct{}

// GetSettings 读设置;未初始化时返回零值。
func (s *SettingsService) GetSettings() (Settings, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return loadSettings()
}

// SaveSettings 保存设置;DefaultPath 不存在则 MkdirAll。
func (s *SettingsService) SaveSettings(settings Settings) (Settings, error) {
	storeMu.Lock()
	defer storeMu.Unlock()
	if strings.TrimSpace(settings.DefaultPath) != "" {
		abs, err := absClean(settings.DefaultPath)
		if err != nil {
			return Settings{}, err
		}
		if err := os.MkdirAll(abs, 0o755); err != nil {
			return Settings{}, fmt.Errorf("创建目录失败: %w", err)
		}
		settings.DefaultPath = abs
	}
	if err := saveSettings(settings); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

// PickDirectory 原生目录选择框;取消返回 "", nil。
func (s *SettingsService) PickDirectory() (string, error) {
	path, err := application.Get().Dialog.OpenFile().
		CanChooseFiles(false).
		CanChooseDirectories(true).
		CanCreateDirectories(true).
		SetTitle("选择目录").
		PromptForSingleSelection()
	if err != nil {
		// Linux 实现对取消可能返回错误,错误串含 cancel/取消 按取消处理
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "cancel") || strings.Contains(err.Error(), "取消") {
			return "", nil
		}
		return "", fmt.Errorf("打开目录选择框失败: %w", err)
	}
	return path, nil
}
