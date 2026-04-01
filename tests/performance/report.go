package performance

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

// PerformanceReport 性能测试报告
type PerformanceReport struct {
	GeneratedAt     time.Time          `json:"generated_at"`
	SystemInfo      SystemInfo         `json:"system_info"`
	TestResults     []TestResult       `json:"test_results"`
	Baselines       []Baseline         `json:"baselines"`
	Bottlenecks     []Bottleneck       `json:"bottlenecks"`
	Recommendations []Recommendation   `json:"recommendations"`
	Summary         PerformanceSummary `json:"summary"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	OS              string `json:"os"`
	Arch            string `json:"arch"`
	CPUCount        int    `json:"cpu_count"`
	GOVersion       string `json:"go_version"`
	TotalMemory     uint64 `json:"total_memory"`
	AvailableMemory uint64 `json:"available_memory"`
}

// TestResult 测试结果
type TestResult struct {
	Name           string                 `json:"name"`
	Category       string                 `json:"category"`
	Duration       time.Duration          `json:"duration"`
	Operations     int64                  `json:"operations"`
	OpsPerSecond   float64                `json:"ops_per_second"`
	AvgLatency     time.Duration          `json:"avg_latency"`
	P50Latency     time.Duration          `json:"p50_latency"`
	P95Latency     time.Duration          `json:"p95_latency"`
	P99Latency     time.Duration          `json:"p99_latency"`
	MaxLatency     time.Duration          `json:"max_latency"`
	MinLatency     time.Duration          `json:"min_latency"`
	MemoryAllocMB  float64                `json:"memory_alloc_mb"`
	MemoryTotalMB  float64                `json:"memory_total_mb"`
	CPUUsage       float64                `json:"cpu_usage"`
	SuccessRate    float64                `json:"success_rate"`
	ErrorCount     int64                  `json:"error_count"`
	CustomMetrics  map[string]interface{} `json:"custom_metrics"`
	Status         string                 `json:"status"` // pass, fail, warning
	BaselineDiff   float64                `json:"baseline_diff"` // 与基准线的差异百分比
}

// Baseline 基准线
type Baseline struct {
	Name          string        `json:"name"`
	Category      string        `json:"category"`
	OpsPerSecond  float64       `json:"ops_per_second"`
	AvgLatency    time.Duration `json:"avg_latency"`
	MemoryAllocMB float64       `json:"memory_alloc_mb"`
	SuccessRate   float64       `json:"success_rate"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// Bottleneck 瓶颈分析
type Bottleneck struct {
	Category       string   `json:"category"`
	Description    string   `json:"description"`
	Severity       string   `json:"severity"` // critical, high, medium, low
	Metrics        []string `json:"metrics"`
	Suggestions    []string `json:"suggestions"`
	AffectedTests  []string `json:"affected_tests"`
}

// Recommendation 优化建议
type Recommendation struct {
	Category       string   `json:"category"`
	Priority       string   `json:"priority"` // critical, high, medium, low
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	ExpectedImpact string   `json:"expected_impact"`
	Implementation string   `json:"implementation"`
	RelatedTests   []string `json:"related_tests"`
}

// PerformanceSummary 性能总结
type PerformanceSummary struct {
	TotalTests       int           `json:"total_tests"`
	PassedTests      int           `json:"passed_tests"`
	FailedTests      int           `json:"failed_tests"`
	WarningTests     int           `json:"warning_tests"`
	AvgOpsPerSecond  float64       `json:"avg_ops_per_second"`
	AvgLatency       time.Duration `json:"avg_latency"`
	TotalMemoryUsed  float64       `json:"total_memory_used_mb"`
	AvgCPUUsage      float64       `json:"avg_cpu_usage"`
	OverallScore     float64       `json:"overall_score"` // 0-100
	PerformanceGrade string        `json:"performance_grade"` // A, B, C, D, F
}

// ReportGenerator 报告生成器
type ReportGenerator struct {
	results     []TestResult
	baselines   []Baseline
	mu          sync.RWMutex
	outputDir   string
}

// NewReportGenerator 创建报告生成器
func NewReportGenerator(outputDir string) *ReportGenerator {
	return &ReportGenerator{
		results:   make([]TestResult, 0),
		baselines: make([]Baseline, 0),
		outputDir: outputDir,
	}
}

// AddTestResult 添加测试结果
func (g *ReportGenerator) AddTestResult(result TestResult) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.results = append(g.results, result)
}

