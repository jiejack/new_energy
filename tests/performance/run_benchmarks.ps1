# 新能源监控系统性能测试脚本 (Windows PowerShell)
# 用于运行所有性能测试并生成报告

param(
    [Parameter(Position=0)]
    [ValidateSet("all", "collector", "query", "stress", "cpu-profile", "mem-profile", "report")]
    [string]$TestType = "all"
)

# 颜色函数
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

# 项目路径
$ProjectRoot = Split-Path -Parent $PSScriptRoot | Split-Path -Parent
$TestDir = Join-Path $ProjectRoot "tests\performance"
$ReportDir = Join-Path $ProjectRoot "reports\performance"

# 创建报告目录
if (-not (Test-Path $ReportDir)) {
    New-Item -ItemType Directory -Path $ReportDir -Force | Out-Null
}

Write-ColorOutput Cyan "========================================"
Write-ColorOutput Cyan "新能源监控系统 - 性能测试"
Write-ColorOutput Cyan "========================================"
Write-Output ""

# 函数：运行性能测试
function Run-Benchmark {
    param(
        [string]$Name,
        [string]$Pattern
    )

    $OutputFile = Join-Path $ReportDir "$Name.txt"

    Write-ColorOutput Yellow "运行测试: $Name"
    Write-Output "开始时间: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"

    # 运行benchmark
    Set-Location $ProjectRoot
    go test -bench="$Pattern" `
        -benchmem `
        -benchtime=3s `
        -run=^$ `
        -timeout=30m `
        ./tests/performance/... | Tee-Object -FilePath $OutputFile

    Write-ColorOutput Green "测试完成: $Name"
    Write-Output "报告保存至: $OutputFile"
    Write-Output ""
}

# 函数：生成CPU性能分析
function Generate-CPUProfile {
    param(
        [string]$Name,
        [string]$Pattern
    )

    $ProfileFile = Join-Path $ReportDir "${Name}_cpu.prof"

    Write-ColorOutput Yellow "生成CPU性能分析: $Name"

    Set-Location $ProjectRoot
    go test -bench="$Pattern" `
        -cpuprofile="$ProfileFile" `
        -run=^$ `
        -timeout=10m `
        ./tests/performance/...

    Write-ColorOutput Green "CPU性能分析完成: $ProfileFile"
    Write-Output ""
}

# 函数：生成内存性能分析
function Generate-MemProfile {
    param(
        [string]$Name,
        [string]$Pattern
    )

    $ProfileFile = Join-Path $ReportDir "${Name}_mem.prof"

    Write-ColorOutput Yellow "生成内存性能分析: $Name"

    Set-Location $ProjectRoot
    go test -bench="$Pattern" `
        -memprofile="$ProfileFile" `
        -run=^$ `
        -timeout=10m `
        ./tests/performance/...

    Write-ColorOutput Green "内存性能分析完成: $ProfileFile"
    Write-Output ""
}

# 函数：生成性能报告
function Generate-Report {
    Write-ColorOutput Yellow "生成性能测试报告..."

    Set-Location $ProjectRoot
    go test -v -run=TestReportGeneration ./tests/performance/...

    Write-ColorOutput Green "性能测试报告生成完成"
    Write-Output ""
}

# 函数：查看性能分析
function View-Profile {
    param([string]$ProfileFile)

    if (-not (Test-Path $ProfileFile)) {
        Write-ColorOutput Red "性能分析文件不存在: $ProfileFile"
        Write-Output "请先运行性能分析生成"
        return
    }

    Write-ColorOutput Yellow "启动pprof web界面..."
    go tool pprof -http=:8080 $ProfileFile
}

# 主逻辑
switch ($TestType) {
    "all" {
        Write-ColorOutput Green "运行所有性能测试..."
        Run-Benchmark -Name "collector_bench" -Pattern "BenchmarkCollector"
        Run-Benchmark -Name "query_bench" -Pattern "BenchmarkQuery"
        Run-Benchmark -Name "stress_test" -Pattern "BenchmarkAPI|BenchmarkWebSocket|BenchmarkDatabase|BenchmarkMessage"
        Generate-Report
    }

    "collector" {
        Run-Benchmark -Name "collector_bench" -Pattern "BenchmarkCollector"
    }

    "query" {
        Run-Benchmark -Name "query_bench" -Pattern "BenchmarkQuery"
    }

    "stress" {
        Run-Benchmark -Name "stress_test" -Pattern "BenchmarkAPI|BenchmarkWebSocket|BenchmarkDatabase|BenchmarkMessage"
    }

    "cpu-profile" {
        Write-ColorOutput Yellow "选择CPU性能分析类型:"
        Write-Output "1) 采集器CPU分析"
        Write-Output "2) 查询CPU分析"
        Write-Output "3) 压力测试CPU分析"
        $Choice = Read-Host "请选择"

        switch ($Choice) {
            "1" { Generate-CPUProfile -Name "collector" -Pattern "BenchmarkCollectorMillionPoints" }
            "2" { Generate-CPUProfile -Name "query" -Pattern "BenchmarkQueryMillionRecords" }
            "3" { Generate-CPUProfile -Name "stress" -Pattern "BenchmarkAPIPressure" }
            default { Write-ColorOutput Red "无效选项" }
        }
    }

    "mem-profile" {
        Write-ColorOutput Yellow "选择内存性能分析类型:"
        Write-Output "1) 采集器内存分析"
        Write-Output "2) 查询内存分析"
        Write-Output "3) 压力测试内存分析"
        $Choice = Read-Host "请选择"

        switch ($Choice) {
            "1" { Generate-MemProfile -Name "collector" -Pattern "BenchmarkCollectorMemoryUsage" }
            "2" { Generate-MemProfile -Name "query" -Pattern "BenchmarkQueryMemoryAllocation" }
            "3" { Generate-MemProfile -Name "stress" -Pattern "BenchmarkMemoryPressure" }
            default { Write-ColorOutput Red "无效选项" }
        }
    }

    "report" {
        Generate-Report
    }
}

Write-Output ""
Write-ColorOutput Green "性能测试完成！"
Write-Output "报告目录: $ReportDir"
Write-Output ""
Write-Output "使用方法:"
Write-Output "  .\run_benchmarks.ps1 -TestType all           # 运行所有测试"
Write-Output "  .\run_benchmarks.ps1 -TestType collector     # 采集性能测试"
Write-Output "  .\run_benchmarks.ps1 -TestType query         # 查询性能测试"
Write-Output "  .\run_benchmarks.ps1 -TestType stress        # 压力测试"
Write-Output "  .\run_benchmarks.ps1 -TestType cpu-profile   # CPU性能分析"
Write-Output "  .\run_benchmarks.ps1 -TestType mem-profile   # 内存性能分析"
Write-Output "  .\run_benchmarks.ps1 -TestType report        # 生成报告"
