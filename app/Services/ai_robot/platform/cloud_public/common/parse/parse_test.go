package parse

import (
	"strings"
	"testing"
	"time"
)

func TestExtractProjectFiltersFromQuestion(t *testing.T) {
	got := ExtractProjectFiltersFromQuestion(`帮我查项目编号 MWXS-25-10500-a 的产品名称为 "蛋白组" 的项目`)
	if got.Number != "MWXS-25-10500-a" {
		t.Fatalf("unexpected project number: %q", got.Number)
	}
	if got.WorkflowNameCN != "蛋白组" {
		t.Fatalf("unexpected workflow name: %q", got.WorkflowNameCN)
	}
	if got.IsFilterTime != 2 {
		t.Fatalf("unexpected is_filter_time: %d", got.IsFilterTime)
	}
}

func TestExtractTaskFiltersFromQuestion(t *testing.T) {
	got := ExtractTaskFiltersFromQuestion(`查询项目名称为"玉米代谢组"的运行中任务`, "86ce0f8b-ee05-47af-9495-6a48addd8a39")
	if got.UUID == "" || got.StatusValue != "running" || got.ProjectName != "玉米代谢组" {
		t.Fatalf("unexpected task filters: %+v", got)
	}
}

func TestExtractTaskFiltersWithAliasAndDate(t *testing.T) {
	got := ExtractTaskFiltersFromQuestion(`查最近7天工具名为"KEGG"的工单号 86ce0f8b-ee05-47af-9495-6a48addd8a39`, "86ce0f8b-ee05-47af-9495-6a48addd8a39")
	if got.ToolName != "KEGG" {
		t.Fatalf("unexpected tool name: %q", got.ToolName)
	}
	if got.CreatedAtStart == "" || got.CreatedAtEnd != time.Now().Format("2006-01-02") {
		t.Fatalf("unexpected time range: %+v", got)
	}
}

func TestExtractProjectArticleFiltersFromQuestion(t *testing.T) {
	got := ExtractProjectArticleFiltersFromQuestion(`帮我找期刊名是"Nature"且标签为"植物测试"的项目文章`)
	if len(got.SearchFilter) != 2 {
		t.Fatalf("unexpected search filters count: %d", len(got.SearchFilter))
	}
}

func TestExtractProjectArticleFiltersExtended(t *testing.T) {
	got := ExtractProjectArticleFiltersFromQuestion(`找物种名称是"水稻"、样本类型是"叶片"、地区是"华南"的项目文章`)
	if len(got.SearchFilter) != 3 {
		t.Fatalf("unexpected search filters count: %d", len(got.SearchFilter))
	}
}

func TestTaskUUIDGuidanceForQuestion(t *testing.T) {
	got := TaskUUIDGuidanceForQuestion(`帮我下载项目编号 MWXS-25-10500-a 的结题报告`)
	if !strings.Contains(got, "更像是在查项目资料") {
		t.Fatalf("unexpected guidance: %s", got)
	}
}

func TestProjectMatchGuidanceForQuestion(t *testing.T) {
	got := ProjectMatchGuidanceForQuestion(`这个工单跑完没`)
	if !strings.Contains(got, "更像是在查任务") {
		t.Fatalf("unexpected guidance: %s", got)
	}
}
