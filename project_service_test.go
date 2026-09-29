package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddProject(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) (path, name, note string)
		wantErr   string
		wantName  string
		wantIsTemp bool
	}{
		{
			name: "正常登记_名称取basename",
			setup: func(t *testing.T) (string, string, string) {
				dir := withTempConfig(t)
				p := filepath.Join(dir, "demo")
				os.MkdirAll(p, 0o755)
				return p, "", ""
			},
			wantName: "demo",
		},
		{
			name: "指定名称与备注",
			setup: func(t *testing.T) (string, string, string) {
				dir := withTempConfig(t)
				p := filepath.Join(dir, "demo2")
				os.MkdirAll(p, 0o755)
				return p, "我的项目", "备注A"
			},
			wantName: "我的项目",
		},
		{
			name: "目录不存在",
			setup: func(t *testing.T) (string, string, string) {
				dir := withTempConfig(t)
				return filepath.Join(dir, "missing"), "", ""
			},
			wantErr: "目录不存在",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProjectService{}
			path, name, note := tt.setup(t)
			p, err := s.AddProject(path, name, note)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("期望错误含 %q,得到 %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("AddProject: %v", err)
			}
			if p.Name != tt.wantName {
				t.Fatalf("名称=%q,期望 %q", p.Name, tt.wantName)
			}
			if !strings.HasPrefix(p.ID, "p-") {
				t.Fatalf("ID 前缀错误: %s", p.ID)
			}
			if p.IsTemp {
				t.Fatalf("非临时项目 IsTemp 应为 false")
			}
		})
	}
}

func TestAddProjectDuplicate(t *testing.T) {
	withTempConfig(t)
	s := &ProjectService{}
	p, err := s.AddProject(".", "a", "")
	if err != nil {
		t.Fatalf("第一次添加: %v", err)
	}
	abs := p.Path
	_, err = s.AddProject(filepath.Join(filepath.Dir(abs), ".", filepath.Base(abs)), "b", "")
	if err == nil || err.Error() != "该项目已在列表中" {
		t.Fatalf("重复路径应报「该项目已在列表中」,得到 %v", err)
	}
}

func TestUpdateAndDeleteProject(t *testing.T) {
	withTempConfig(t)
	s := &ProjectService{}
	p, _ := s.AddProject(".", "", "")

	renamed, err := s.UpdateProject(Project{ID: p.ID, Name: "新名", Note: "新备注"})
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if renamed.Name != "新名" || renamed.Note != "新备注" {
		t.Fatalf("更新未生效: %+v", renamed)
	}
	if renamed.CreatedAt != p.CreatedAt {
		t.Fatalf("CreatedAt 不应被改: %d != %d", renamed.CreatedAt, p.CreatedAt)
	}

	if _, err := s.UpdateProject(Project{ID: "p-nope", Name: "x"}); err == nil {
		t.Fatalf("不存在的 ID 应报错")
	} else if !strings.Contains(err.Error(), "项目不存在") {
		t.Fatalf("错误消息不符: %v", err)
	}

	if err := s.DeleteProject(p.ID, false); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	if err := s.DeleteProject(p.ID, false); err == nil || !strings.Contains(err.Error(), "项目不存在") {
		t.Fatalf("重复删除应报项目不存在: %v", err)
	}
}

func TestCreateProject(t *testing.T) {
	tests := []struct {
		name       string
		nameArg    string
		parentArg  string
		preExists  string // 预先创建的目标目录(相对临时根)
		wantErr    string
	}{
		{name: "正常创建", nameArg: "alpha", parentArg: "parent"},
		{name: "父目录不存在则MkdirAll", nameArg: "beta", parentArg: "deep/nested"},
		{name: "名称为空", nameArg: "  ", parentArg: "p", wantErr: "名称不能为空"},
		{name: "名称含斜杠", nameArg: "a/b", parentArg: "p", wantErr: "名称不能包含路径分隔符"},
		{name: "名称含反斜杠", nameArg: `a\b`, parentArg: "p", wantErr: "名称不能包含路径分隔符"},
		{name: "目标已存在", nameArg: "same", parentArg: "p", preExists: "p/same", wantErr: "目标目录已存在"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := withTempConfig(t)
			st, _ := loadSettings()
			st.DefaultPath = root
			saveSettings(st)
			if tt.preExists != "" {
				os.MkdirAll(filepath.Join(root, tt.preExists), 0o755)
			}
			s := &ProjectService{}
			parent := tt.parentArg
			if parent != "" {
				parent = filepath.Join(root, parent)
			}
			p, err := s.CreateProject(tt.nameArg, "note", parent)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("期望错误含 %q,得到 %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateProject: %v", err)
			}
			if info, err := os.Stat(p.Path); err != nil || !info.IsDir() {
				t.Fatalf("目录应已创建: %v", err)
			}
		})
	}
}

