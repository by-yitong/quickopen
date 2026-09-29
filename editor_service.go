package main

import (
	"fmt"
	"sort"
	"strings"
)

// EditorService 编辑器/终端管理,绑定给前端。
type EditorService struct{}

// scanEditorsFunc 可注入,测试里替换成假扫描器。
var scanEditorsFunc = scanSystemEditors

// ListEditors 返回全部编辑器/终端。
func (s *EditorService) ListEditors() ([]Editor, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return loadEditors()
}

// ScanEditors 全量重扫:保留 manual;同名 ID 保留 Enabled/Args 人工改动;消失的 scan 项删除。
// 扫描在锁外进行(文件系统慢,不阻塞其他存储调用),合并落盘才持锁。
func (s *EditorService) ScanEditors() ([]Editor, error) {
	found, err := scanEditorsFunc()
	if err != nil {
		return nil, fmt.Errorf("扫描编辑器失败: %w", err)
	}
	storeMu.Lock()
	defer storeMu.Unlock()
	existing, err := loadEditors()
	if err != nil {
		return nil, err
	}
	merged := mergeScan(existing, found)
	if err := saveEditors(merged); err != nil {
		return nil, err
	}
	return merged, nil
}

// mergeScan 合并扫描结果与存量:同名 ID 保留 Enabled/Args 人工改动;
// 已有 scan 项保持旧相对顺序(用户可拖拽排序),新 scan 项按扫描顺序追加,manual 保持在末尾的原相对顺序。
func mergeScan(existing, found []Editor) []Editor {
	oldIdx := make(map[string]int, len(existing))
	tweaks := make(map[string]Editor, len(existing))
	manuals := make([]Editor, 0)
	for i, e := range existing {
		oldIdx[e.ID] = i
		if e.Source == "manual" {
			manuals = append(manuals, e)
		} else {
			tweaks[e.ID] = e
		}
	}
	updated := make([]Editor, 0, len(found))
	for _, f := range found {
		if old, ok := tweaks[f.ID]; ok {
			f.Enabled = old.Enabled
			f.Args = old.Args
		}
		updated = append(updated, f)
	}
	sort.SliceStable(updated, func(i, j int) bool {
		ii, iok := oldIdx[updated[i].ID]
		jj, jok := oldIdx[updated[j].ID]
		if iok && jok {
			return ii < jj
		}
		return iok // 已存在的排前面,新增的按 found 顺序稳定追加
	})
	return append(updated, manuals...)
}

// AddManualEditor 手动添加编辑器或终端。
func (s *EditorService) AddManualEditor(name, command, args string, kind EditorKind) (Editor, error) {
	storeMu.Lock()
	defer storeMu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		return Editor{}, fmt.Errorf("名称不能为空")
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return Editor{}, fmt.Errorf("命令不能为空")
	}
	if kind != KindEditor && kind != KindTerminal && kind != KindCLI {
		return Editor{}, fmt.Errorf("类别无效: %s", kind)
	}
	if strings.TrimSpace(args) == "" && kind != KindCLI {
		args = "{path}"
	}
	ed := Editor{
		ID:      newID("m-"),
		Name:    name,
		Icon:    "generic",
		Command: command,
		Args:    args,
		Kind:    kind,
		Source:  "manual",
		Enabled: true,
	}
	es, err := loadEditors()
	if err != nil {
		return Editor{}, err
	}
	es = append(es, ed)
	if err := saveEditors(es); err != nil {
		return Editor{}, err
	}
	return ed, nil
}

// UpdateEditor 按 ID 全量更新(Icon 留空则沿用旧值)。
func (s *EditorService) UpdateEditor(editor Editor) (Editor, error) {
	storeMu.Lock()
	defer storeMu.Unlock()
	es, err := loadEditors()
	if err != nil {
		return Editor{}, err
	}
	for i, e := range es {
		if e.ID != editor.ID {
			continue
		}
		if strings.TrimSpace(editor.Name) != "" {
			es[i].Name = editor.Name
		}
		if strings.TrimSpace(editor.Command) != "" {
			es[i].Command = editor.Command
		}
		if strings.TrimSpace(editor.Args) != "" {
			es[i].Args = editor.Args
		}
		if strings.TrimSpace(editor.Icon) != "" {
			es[i].Icon = editor.Icon
		}
		if editor.Kind == KindEditor || editor.Kind == KindTerminal {
			es[i].Kind = editor.Kind
		}
		es[i].Enabled = editor.Enabled
		es[i].InTerminal = editor.InTerminal
		if err := saveEditors(es); err != nil {
			return Editor{}, err
		}
		return es[i], nil
	}
	return Editor{}, fmt.Errorf("编辑器不存在: %s", editor.ID)
}

// ReorderEditors 按给定 ID 顺序重排编辑器列表(拖拽排序落盘);ids 必须恰好覆盖全部现有 ID。
func (s *EditorService) ReorderEditors(ids []string) error {
	storeMu.Lock()
	defer storeMu.Unlock()
	es, err := loadEditors()
	if err != nil {
		return err
	}
	if len(ids) != len(es) {
		return fmt.Errorf("ID 数量与编辑器数量不符: %d != %d", len(ids), len(es))
	}
	byID := make(map[string]Editor, len(es))
	for _, e := range es {
		byID[e.ID] = e
	}
	out := make([]Editor, 0, len(es))
	seen := make(map[string]bool, len(es))
	for _, id := range ids {
		e, ok := byID[id]
		if !ok {
			return fmt.Errorf("编辑器不存在: %s", id)
		}
		if seen[id] {
			return fmt.Errorf("ID 重复: %s", id)
		}
		seen[id] = true
		out = append(out, e)
	}
	return saveEditors(out)
}

// DeleteEditor 按 ID 删除。
func (s *EditorService) DeleteEditor(id string) error {
	storeMu.Lock()
	defer storeMu.Unlock()
	es, err := loadEditors()
	if err != nil {
		return err
	}
	for i, e := range es {
		if e.ID == id {
			es = append(es[:i], es[i+1:]...)
			return saveEditors(es)
		}
	}
	return fmt.Errorf("编辑器不存在: %s", id)
}

// SetDefaultEditor 设默认编辑器;"" 清除;需已存在。
func (s *EditorService) SetDefaultEditor(id string) error {
	return setDefault(id, true)
}

// SetDefaultTerminal 设默认终端;"" 清除;需已存在。
func (s *EditorService) SetDefaultTerminal(id string) error {
	return setDefault(id, false)
}

func setDefault(id string, isEditor bool) error {
	storeMu.Lock()
	defer storeMu.Unlock()
	if strings.TrimSpace(id) != "" {
		es, err := loadEditors()
		if err != nil {
			return err
		}
		found := false
		for _, e := range es {
			if e.ID == id {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("编辑器不存在: %s", id)
		}
	}
	st, err := loadSettings()
	if err != nil {
		return err
	}
	if isEditor {
		st.DefaultEditorID = id
	} else {
		st.DefaultTerminalID = id
	}
	return saveSettings(st)
}