// AddBaseline 添加基准线
func (g *ReportGenerator) AddBaseline(baseline Baseline) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.baselines = append(g.baselines, baseline)
}

// GenerateReport 生成报告
func (g *ReportGenerator) GenerateReport() (*PerformanceReport, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 收集系统信息
	sysInfo := g.collectSystemInfo()

	// 分析瓶颈
	bottlenecks := g.analyzeBottlenecks()

	// 生成优化建议
	recommendations := g.generateRecommendations()

	// 计算总结
	summary := g.calculateSummary()

	report := &PerformanceReport{
		GeneratedAt:     time.Now(),
		SystemInfo:      sysInfo,
		TestResults:     g.results,
		Baselines:       g.baselines,
		Bottlenecks:     bottlenecks,
		Recommendations: recommendations,
		Summary:         summary,
	}

	return report, nil
}

// SaveJSONReport 保存JSON格式报告
func (g *ReportGenerator) SaveJSONReport(report *PerformanceReport, filename string) error {
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filePath := filepath.Join(g.outputDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}

	return nil
}

// SaveHTMLReport 保存HTML格式报告
func (g *ReportGenerator) SaveHTMLReport(report *PerformanceReport, filename string) error {
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filePath := filepath.Join(g.outputDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	// 定义模板函数
	funcMap := template.FuncMap{
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}

	tmpl := template.Must(template.New("report").Funcs(funcMap).Parse(htmlTemplate))
	if err := tmpl.Execute(file, report); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// collectSystemInfo 收集系统信息
func (g *ReportGenerator) collectSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		CPUCount:        runtime.NumCPU(),
		GOVersion:       runtime.Version(),
		TotalMemory:     m.Sys,
		AvailableMemory: m.Sys - m.Alloc,
	}
}

// analyzeBottlenecks 分析瓶颈
func (g *ReportGenerator) analyzeBottlenecks() []Bottleneck {
	bottlenecks := make([]Bottleneck, 0)

	// 分析CPU瓶颈
	cpuBottleneck := g.analyzeCPUBottleneck()
	if cpuBottleneck != nil {
		bottlenecks = append(bottlenecks, *cpuBottleneck)
	}

	// 分析内存瓶颈
	memBottleneck := g.analyzeMemoryBottleneck()
	if memBottleneck != nil {
		bottlenecks = append(bottlenecks, *memBottleneck)
	}

	// 分析IO瓶颈
	ioBottleneck := g.analyzeIOBottleneck()
	if ioBottleneck != nil {
		bottlenecks = append(bottlenecks, *ioBottleneck)
	}

	// 分析并发瓶颈
	concurrencyBottleneck := g.analyzeConcurrencyBottleneck()
	if concurrencyBottleneck != nil {
		bottlenecks = append(bottlenecks, *concurrencyBottleneck)
	}

	return bottlenecks
}

// analyzeCPUBottleneck 分析CPU瓶颈
func (g *ReportGenerator) analyzeCPUBottleneck() *Bottleneck {
	var highCPUTests []string
	var totalCPUUsage float64

	for _, result := range g.results {
		if result.CPUUsage > 80 {
			highCPUTests = append(highCPUTests, result.Name)
		}
		totalCPUUsage += result.CPUUsage
	}

	if len(highCPUTests) == 0 {
		return nil
	}

	avgCPU := totalCPUUsage / float64(len(g.results))
	severity := "medium"
	if avgCPU > 90 {
		severity = "critical"
	} else if avgCPU > 80 {
		severity = "high"
	}

	return &Bottleneck{
		Category:    "CPU",
		Description: fmt.Sprintf("CPU使用率过高，平均使用率 %.2f%%", avgCPU),
		Severity:    severity,
		Metrics:     []string{"cpu_usage", "ops_per_second"},
		Suggestions: []string{
			"优化算法复杂度",
			"使用协程池减少协程创建开销",
			"减少不必要的计算",
			"使用缓存减少重复计算",
		},
		AffectedTests: highCPUTests,
	}
}

