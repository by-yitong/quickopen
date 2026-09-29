package main

import (
	"os"
	"path/filepath"
	"testing"
)

// withTempConfig 把配置目录指向临时目录(含 quickopen 子目录),测试结束自动还原。
// 返回 quickopen 配置目录本身。
func withTempConfig(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	dir := filepath.Join(base, "quickopen")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("创建临时配置目录: %v", err)
	}
	old := userConfigDir
	userConfigDir = func() (string, error) { return base, nil }
	t.Cleanup(func() { userConfigDir = old })
	return dir
}

// withTempHome 把主目录指向临时目录。
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := userHomeDir
	userHomeDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { userHomeDir = old })
	return dir
}

func TestSaveLoadJSONRoundTrip(t *testing.T) {
	withTempConfig(t)
	ps := []Project{{ID: "p-1", Name: "demo", Path: "/tmp/demo", CreatedAt: 100}}
	if err := saveProjects(ps); err != nil {
		t.Fatalf("saveProjects: %v", err)
	}
	got, err := loadProjects()
	if err != nil {
		t.Fatalf("loadProjects: %v", err)
	}
	if len(got) != 1 || got[0].Name != "demo" || got[0].CreatedAt != 100 {
		t.Fatalf("往返数据不一致: %+v", got)
	}
}

func TestSaveJSONAtomicNoTmpLeftover(t *testing.T) {
	dir := withTempConfig(t)
	ps := []Project{{ID: "p-1", Name: "a"}}
	if err := saveProjects(ps); err != nil {
		t.Fatalf("saveProjects: %v", err)
	}
	if err := saveProjects(append(ps, Project{ID: "p-2", Name: "b"})); err != nil {
		t.Fatalf("saveProjects 第二次: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "projects.json.tmp")); !os.IsNotExist(err) {
		t.Fatalf("临时文件未被清理: %v", err)
	}
	got, err := loadProjects()
	if err != nil {
		t.Fatalf("loadProjects: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("原子替换后应有 2 条,得到 %d", len(got))
	}
}

func TestLoadJSONMissingFile(t *testing.T) {
	withTempConfig(t)
	got, err := loadProjects()
	if err != nil {
		t.Fatalf("文件缺失应返回空列表,得到错误: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("应为空切片: %+v", got)
	}
	st, err := loadSettings()
	if err != nil {
		t.Fatalf("loadSettings 缺文件: %v", err)
	}
	if st != (Settings{}) {
		t.Fatalf("缺文件应返回零值设置: %+v", st)
	}
}

func TestNewID(t *testing.T) {
	a, b := newID("p-"), newID("p-")
	if len(a) != len("p-")+8 {
		t.Fatalf("ID 长度应为 10: %s", a)
	}
	if a == b {
		t.Fatalf("ID 不应重复: %s", a)
	}
}
