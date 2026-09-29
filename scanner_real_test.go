package main

import "testing"

// TestScanSystemEditorsRealPrint 真机扫描自检(无机器相关断言):
// go test . -run TestScanSystemEditorsRealPrint -v 可查看当前机器扫到的全部条目。
func TestScanSystemEditorsRealPrint(t *testing.T) {
	eds, err := scanSystemEditors()
	if err != nil {
		t.Fatalf("scanSystemEditors: %v", err)
	}
	for _, e := range eds {
		t.Logf("%-8s %-16s %-14s %s %s", e.Kind, e.ID, e.Name, e.Command, e.Args)
	}
	if len(eds) == 0 {
		t.Log("本机未扫到任何编辑器/终端/CLI(裸机器也允许)")
	}
}