// analyzeMemoryBottleneck 分析内存瓶颈
func (g *ReportGenerator) analyzeMemoryBottleneck() *Bottleneck {
	var highMemTests []string
	var totalMemUsage float64

	for _, result := range g.results {
		if result.MemoryAllocMB > 1000 { // 超过1GB
			highMemTests = append(highMemTests, result.Name)
		}
		totalMemUsage += result.MemoryAllocMB
	}

	if len(highMemTests) == 0 {
		return nil
	}

	avgMem := totalMemUsage / float64(len(g.results))
	severity := "medium"
	if avgMem > 2000 {
		severity = "critical"
	} else if avgMem > 1000 {
		severity = "high"
	}

	return &Bottleneck{
		Category:    "Memory",
		Description: fmt.Sprintf("内存使用过高，平均使用 %.2f MB", avgMem),
		Severity:    severity,
		Metrics:     []string{"memory_alloc_mb", "memory_total_mb"},
		Suggestions: []string{
			"优化数据结构，减少内存占用",
			"使用对象池减少内存分配",
			"及时释放不再使用的对象",
			"使用流式处理减少内存缓存",
		},
		AffectedTests: highMemTests,
	}
}

// analyzeIOBottleneck 分析IO瓶颈
func (g *ReportGenerator) analyzeIOBottleneck() *Bottleneck {
	var slowTests []string
	var totalLatency time.Duration

	for _, result := range g.results {
		if result.AvgLatency > 100*time.Millisecond {
			slowTests = append(slowTests, result.Name)
		}
		totalLatency += result.AvgLatency
	}

	if len(slowTests) == 0 {
		return nil
	}

	avgLatency := totalLatency / time.Duration(len(g.results))
	severity := "medium"
	if avgLatency > 1*time.Second {
		severity = "critical"
	} else if avgLatency > 500*time.Millisecond {
		severity = "high"
	}

	return &Bottleneck{
		Category:    "IO",
		Description: fmt.Sprintf("IO延迟过高，平均延迟 %v", avgLatency),
		Severity:    severity,
		Metrics:     []string{"avg_latency", "p95_latency", "p99_latency"},
		Suggestions: []string{
			"使用批量操作减少IO次数",
			"增加缓存层",
			"使用异步IO",
			"优化数据库查询",
		},
		AffectedTests: slowTests,
	}
}

// analyzeConcurrencyBottleneck 分析并发瓶颈
func (g *ReportGenerator) analyzeConcurrencyBottleneck() *Bottleneck {
	var lowScalingTests []string

	for _, result := range g.results {
		// 检查并发扩展性
		if result.CustomMetrics != nil {
			if scaling, ok := result.CustomMetrics["scaling_efficiency"].(float64); ok {
				if scaling < 0.7 { // 扩展效率低于70%
					lowScalingTests = append(lowScalingTests, result.Name)
				}
			}
		}
	}

	if len(lowScalingTests) == 0 {
		return nil
	}

	return &Bottleneck{
		Category:    "Concurrency",
		Description: "并发扩展性不足，多核利用率低",
		Severity:    "medium",
		Metrics:     []string{"scaling_efficiency", "ops_per_second"},
		Suggestions: []string{
			"减少锁竞争",
			"使用无锁数据结构",
			"优化协程调度",
			"使用分片技术减少共享资源",
		},
		AffectedTests: lowScalingTests,
	}
}

