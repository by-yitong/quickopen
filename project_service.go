package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProjectService 项目管理,绑定给前端。
type ProjectService struct{}

// ListProjects 返回全部项目(按创建顺序)。
func (s *ProjectService) ListProjects() ([]Project, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return loadProjects()
}

// AddProject 登记已存在目录;name 为空取 basename。
func (s *ProjectService) AddProject(path, name, note string) (Project, error) {
	storeMu.Lock()
	defer storeMu.Unlock()
	ps, err := loadProjects()
	if err != nil {
		return Project{}, err
	}
	if strings.TrimSpace(path) == "" {
		return Project{}, fmt.Errorf("路径不能为空")
	}
	abs, err := absClean(path)
	if err != nil {
		return Project{}, err
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return Project{}, fmt.Errorf("目录不存在: %s", abs)
	}
	for _, p := range ps {
		if p.Path == abs {
			return Project{}, fmt.Errorf("该项目已在列表中")
		}
	}
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(abs)
	}
	p := Project{
		ID:        newID("p-"),
		Name:      name,
		Note:      note,
		Path:      abs,
		CreatedAt: nowMs(),
	}
	ps = append(ps, p)
	if err := saveProjects(ps); err != nil {
		return Project{}, err
	}
	return p, nil
}

// UpdateProject 按 ID 全量更新 Name/Note/Path,统计字段沿用存量。
func (s *ProjectService) UpdateProject(project Project) (Project, error) {
	storeMu.Lock()
	defer storeMu.Unlock()
	ps, err := loadProjects()
	if err != nil {
		return Project{}, err
	}
	idx := -1
	for i, p := range ps {
		if p.ID == project.ID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return Project{}, fmt.Errorf("项目不存在: %s", project.ID)
	}
	stored := &ps[idx]
	if strings.TrimSpace(project.Name) != "" {
		stored.Name = project.Name
	}
	stored.Note = project.Note
	if strings.TrimSpace(project.Path) != "" {
		abs, err := absClean(project.Path)
		if err != nil {
			return Project{}, err
		}
		for i, p := range ps {
			if i != idx && p.Path == abs {
				return Project{}, fmt.Errorf("该项目已在列表中")
			}
		}
		stored.Path = abs
	}
	if err := saveProjects(ps); err != nil {
		return Project{}, err
	}
	return *stored, nil
}

// DeleteProject 按 ID 删除。
func (s *ProjectService) DeleteProject(id string) error {
	storeMu.Lock()
	defer storeMu.Unlock()
	ps, err := loadProjects()
	if err != nil {
		return err
	}
	for i, p := range ps {
		if p.ID == id {
			ps = append(ps[:i], ps[i+1:]...)
			return saveProjects(ps)
		}
	}
	return fmt.Errorf("项目不存在: %s", id)
}

// OpenProject 用编辑器或终端打开项目;成功才更新 LastOpenedAt/OpenCount。
func (s *ProjectService) OpenProject(id, editorID string) error {
	storeMu.Lock()
	defer storeMu.Unlock()
	ps, err := loadProjects()
	if err != nil {
		return err
	}
	var proj *Project
	for i := range ps {
		if ps[i].ID == id {
			proj = &ps[i]
			break
		}
	}
	if proj == nil {
		return fmt.Errorf("项目不存在: %s", id)
	}
	if !fileExists(proj.Path) {
		return fmt.Errorf("目录不存在: %s", proj.Path)
	}
	editors, err := loadEditors()
	if err != nil {
		return err
	}
	settings, err := loadSettings()
	if err != nil {
		return err
	}
	ed, err := pickEditor(editorID, editors, settings)
	if err != nil {
		return err
	}
	if err := launchProject(ed, proj.Path, editors, settings); err != nil {
		return err // 启动失败不更新统计
	}
	proj.LastOpenedAt = nowMs()
	proj.OpenCount++
	return saveProjects(ps)
}

