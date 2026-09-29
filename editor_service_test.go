package main

import (
	"errors"
	"strings"
	"testing"
)

func TestAddManualEditor(t *testing.T) {
	tests := []struct {
		name    string
		eName   string
		command string
		args    string
		kind    EditorKind
		wantErr string
	}{
		{name: "正常添加", eName: "我的终端", command: "myterm", args: "-d {path}", kind: KindTerminal},
		{name: "args缺省", eName: "我的编辑器", command: "myed", args: "", kind: KindEditor},
		{name: "名称为空", eName: " ", command: "x", args: "", kind: KindEditor, wantErr: "名称不能为空"},
		{name: "命令为空", eName: "x", command: "", args: "", kind: KindEditor, wantErr: "命令不能为空"},
		{name: "类别无效", eName: "x", command: "y", args: "", kind: "bogus", wantErr: "类别无效"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTempConfig(t)
			s := &EditorService{}
			ed, err := s.AddManualEditor(tt.eName, tt.command, tt.args, tt.kind)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("期望错误含 %q,得到 %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("AddManualEditor: %v", err)
			}
			if !strings.HasPrefix(ed.ID, "m-") || len(ed.ID) != len("m-")+8 {
				t.Fatalf("手动 ID 应为 m-+8hex: %s", ed.ID)
			}
			if ed.Source != "manual" || !ed.Enabled {
				t.Fatalf("Source/Enabled 不符: %+v", ed)
			}
			wantArgs := tt.args
			if wantArgs == "" {
				wantArgs = "{path}"
			}
			if ed.Args != wantArgs {
				t.Fatalf("Args=%q,期望 %q", ed.Args, wantArgs)
			}
		})
	}
}

func TestUpdateDeleteEditor(t *testing.T) {
	withTempConfig(t)
	s := &EditorService{}
	ed, err := s.AddManualEditor("旧名", "old", "", KindEditor)
	if err != nil {
		t.Fatalf("AddManualEditor: %v", err)
	}

	upd, err := s.UpdateEditor(Editor{ID: ed.ID, Name: "新名", Command: "new", Args: "-a", Icon: "zed", Kind: KindTerminal, Enabled: false, InTerminal: true})
	if err != nil {
		t.Fatalf("UpdateEditor: %v", err)
	}
	if upd.Name != "新名" || upd.Command != "new" || upd.Args != "-a" || upd.Icon != "zed" || upd.Kind != KindTerminal || upd.Enabled || !upd.InTerminal {
		t.Fatalf("更新结果不符: %+v", upd)
	}

	if _, err := s.UpdateEditor(Editor{ID: "m-00000000", Name: "x"}); err == nil || !strings.Contains(err.Error(), "编辑器不存在") {
		t.Fatalf("更新不存在的编辑器应报错: %v", err)
	}

	if err := s.DeleteEditor(ed.ID); err != nil {
		t.Fatalf("DeleteEditor: %v", err)
	}
	if err := s.DeleteEditor(ed.ID); err == nil || !strings.Contains(err.Error(), "编辑器不存在") {
		t.Fatalf("重复删除应报错: %v", err)
	}
}

func TestScanEditorsMerge(t *testing.T) {
	withTempConfig(t)
	// 存量:vscode 被用户改过 Args 且禁用;konsole 已消失;手动项保留
	old := []Editor{
		{ID: "vscode", Name: "VS Code", Command: "code", Args: "--new-window {path}", Kind: KindEditor, Source: "scan", Enabled: false},
		{ID: "konsole", Name: "Konsole", Command: "konsole", Kind: KindTerminal, Source: "scan", Enabled: true},
		{ID: "m-abcd1234", Name: "手动", Command: "myed", Kind: KindEditor, Source: "manual", Enabled: true},
	}
	if err := saveEditors(old); err != nil {
		t.Fatalf("saveEditors: %v", err)
	}
	found := []Editor{
		{ID: "vscode", Name: "VS Code", Command: "code", Args: "{path}", Kind: KindEditor, Source: "scan", Enabled: true},
		{ID: "zed", Name: "Zed", Command: "zed", Args: "{path}", Kind: KindEditor, Source: "scan", Enabled: true},
	}
	oldScan := scanEditorsFunc
	scanEditorsFunc = func() ([]Editor, error) { return found, nil }
	t.Cleanup(func() { scanEditorsFunc = oldScan })

	s := &EditorService{}
	got, err := s.ScanEditors()
	if err != nil {
		t.Fatalf("ScanEditors: %v", err)
	}
	byID := map[string]Editor{}
	for _, e := range got {
		byID[e.ID] = e
	}
	if _, ok := byID["konsole"]; ok {
		t.Fatalf("消失的 scan 项应删除")
	}
	vs := byID["vscode"]
	if vs.Enabled || vs.Args != "--new-window {path}" {
		t.Fatalf("同名 ID 应保留 Enabled/Args 人工改动(用户已禁用应为 false): %+v", vs)
	}
	if byID["zed"].Name != "Zed" {
		t.Fatalf("新扫描项应加入: %+v", byID["zed"])
	}
	man := byID["m-abcd1234"]
	if man.Name != "手动" {
		t.Fatalf("manual 项应保留: %+v", man)
	}
	// 顺序:已有 scan 项保持旧相对顺序,新 scan 项追加,manual 在末尾
	order := []string{got[0].ID, got[1].ID, got[2].ID}
	if order[0] != "vscode" || order[1] != "zed" || order[2] != "m-abcd1234" {
		t.Fatalf("合并后顺序应为 vscode→zed→手动: %v", order)
	}
}

