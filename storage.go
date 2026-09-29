package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 存储层:配置目录下三个 JSON 文件,RWMutex 保护,写盘用临时文件+rename 原子替换。

const appDirName = "quickopen"

// userConfigDir 可注入,测试里替换成临时目录。
var userConfigDir = os.UserConfigDir

// userHomeDir 可注入,bootstrap 测试用。
var userHomeDir = os.UserHomeDir

// nowMs 返回当前 unix 毫秒。
func nowMs() int64 {
	return time.Now().UnixMilli()
}

// storeMu 保护三个 JSON 文件的读写。
var storeMu sync.RWMutex

func appConfigDir() (string, error) {
	base, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取配置目录失败: %w", err)
	}
	return filepath.Join(base, appDirName), nil
}

func projectsFile() (string, error) {
	dir, err := appConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "projects.json"), nil
}

func editorsFile() (string, error) {
	dir, err := appConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "editors.json"), nil
}

func settingsFile() (string, error) {
	dir, err := appConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// loadJSON 读入 JSON 文件;文件不存在时返回 wrapped os.ErrNotExist,调用方用 errors.Is 判断。
func loadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 %s 失败: %w", filepath.Base(path), err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("解析 %s 失败: %w", filepath.Base(path), err)
	}
	return nil
}

// saveJSON 原子写盘:同目录随机名临时文件(双开实例不互踩)+ fsync + rename。
func saveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}
	tmp := f.Name()
	if _, err := f.Write(data); err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp, 0o644)
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
		if strings.Contains(err.Error(), "rename") {
			return fmt.Errorf("替换文件失败: %w", err)
		}
		return fmt.Errorf("写入临时文件失败: %w", err)
	}
	return nil
}

func loadProjects() ([]Project, error) {
	path, err := projectsFile()
	if err != nil {
		return nil, err
	}
	var ps []Project
	if err := loadJSON(path, &ps); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Project{}, nil
		}
		return nil, err
	}
	if ps == nil {
		ps = []Project{}
	}
	return ps, nil
}

func saveProjects(ps []Project) error {
	path, err := projectsFile()
	if err != nil {
		return err
	}
	return saveJSON(path, ps)
}

func loadEditors() ([]Editor, error) {
	path, err := editorsFile()
	if err != nil {
		return nil, err
	}
	var es []Editor
	if err := loadJSON(path, &es); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Editor{}, nil
		}
		return nil, err
	}
	if es == nil {
		es = []Editor{}
	}
	return es, nil
}

func saveEditors(es []Editor) error {
	path, err := editorsFile()
	if err != nil {
		return err
	}
	return saveJSON(path, es)
}

func loadSettings() (Settings, error) {
	path, err := settingsFile()
	if err != nil {
		return Settings{}, err
	}
	var st Settings
	if err := loadJSON(path, &st); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Settings{}, nil
		}
		return Settings{}, err
	}
	return st, nil
}

func saveSettings(st Settings) error {
	path, err := settingsFile()
	if err != nil {
		return err
	}
	return saveJSON(path, st)
}

// newID 生成 "前缀"+8位hex 的 ID。
func newID(prefix string) string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand 失败极罕见,退回时间戳保证唯一性
		return fmt.Sprintf("%s%08x", prefix, time.Now().UnixNano()&0xffffffff)
	}
	return prefix + hex.EncodeToString(buf)
}

// absClean 转绝对路径并清洗。
func absClean(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("解析路径失败: %w", err)
	}
	return filepath.Clean(abs), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