// pickEditor 选编辑器:editorID 优先,其次 Settings.DefaultEditorID。
func pickEditor(editorID string, editors []Editor, settings Settings) (Editor, error) {
	want := editorID
	if want == "" {
		want = settings.DefaultEditorID
	}
	if want == "" {
		return Editor{}, errors.New("未找到可用编辑器,请先在设置中扫描或添加")
	}
	for _, e := range editors {
		if e.ID == want {
			return e, nil
		}
	}
	return Editor{}, fmt.Errorf("编辑器不存在: %s", want)
}

// CreateProject 创建新项目目录;parentPath 为空用 Settings.DefaultPath。
func (s *ProjectService) CreateProject(name, note, parentPath string) (Project, error) {
	storeMu.Lock()
	defer storeMu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		return Project{}, fmt.Errorf("名称不能为空")
	}
	if strings.ContainsAny(name, `/\`) {
		return Project{}, fmt.Errorf("名称不能包含路径分隔符")
	}
	parent := parentPath
	if strings.TrimSpace(parent) == "" {
		st, err := loadSettings()
		if err != nil {
			return Project{}, err
		}
		parent = st.DefaultPath
	}
	if strings.TrimSpace(parent) == "" {
		return Project{}, fmt.Errorf("未设置默认目录,请先在设置中配置默认路径")
	}
	parentAbs, err := absClean(parent)
	if err != nil {
		return Project{}, err
	}
	ps, err := loadProjects()
	if err != nil {
		return Project{}, err
	}
	target := filepath.Join(parentAbs, name)
	for _, p := range ps {
		if p.Path == target {
			return Project{}, fmt.Errorf("该项目已在列表中")
		}
	}
	if _, err := os.Stat(target); err == nil {
		return Project{}, fmt.Errorf("目标目录已存在: %s", target)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return Project{}, fmt.Errorf("创建目录失败: %w", err)
	}
	p := Project{
		ID:        newID("p-"),
		Name:      name,
		Note:      note,
		Path:      target,
		CreatedAt: nowMs(),
	}
	ps = append(ps, p)
	if err := saveProjects(ps); err != nil {
		return Project{}, err
	}
	return p, nil
}

// CreateTempProject 在 <DefaultPath>/tmp 下建临时项目;name 为空自动生成,绝不自动删除。
func (s *ProjectService) CreateTempProject(name, note string) (Project, error) {
	storeMu.Lock()
	defer storeMu.Unlock()
	st, err := loadSettings()
	if err != nil {
		return Project{}, err
	}
	if strings.TrimSpace(st.DefaultPath) == "" {
		return Project{}, fmt.Errorf("未设置默认目录,请先在设置中配置默认路径")
	}
	parent := filepath.Join(st.DefaultPath, "tmp")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return Project{}, fmt.Errorf("创建目录失败: %w", err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name, err = nextTempName(parent, "tmp-"+time.Now().Format("20060102-150405"))
		if err != nil {
			return Project{}, err
		}
	} else {
		if strings.ContainsAny(name, `/\`) {
			return Project{}, fmt.Errorf("名称不能包含路径分隔符")
		}
		target := filepath.Join(parent, name)
		if _, err := os.Stat(target); err == nil {
			return Project{}, fmt.Errorf("目标目录已存在: %s", target)
		}
	}
	target := filepath.Join(parent, name)
	if err := os.Mkdir(target, 0o755); err != nil {
		return Project{}, fmt.Errorf("创建目录失败: %w", err)
	}
	p := Project{
		ID:        newID("p-"),
		Name:      name,
		Note:      note,
		Path:      target,
		IsTemp:    true,
		CreatedAt: nowMs(),
	}
	ps, err := loadProjects()
	if err != nil {
		return Project{}, err
	}
	ps = append(ps, p)
	if err := saveProjects(ps); err != nil {
		return Project{}, err
	}
	return p, nil
}

// nextTempName 基名已存在则追加 -2、-3…
func nextTempName(parent, base string) (string, error) {
	candidate := base
	for i := 2; fileExists(filepath.Join(parent, candidate)); i++ {
		if i > 1000 {
			return "", fmt.Errorf("临时目录名生成失败,重试过多")
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return candidate, nil
}