// generateRecommendations 生成优化建议
func (g *ReportGenerator) generateRecommendations() []Recommendation {
	recommendations := make([]Recommendation, 0)

	// 根据测试结果生成建议
	for _, result := range g.results {
		// 低吞吐量建议
		if result.OpsPerSecond < 1000 && result.Category == "collector" {
			recommendations = append(recommendations, Recommendation{
				Category:       "Performance",
				Priority:       "high",
				Title:          "提高采集吞吐量",
				Description:    fmt.Sprintf("测试 %s 的吞吐量仅为 %.0f ops/s，建议优化", result.Name, result.OpsPerSecond),
				ExpectedImpact: "吞吐量提升50%-100%",
				Implementation: "使用批量采集、优化数据结构、减少内存分配",
				RelatedTests:   []string{result.Name},
			})
		}

		// 高延迟建议
		if result.AvgLatency > 100*time.Millisecond && result.Category == "query" {
			recommendations = append(recommendations, Recommendation{
				Category:       "Performance",
				Priority:       "high",
				Title:          "降低查询延迟",
				Description:    fmt.Sprintf("测试 %s 的平均延迟为 %v，建议优化", result.Name, result.AvgLatency),
				ExpectedImpact: "延迟降低50%-80%",
				Implementation: "添加索引、优化查询计划、使用缓存",
				RelatedTests:   []string{result.Name},
			})
		}

		// 高内存使用建议
		if result.MemoryAllocMB > 500 {
			recommendations = append(recommendations, Recommendation{
				Category:       "Memory",
				Priority:       "medium",
				Title:          "优化内存使用",
				Description:    fmt.Sprintf("测试 %s 的内存分配为 %.2f MB，建议优化", result.Name, result.MemoryAllocMB),
				ExpectedImpact: "内存使用降低30%-50%",
				Implementation: "使用对象池、优化数据结构、及时释放资源",
				RelatedTests:   []string{result.Name},
			})
		}

		// 低成功率建议
		if result.SuccessRate < 99 {
			recommendations = append(recommendations, Recommendation{
				Category:       "Reliability",
				Priority:       "critical",
				Title:          "提高成功率",
				Description:    fmt.Sprintf("测试 %s 的成功率为 %.2f%%，低于99%%", result.Name, result.SuccessRate),
				ExpectedImpact: "成功率提升至99.9%以上",
				Implementation: "增加错误处理、实现重试机制、优化超时设置",
				RelatedTests:   []string{result.Name},
			})
		}
	}

	// 添加通用建议
	recommendations = append(recommendations, Recommendation{
		Category:       "Architecture",
		Priority:       "medium",
		Title:          "实施性能监控",
		Description:    "建议在生产环境实施持续性能监控",
		ExpectedImpact: "及时发现性能问题，减少故障时间",
		Implementation: "集成Prometheus监控、配置告警规则、定期性能测试",
		RelatedTests:   []string{},
	})

	return recommendations
}

// calculateSummary 计算总结
func (g *ReportGenerator) calculateSummary() PerformanceSummary {
	var passed, failed, warning int
	var totalOps, totalLatency, totalMem, totalCPU float64

	for _, result := range g.results {
		switch result.Status {
		case "pass":
			passed++
		case "fail":
			failed++
		case "warning":
			warning++
		}

		totalOps += result.OpsPerSecond
		totalLatency += float64(result.AvgLatency)
		totalMem += result.MemoryAllocMB
		totalCPU += result.CPUUsage
	}

	count := float64(len(g.results))
	if count == 0 {
		count = 1
	}

	avgOps := totalOps / count
	avgLatency := time.Duration(totalLatency / count)
	avgMem := totalMem / count
	avgCPU := totalCPU / count

	// 计算综合评分
	score := g.calculateScore(avgOps, avgLatency, avgMem, avgCPU, passed, failed, warning)

	// 确定性能等级
	grade := "F"
	if score >= 90 {
		grade = "A"
	} else if score >= 80 {
		grade = "B"
	} else if score >= 70 {
		grade = "C"
	} else if score >= 60 {
		grade = "D"
	}

	return PerformanceSummary{
		TotalTests:       len(g.results),
		PassedTests:      passed,
		FailedTests:      failed,
		WarningTests:     warning,
		AvgOpsPerSecond:  avgOps,
		AvgLatency:       avgLatency,
		TotalMemoryUsed:  totalMem,
		AvgCPUUsage:      avgCPU,
		OverallScore:     score,
		PerformanceGrade: grade,
	}
}