func TestCreateProjectDuplicateInList(t *testing.T) {
	root := withTempConfig(t)
	st, _ := loadSettings()
	st.DefaultPath = root
	saveSettings(st)
	s := &ProjectService{}
	// 列表里登记了一个目录,再创建同名新目录应去重报错
	os.MkdirAll(filepath.Join(root, "dup"), 0o755)
	if _, err := s.AddProject(filepath.Join(root, "dup"), "", ""); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	os.RemoveAll(filepath.Join(root, "dup"))
	_, err := s.CreateProject("dup", "", "")
	if err == nil || err.Error() != "该项目已在列表中" {
		t.Fatalf("期望「该项目已在列表中」,得到 %v", err)
	}
}

func TestCreateProjectNoDefaultPath(t *testing.T) {
	withTempConfig(t)
	s := &ProjectService{}
	_, err := s.CreateProject("x", "", "")
	if err == nil || !strings.Contains(err.Error(), "未设置默认目录") {
		t.Fatalf("期望未设置默认目录错误,得到 %v", err)
	}
}

func TestCreateTempProject(t *testing.T) {
	root := withTempConfig(t)
	st, _ := loadSettings()
	st.DefaultPath = root
	saveSettings(st)
	s := &ProjectService{}

	p1, err := s.CreateTempProject("", "临时1")
	if err != nil {
		t.Fatalf("CreateTempProject: %v", err)
	}
	if !p1.IsTemp {
		t.Fatalf("IsTemp 应为 true")
	}
	if !strings.HasPrefix(p1.Name, "tmp-") {
		t.Fatalf("名称应以 tmp- 开头: %s", p1.Name)
	}
	if filepath.Dir(p1.Path) != filepath.Join(root, "tmp") {
		t.Fatalf("父目录应为 <DefaultPath>/tmp: %s", p1.Path)
	}
	if _, err := os.Stat(p1.Path); err != nil {
		t.Fatalf("目录应已创建: %v", err)
	}

	// 同秒碰撞 → 追加后缀,路径必然不同
	p2, err := s.CreateTempProject("", "临时2")
	if err != nil {
		t.Fatalf("第二次创建: %v", err)
	}
	if p2.Path == p1.Path {
		t.Fatalf("碰撞时路径不应相同")
	}
	if !p2.IsTemp || p2.Note != "临时2" {
		t.Fatalf("第二个临时项目字段不符: %+v", p2)
	}

	// 用户自定义名称 → 直接用该名称
	p3, err := s.CreateTempProject("my-sandbox", "")
	if err != nil {
		t.Fatalf("自定义名称创建: %v", err)
	}
	if p3.Name != "my-sandbox" || filepath.Base(p3.Path) != "my-sandbox" {
		t.Fatalf("自定义名称未生效: %+v", p3)
	}
	// 同名再建 → 报已存在
	if _, err := s.CreateTempProject("my-sandbox", ""); err == nil || !strings.Contains(err.Error(), "已存在") {
		t.Fatalf("期望已存在错误,得到 %v", err)
	}
	// 非法名称 → 报路径分隔符
	if _, err := s.CreateTempProject("a/b", ""); err == nil || !strings.Contains(err.Error(), "路径分隔符") {
		t.Fatalf("期望路径分隔符错误,得到 %v", err)
	}

	// 直接占位下一个名字,验证 -2 逻辑
	os.MkdirAll(filepath.Join(root, "tmp", "tmp-fixed"), 0o755)
	got, err := nextTempName(filepath.Join(root, "tmp"), "tmp-fixed")
	if err != nil {
		t.Fatalf("nextTempName: %v", err)
	}
	if got != "tmp-fixed-2" {
		t.Fatalf("碰撞应追加 -2,得到 %s", got)
	}
	if err := os.MkdirAll(filepath.Join(root, "tmp", "tmp-fixed-2"), 0o755); err != nil {
		t.Fatalf("占位: %v", err)
	}
	got, err = nextTempName(filepath.Join(root, "tmp"), "tmp-fixed")
	if err != nil || got != "tmp-fixed-3" {
		t.Fatalf("二次碰撞应追加 -3,得到 %s err=%v", got, err)
	}
}