func TestMergeScanKeepsReorderedLayout(t *testing.T) {
	withTempConfig(t)
	// 用户已拖拽成 zed→code 的顺序;重扫后应保持,新扫描到的 cursor 追加
	old := []Editor{
		{ID: "zed", Name: "Zed", Command: "zed", Kind: KindEditor, Source: "scan", Enabled: true},
		{ID: "code", Name: "VS Code", Command: "code", Kind: KindEditor, Source: "scan", Enabled: true},
	}
	found := []Editor{
		{ID: "code", Name: "VS Code", Command: "code", Kind: KindEditor, Source: "scan", Enabled: true},
		{ID: "cursor", Name: "Cursor", Command: "cursor", Kind: KindEditor, Source: "scan", Enabled: true},
		{ID: "zed", Name: "Zed", Command: "zed", Kind: KindEditor, Source: "scan", Enabled: true},
	}
	got := mergeScan(old, found)
	if got[0].ID != "zed" || got[1].ID != "code" || got[2].ID != "cursor" {
		t.Fatalf("重扫应保持用户顺序并追加新项: %v", []string{got[0].ID, got[1].ID, got[2].ID})
	}
}

func TestReorderEditors(t *testing.T) {
	withTempConfig(t)
	old := []Editor{
		{ID: "code", Name: "VS Code", Command: "code", Kind: KindEditor, Source: "scan", Enabled: true},
		{ID: "zed", Name: "Zed", Command: "zed", Kind: KindEditor, Source: "scan", Enabled: true},
		{ID: "konsole", Name: "Konsole", Command: "konsole", Kind: KindTerminal, Source: "scan", Enabled: true},
	}
	if err := saveEditors(old); err != nil {
		t.Fatalf("saveEditors: %v", err)
	}
	s := &EditorService{}
	if err := s.ReorderEditors([]string{"zed", "konsole", "code"}); err != nil {
		t.Fatalf("ReorderEditors: %v", err)
	}
	got, _ := loadEditors()
	if got[0].ID != "zed" || got[1].ID != "konsole" || got[2].ID != "code" {
		t.Fatalf("顺序未落盘: %v", []string{got[0].ID, got[1].ID, got[2].ID})
	}
	// 数量不符 / 未知 ID / 重复 ID 都要报错
	if err := s.ReorderEditors([]string{"zed", "konsole"}); err == nil {
		t.Fatalf("数量不符应报错")
	}
	if err := s.ReorderEditors([]string{"zed", "konsole", "ghost"}); err == nil {
		t.Fatalf("未知 ID 应报错")
	}
	if err := s.ReorderEditors([]string{"zed", "zed", "code"}); err == nil {
		t.Fatalf("重复 ID 应报错")
	}
	// 失败后存量不变
	got, _ = loadEditors()
	if got[0].ID != "zed" || len(got) != 3 {
		t.Fatalf("失败调用不应改动存量: %+v", got)
	}
}

func TestScanEditorsScanError(t *testing.T) {
	withTempConfig(t)
	oldScan := scanEditorsFunc
	scanEditorsFunc = func() ([]Editor, error) { return nil, errFake }
	t.Cleanup(func() { scanEditorsFunc = oldScan })
	s := &EditorService{}
	if _, err := s.ScanEditors(); err == nil || !strings.Contains(err.Error(), "扫描编辑器失败") {
		t.Fatalf("扫描失败应报错: %v", err)
	}
}

var errFake = errors.New("fake scan error")

func TestSetDefaultEditorAndTerminal(t *testing.T) {
	withTempConfig(t)
	s := &EditorService{}
	if err := s.SetDefaultEditor("vscode"); err == nil || !strings.Contains(err.Error(), "编辑器不存在") {
		t.Fatalf("不存在的 ID 应报错: %v", err)
	}
	saveEditors([]Editor{codeEditor, konsoleEd})

	if err := s.SetDefaultEditor("vscode"); err != nil {
		t.Fatalf("SetDefaultEditor: %v", err)
	}
	if err := s.SetDefaultTerminal("konsole"); err != nil {
		t.Fatalf("SetDefaultTerminal: %v", err)
	}
	st, _ := loadSettings()
	if st.DefaultEditorID != "vscode" || st.DefaultTerminalID != "konsole" {
		t.Fatalf("设置未落盘: %+v", st)
	}

	if err := s.SetDefaultEditor(""); err != nil {
		t.Fatalf("清除应成功: %v", err)
	}
	if err := s.SetDefaultTerminal(""); err != nil {
		t.Fatalf("清除应成功: %v", err)
	}
	st, _ = loadSettings()
	if st.DefaultEditorID != "" || st.DefaultTerminalID != "" {
		t.Fatalf("应已清除: %+v", st)
	}
}

func TestListEditorsEmpty(t *testing.T) {
	withTempConfig(t)
	s := &EditorService{}
	got, err := s.ListEditors()
	if err != nil || len(got) != 0 {
		t.Fatalf("空配置应返回空列表: %v %v", got, err)
	}
}