// calculateScore 计算综合评分
func (g *ReportGenerator) calculateScore(avgOps float64, avgLatency time.Duration, avgMem, avgCPU float64, passed, failed, warning int) float64 {
	score := 100.0

	// 扣除失败测试的分数
	total := passed + failed + warning
	if total > 0 {
		score -= float64(failed) / float64(total) * 30
		score -= float64(warning) / float64(total) * 10
	}

	// 根据性能指标扣分
	if avgOps < 1000 {
		score -= 10
	}
	if avgLatency > 100*time.Millisecond {
		score -= 10
	}
	if avgMem > 1000 {
		score -= 10
	}
	if avgCPU > 80 {
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	return score
}

// CompareWithBaseline 与基准线对比
func (g *ReportGenerator) CompareWithBaseline(result *TestResult, baseline *Baseline) float64 {
	if baseline == nil {
		return 0
	}

	// 计算性能差异百分比
	opsDiff := (result.OpsPerSecond - baseline.OpsPerSecond) / baseline.OpsPerSecond * 100
	latencyDiff := float64(result.AvgLatency-baseline.AvgLatency) / float64(baseline.AvgLatency) * 100
	memDiff := (result.MemoryAllocMB - baseline.MemoryAllocMB) / baseline.MemoryAllocMB * 100

	// 综合差异（负数表示性能下降）
	return (opsDiff - latencyDiff - memDiff) / 3
}

// HTML模板
const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>新能源监控系统 - 性能测试报告</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
            color: #333;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 15px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        .header p {
            font-size: 1.1em;
            opacity: 0.9;
        }
        .content {
            padding: 40px;
        }
        .section {
            margin-bottom: 40px;
        }
        .section-title {
            font-size: 1.8em;
            color: #667eea;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 3px solid #667eea;
        }
        .summary-cards {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .card {
            background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
            padding: 25px;
            border-radius: 10px;
            text-align: center;
            transition: transform 0.3s;
        }
        .card:hover {
            transform: translateY(-5px);
        }
        .card-value {
            font-size: 2.5em;
            font-weight: bold;
            color: #667eea;
            margin: 10px 0;
        }
        .card-label {
            font-size: 1em;
            color: #666;
        }
        .grade {
            display: inline-block;
            padding: 10px 30px;
            border-radius: 50px;
            font-size: 2em;
            font-weight: bold;
            color: white;
        }
        .grade-A { background: #27ae60; }
        .grade-B { background: #3498db; }
        .grade-C { background: #f39c12; }
        .grade-D { background: #e67e22; }
        .grade-F { background: #e74c3c; }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
            background: white;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            border-radius: 10px;
            overflow: hidden;
        }
        th, td {
            padding: 15px;
            text-align: left;
            border-bottom: 1px solid #eee;
        }
        th {
            background: #667eea;
            color: white;
            font-weight: 600;
        }
        tr:hover {
            background: #f5f7fa;
        }
        .status-pass { color: #27ae60; font-weight: bold; }
        .status-fail { color: #e74c3c; font-weight: bold; }
        .status-warning { color: #f39c12; font-weight: bold; }
        .bottleneck {
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
        }
        .bottleneck.critical {
            background: #f8d7da;
            border-left-color: #dc3545;
        }
        .bottleneck.high {
            background: #fff3cd;
            border-left-color: #ffc107;
        }
        .recommendation {
            background: #d1ecf1;
            border-left: 4px solid #17a2b8;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
        }
        .recommendation.critical {
            background: #f8d7da;
            border-left-color: #dc3545;
        }
        .recommendation.high {
            background: #fff3cd;
            border-left-color: #ffc107;
        }
        .system-info {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            background: #f5f7fa;
            padding: 20px;
            border-radius: 10px;
        }
        .system-info-item {
            text-align: center;
        }
        .system-info-label {
            font-size: 0.9em;
            color: #666;
        }
        .system-info-value {
            font-size: 1.1em;
            font-weight: bold;
            color: #333;
        }
        .chart-placeholder {
            background: #f5f7fa;
            padding: 40px;
            text-align: center;
            border-radius: 10px;
            color: #666;
        }
        .footer {
            background: #f5f7fa;
            padding: 20px;
            text-align: center;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>新能源监控系统</h1>
            <p>性能测试报告</p>
            <p style="margin-top: 10px; font-size: 0.9em;">生成时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
        </div>

        <div class="content">
            <!-- 性能等级 -->
            <div class="section">
                <h2 class="section-title">性能等级</h2>
                <div style="text-align: center; padding: 20px;">
                    <span class="grade grade-{{.Summary.PerformanceGrade}}">{{.Summary.PerformanceGrade}}</span>
                    <p style="margin-top: 15px; font-size: 1.2em;">综合评分: {{printf "%.1f" .Summary.OverallScore}}/100</p>
                </div>
            </div>

            <!-- 性能总结 -->
            <div class="section">
                <h2 class="section-title">性能总结</h2>
                <div class="summary-cards">
                    <div class="card">
                        <div class="card-label">总测试数</div>
                        <div class="card-value">{{.Summary.TotalTests}}</div>
                    </div>
                    <div class="card">
                        <div class="card-label">通过测试</div>
                        <div class="card-value" style="color: #27ae60;">{{.Summary.PassedTests}}</div>
                    </div>
                    <div class="card">
                        <div class="card-label">失败测试</div>
                        <div class="card-value" style="color: #e74c3c;">{{.Summary.FailedTests}}</div>
                    </div>
                    <div class="card">
                        <div class="card-label">警告测试</div>
                        <div class="card-value" style="color: #f39c12;">{{.Summary.WarningTests}}</div>
                    </div>
                    <div class="card">
                        <div class="card-label">平均吞吐量</div>
                        <div class="card-value">{{printf "%.0f" .Summary.AvgOpsPerSecond}}</div>
                        <div class="card-label">ops/s</div>
                    </div>
                    <div class="card">
                        <div class="card-label">平均延迟</div>
                        <div class="card-value">{{.Summary.AvgLatency}}</div>
                    </div>
                    <div class="card">
                        <div class="card-label">总内存使用</div>
                        <div class="card-value">{{printf "%.2f" .Summary.TotalMemoryUsed}}</div>
                        <div class="card-label">MB</div>
                    </div>
                    <div class="card">
                        <div class="card-label">平均CPU使用率</div>
                        <div class="card-value">{{printf "%.1f" .Summary.AvgCPUUsage}}</div>
                        <div class="card-label">%</div>
                    </div>
                </div>
            </div>

            <!-- 系统信息 -->
            <div class="section">
                <h2 class="section-title">系统信息</h2>
                <div class="system-info">
                    <div class="system-info-item">
                        <div class="system-info-label">操作系统</div>
                        <div class="system-info-value">{{.SystemInfo.OS}}</div>
                    </div>
                    <div class="system-info-item">
                        <div class="system-info-label">架构</div>
                        <div class="system-info-value">{{.SystemInfo.Arch}}</div>
                    </div>
                    <div class="system-info-item">
                        <div class="system-info-label">CPU核心数</div>
                        <div class="system-info-value">{{.SystemInfo.CPUCount}}</div>
                    </div>
                    <div class="system-info-item">
                        <div class="system-info-label">Go版本</div>
                        <div class="system-info-value">{{.SystemInfo.GOVersion}}</div>
                    </div>
                    <div class="system-info-item">
                        <div class="system-info-label">总内存</div>
                        <div class="system-info-value">{{printf "%.2f" (div (float64 .SystemInfo.TotalMemory) 1073741824.0)}} GB</div>
                    </div>
                    <div class="system-info-item">
                        <div class="system-info-label">可用内存</div>
                        <div class="system-info-value">{{printf "%.2f" (div (float64 .SystemInfo.AvailableMemory) 1073741824.0)}} GB</div>
                    </div>
                </div>
            </div>

            <!-- 测试结果详情 -->
            <div class="section">
                <h2 class="section-title">测试结果详情</h2>
                <table>
                    <thead>
                        <tr>
                            <th>测试名称</th>
                            <th>类别</th>
                            <th>吞吐量 (ops/s)</th>
                            <th>平均延迟</th>
                            <th>P95延迟</th>
                            <th>内存 (MB)</th>
                            <th>CPU (%)</th>
                            <th>成功率</th>
                            <th>状态</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .TestResults}}
                        <tr>
                            <td>{{.Name}}</td>
                            <td>{{.Category}}</td>
                            <td>{{printf "%.0f" .OpsPerSecond}}</td>
                            <td>{{.AvgLatency}}</td>
                            <td>{{.P95Latency}}</td>
                            <td>{{printf "%.2f" .MemoryAllocMB}}</td>
                            <td>{{printf "%.1f" .CPUUsage}}</td>
                            <td>{{printf "%.2f" .SuccessRate}}%</td>
                            <td class="status-{{.Status}}">{{.Status}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>

            <!-- 瓶颈分析 -->
            <div class="section">
                <h2 class="section-title">瓶颈分析</h2>
                {{if .Bottlenecks}}
                    {{range .Bottlenecks}}
                    <div class="bottleneck {{.Severity}}">
                        <h3>{{.Category}} - {{.Severity}}</h3>
                        <p>{{.Description}}</p>
                        <p><strong>建议:</strong></p>
                        <ul>
                            {{range .Suggestions}}
                            <li>{{.}}</li>
                            {{end}}
                        </ul>
                    </div>
                    {{end}}
                {{else}}
                    <p style="color: #27ae60; font-size: 1.1em;">未发现明显性能瓶颈</p>
                {{end}}
            </div>

            <!-- 优化建议 -->
            <div class="section">
                <h2 class="section-title">优化建议</h2>
                {{range .Recommendations}}
                <div class="recommendation {{.Priority}}">
                    <h3>{{.Title}} ({{.Priority}})</h3>
                    <p>{{.Description}}</p>
                    <p><strong>预期影响:</strong> {{.ExpectedImpact}}</p>
                    <p><strong>实施方法:</strong> {{.Implementation}}</p>
                </div>
                {{end}}
            </div>

            <!-- 性能图表占位符 -->
            <div class="section">
                <h2 class="section-title">性能图表</h2>
                <div class="chart-placeholder">
                    <p>性能图表将在完整版本中提供</p>
                    <p>包括: 吞吐量趋势图、延迟分布图、内存使用图、CPU使用图</p>
                </div>
            </div>
        </div>

        <div class="footer">
            <p>新能源监控系统 - 性能测试报告</p>
            <p>生成时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
        </div>
    </div>
</body>
</html>`

// ExampleReportGeneration 示例：生成性能测试报告
func ExampleReportGeneration() {
	// 创建报告生成器
	generator := NewReportGenerator("./reports")

	// 添加测试结果
	generator.AddTestResult(TestResult{
		Name:          "BenchmarkCollectorMillionPoints",
		Category:      "collector",
		Duration:      10 * time.Second,
		Operations:    1000,
		OpsPerSecond:  100000,
		AvgLatency:    10 * time.Millisecond,
		P50Latency:    8 * time.Millisecond,
		P95Latency:    20 * time.Millisecond,
		P99Latency:    50 * time.Millisecond,
		MaxLatency:    100 * time.Millisecond,
		MinLatency:    5 * time.Millisecond,
		MemoryAllocMB: 512.5,
		MemoryTotalMB: 1024.0,
		CPUUsage:      75.5,
		SuccessRate:   99.9,
		ErrorCount:    1,
		Status:        "pass",
	})

	// 添加基准线
	generator.AddBaseline(Baseline{
		Name:          "Baseline_Collector",
		Category:      "collector",
		OpsPerSecond:  80000,
		AvgLatency:    15 * time.Millisecond,
		MemoryAllocMB: 600.0,
		SuccessRate:   99.5,
		UpdatedAt:     time.Now(),
	})

	// 生成报告
	report, err := generator.GenerateReport()
	if err != nil {
		fmt.Printf("Failed to generate report: %v\n", err)
		return
	}

	// 保存JSON报告
	if err := generator.SaveJSONReport(report, "performance_report.json"); err != nil {
		fmt.Printf("Failed to save JSON report: %v\n", err)
	}

	// 保存HTML报告
	if err := generator.SaveHTMLReport(report, "performance_report.html"); err != nil {
		fmt.Printf("Failed to save HTML report: %v\n", err)
	}

	fmt.Println("Performance report generated successfully!")
}

// TestReportGeneration 测试报告生成
func TestReportGeneration(t *testing.T) {
	generator := NewReportGenerator(t.TempDir())

	// 添加测试数据
	generator.AddTestResult(TestResult{
		Name:          "Test1",
		Category:      "collector",
		OpsPerSecond:   10000,
		AvgLatency:     10 * time.Millisecond,
		MemoryAllocMB:  100,
		CPUUsage:       50,
		SuccessRate:    99.5,
		Status:         "pass",
	})

	generator.AddTestResult(TestResult{
		Name:          "Test2",
		Category:      "query",
		OpsPerSecond:   5000,
		AvgLatency:     200 * time.Millisecond,
		MemoryAllocMB:  500,
		CPUUsage:       85,
		SuccessRate:    98.0,
		Status:         "warning",
	})

	// 生成报告
	report, err := generator.GenerateReport()
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// 验证报告内容
	if report.Summary.TotalTests != 2 {
		t.Errorf("Expected 2 tests, got %d", report.Summary.TotalTests)
	}

	if report.Summary.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", report.Summary.PassedTests)
	}

	if report.Summary.WarningTests != 1 {
		t.Errorf("Expected 1 warning test, got %d", report.Summary.WarningTests)
	}

	// 保存JSON报告
	if err := generator.SaveJSONReport(report, "test_report.json"); err != nil {
		t.Errorf("Failed to save JSON report: %v", err)
	}

	// 保存HTML报告
	if err := generator.SaveHTMLReport(report, "test_report.html"); err != nil {
		t.Errorf("Failed to save HTML report: %v", err)
	}
}