// fakeSpawn 记录 argv,可控成败。
type fakeSpawn struct {
	calls   [][]string
	failAt  int // 第 n 次调用(1 起)失败;0=不失败
	lastErr error
}

func (f *fakeSpawn) install(t *testing.T) {
	t.Helper()
	old := startProcessFunc
	startProcessFunc = func(argv []string) error {
		f.calls = append(f.calls, argv)
		if f.failAt > 0 && len(f.calls) == f.failAt {
			f.lastErr = errors.New("spawn 失败")
			return f.lastErr
		}
		return nil
	}
	t.Cleanup(func() { startProcessFunc = old })
}

// seedProjectWithEditors 造一个项目 + 编辑器/终端清单。
func seedProjectWithEditors(t *testing.T, editors ...Editor) Project {
	t.Helper()
	withTempConfig(t)
	dir := t.TempDir()
	pdir := filepath.Join(dir, "proj")
	os.MkdirAll(pdir, 0o755)
	s := &ProjectService{}
	p, err := s.AddProject(pdir, "", "")
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if err := saveEditors(editors); err != nil {
		t.Fatalf("saveEditors: %v", err)
	}
	return p
}

var (
	codeEditor = Editor{ID: "vscode", Name: "VS Code", Icon: "vscode", Command: "code", Args: "{path}", Kind: KindEditor, Source: "scan", Enabled: true}
	nvimEditor = Editor{ID: "nvim", Name: "Neovim", Icon: "nvim", Command: "nvim", Args: "{path}", Kind: KindEditor, Source: "scan", Enabled: true, InTerminal: true}
	konsoleEd  = Editor{ID: "konsole", Name: "Konsole", Icon: "terminal", Command: "konsole", Args: "{path}", Kind: KindTerminal, Source: "scan", Enabled: true}
)

func TestOpenProjectEditor(t *testing.T) {
	tests := []struct {
		name       string
		editors    []Editor
		editorID   string
		setDefault string
		wantErr    string
		check      func(t *testing.T, f *fakeSpawn, p Project, statsUpdated bool)
	}{
		{
			name:     "显式编辑器",
			editors:  []Editor{codeEditor, konsoleEd},
			editorID: "vscode",
			check: func(t *testing.T, f *fakeSpawn, p Project, ok bool) {
				if len(f.calls) != 1 || f.calls[0][0] != "code" {
					t.Fatalf("argv 不符: %v", f.calls)
				}
			},
		},
		{
			name:       "默认编辑器",
			editors:    []Editor{codeEditor, konsoleEd},
			setDefault: "vscode",
			check: func(t *testing.T, f *fakeSpawn, p Project, ok bool) {
				if len(f.calls) != 1 || f.calls[0][0] != "code" {
					t.Fatalf("argv 不符: %v", f.calls)
				}
			},
		},
		{
			name:    "无任何编辑器",
			editors: nil,
			wantErr: "未找到可用编辑器",
		},
		{
			name:     "指定编辑器不存在",
			editors:  []Editor{codeEditor},
			editorID: "zed",
			wantErr:  "编辑器不存在",
		},
		{
			name:     "终端内编辑器走默认终端",
			editors:  []Editor{nvimEditor, konsoleEd},
			editorID: "nvim",
			check: func(t *testing.T, f *fakeSpawn, p Project, ok bool) {
				if len(f.calls) != 1 {
					t.Fatalf("应只启动一次: %v", f.calls)
				}
				argv := f.calls[0]
				if argv[0] != "konsole" || argv[1] != "--workdir" {
					t.Fatalf("终端 argv 不符: %v", argv)
				}
				joined := strings.Join(argv, " ")
				if !strings.Contains(joined, "nvim") || !strings.Contains(joined, p.Path) {
					t.Fatalf("inner 应含编辑器与路径: %v", argv)
				}
			},
		},
		{
			name:     "终端内编辑器但无终端",
			editors:  []Editor{nvimEditor},
			editorID: "nvim",
			wantErr:  "未找到可用终端",
		},
		{
			name:     "终端类条目直接打开",
			editors:  []Editor{konsoleEd},
			editorID: "konsole",
			check: func(t *testing.T, f *fakeSpawn, p Project, ok bool) {
				argv := f.calls[0]
				if argv[0] != "konsole" || argv[1] != "--workdir" || argv[2] != p.Path {
					t.Fatalf("终端 argv 不符: %v", argv)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := seedProjectWithEditors(t, tt.editors...)
			if tt.setDefault != "" {
				saveSettings(Settings{DefaultEditorID: tt.setDefault, DefaultTerminalID: "konsole"})
			} else if tt.editors != nil {
				// 有终端需求时配置默认终端
				for _, e := range tt.editors {
					if e.Kind == KindTerminal {
						saveSettings(Settings{DefaultTerminalID: e.ID})
						break
					}
				}
			}
			f := &fakeSpawn{}
			f.install(t)
			s := &ProjectService{}
			err := s.OpenProject(p.ID, tt.editorID)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("期望错误含 %q,得到 %v", tt.wantErr, err)
				}
				got, _ := loadProjects()
				if got[0].OpenCount != 0 {
					t.Fatalf("失败时统计不应更新")
				}
				return
			}
			if err != nil {
				t.Fatalf("OpenProject: %v", err)
			}
			got, _ := loadProjects()
			if got[0].OpenCount != 1 || got[0].LastOpenedAt == 0 {
				t.Fatalf("成功后应更新统计: %+v", got[0])
			}
			if tt.check != nil {
				tt.check(t, f, p, true)
			}
		})
	}
}

func TestOpenProjectSpawnFailureKeepsStats(t *testing.T) {
	p := seedProjectWithEditors(t, codeEditor, konsoleEd)
	saveSettings(Settings{DefaultTerminalID: "konsole"})
	f := &fakeSpawn{failAt: 1}
	f.install(t)
	s := &ProjectService{}
	if err := s.OpenProject(p.ID, "vscode"); err == nil {
		t.Fatalf("spawn 失败应返回错误")
	}
	got, _ := loadProjects()
	if got[0].OpenCount != 0 || got[0].LastOpenedAt != 0 {
		t.Fatalf("spawn 失败不应更新统计: %+v", got[0])
	}
}

func TestOpenProjectMissing(t *testing.T) {
	withTempConfig(t)
	s := &ProjectService{}
	err := s.OpenProject("p-none", "")
	if err == nil || !strings.Contains(err.Error(), "项目不存在") {
		t.Fatalf("期望项目不存在,得到 %v", err)
	}
}

func TestOpenProjectDirGone(t *testing.T) {
	withTempConfig(t)
	base := t.TempDir()
	pdir := filepath.Join(base, "gone")
	os.MkdirAll(pdir, 0o755)
	s := &ProjectService{}
	p, err := s.AddProject(pdir, "", "")
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	os.Remove(pdir)
	err = s.OpenProject(p.ID, "")
	if err == nil || !strings.Contains(err.Error(), "目录不存在") {
		t.Fatalf("期望目录不存在,得到 %v", err)
	}
}

func TestDeleteProjectWithFiles(t *testing.T) {
	root := withTempConfig(t)
	s := &ProjectService{}
	dir := filepath.Join(root, "proj-with-files")
	if err := os.MkdirAll(filepath.Join(dir, "inner"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	p, err := s.AddProject(dir, "", "")
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	// 不勾选:仅移出列表,磁盘保留
	if err := s.DeleteProject(p.ID, false); err != nil {
		t.Fatalf("DeleteProject(false): %v", err)
	}
	if !fileExists(dir) {
		t.Fatalf("不勾选时目录应保留")
	}

	// 再登记,勾选:磁盘目录连同内容一起删除
	p2, err := s.AddProject(dir, "", "")
	if err != nil {
		t.Fatalf("AddProject 再次登记: %v", err)
	}
	if err := s.DeleteProject(p2.ID, true); err != nil {
		t.Fatalf("DeleteProject(true): %v", err)
	}
	if fileExists(dir) {
		t.Fatalf("勾选后目录应被删除")
	}
	ps, _ := loadProjects()
	if len(ps) != 0 {
		t.Fatalf("列表应为空: %+v", ps)
	}

	// 护栏:根目录与主目录拒绝删除,且列表项保留
	home, _ := userHomeDir()
	p3, err := s.AddProject(home, "", "")
	if err != nil {
		t.Fatalf("AddProject(home): %v", err)
	}
	if err := s.DeleteProject(p3.ID, true); err == nil || !strings.Contains(err.Error(), "拒绝删除") {
		t.Fatalf("主目录应被护栏拦截: %v", err)
	}
	if err := guardRemove("/"); err == nil {
		t.Fatalf("根目录应被护栏拦截")
	}
	ps, _ = loadProjects()
	if len(ps) != 1 || ps[0].ID != p3.ID {
		t.Fatalf("护栏拦截后列表项应保留: %+v", ps)
	}
}
