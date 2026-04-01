# 新能源监控系统 - 性能测试脚本 (Windows PowerShell)
# 用于一键运行所有性能测试并生成报告

param(
    [Parameter(Position=0)]
    [ValidateSet("all", "backend", "frontend", "api", "db", "memory", "collector", "query", "stress", "report", "baseline", "compare")]
    [string]$TestType = "all",
    
    [Parameter(Position=1)]
    [switch]$Verbose,
    
    [switch]$Help
)

# 配置
$PROJECT_ROOT = Split-Path -Parent $PSScriptRoot | Split-Path -Parent
$REPORT_DIR = Join-Path $PROJECT_ROOT "reports\performance"
$TIMESTAMP = Get-Date -Format "yyyyMMdd_HHmmss"
$BASELINE_DIR = Join-Path $PROJECT_ROOT "reports\baseline"

# 创建报告目录
New-Item -ItemType Directory -Force -Path $REPORT_DIR | Out-Null
New-Item -ItemType Directory -Force -Path $BASELINE_DIR | Out-Null

# 颜色函数
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

# 显示帮助信息
function Show-Help {
    Write-Output "用法: .\run_perf_tests.ps1 [测试类型] [选项]"
    Write-Output ""
    Write-Output "测试类型:"
    Write-Output "  all         运行所有性能测试 (默认)"
    Write-Output "  backend     仅运行后端性能测试"
    Write-Output "  frontend    仅运行前端性能测试"
    Write-Output "  api         仅运行API性能测试"
    Write-Output "  db          仅运行数据库性能测试"
    Write-Output "  memory      仅运行内存泄漏测试"
    Write-Output "  collector   仅运行采集器性能测试"
    Write-Output "  query       仅运行查询性能测试"
    Write-Output "  stress      仅运行压力测试"
    Write-Output "  report      仅生成性能报告"
    Write-Output "  baseline    设置当前性能为基线"
    Write-Output "  compare     与基线性能对比"
    Write-Output ""
    Write-Output "选项:"
    Write-Output "  -Verbose    显示详细输出"
    Write-Output "  -Help       显示帮助信息"
    Write-Output ""
}

# 检查依赖
function Check-Dependencies {
    Write-ColorOutput Yellow "检查依赖..."
    
    # 检查Go
    if (Get-Command go -ErrorAction SilentlyContinue) {
        $goVersion = (go version).Split()[2]
        Write-ColorOutput Green "✓ Go $goVersion"
    } else {
        Write-ColorOutput Red "错误: 未找到Go环境"
        exit 1
    }
    
    # 检查Node.js
    if (Get-Command node -ErrorAction SilentlyContinue) {
        $nodeVersion = node -v
        Write-ColorOutput Green "✓ Node $nodeVersion"
    } else {
        Write-ColorOutput Yellow "! Node.js 未安装，跳过前端测试"
    }
    
    # 检查k6
    if (Get-Command k6 -ErrorAction SilentlyContinue) {
        Write-ColorOutput Green "✓ k6 installed"
    } else {
        Write-ColorOutput Yellow "! k6 未安装，部分负载测试将跳过"
    }
    
    Write-Output ""
}

# 运行后端性能测试
function Run-BackendTests {
    Write-ColorOutput Cyan "========================================"
    Write-ColorOutput Cyan "  运行后端性能测试"
    Write-ColorOutput Cyan "========================================"
    Write-Output ""
    
    Set-Location $PROJECT_ROOT
    
    # API性能测试
    if ($TestType -in @("all", "backend", "api")) {
        Write-ColorOutput Yellow "运行API性能测试..."
        go test -bench=BenchmarkAPI -benchmem -benchtime=5s ./tests/performance/... 2>&1 | Tee-Object -FilePath "$REPORT_DIR\api_perf_$TIMESTAMP.txt"
        Write-Output ""
    }
    
    # 数据库性能测试
    if ($TestType -in @("all", "backend", "db")) {
        Write-ColorOutput Yellow "运行数据库性能测试..."
        go test -bench=BenchmarkDatabase -benchmem -benchtime=5s ./tests/performance/... 2>&1 | Tee-Object -FilePath "$REPORT_DIR\db_perf_$TIMESTAMP.txt"
        Write-Output ""
    }
    
    # 内存泄漏测试
    if ($TestType -in @("all", "backend", "memory")) {
        Write-ColorOutput Yellow "运行内存泄漏测试..."
        go test -bench=BenchmarkMemory -benchmem -benchtime=10s ./tests/performance/... 2>&1 | Tee-Object -FilePath "$REPORT_DIR\memory_leak_$TIMESTAMP.txt"
        Write-Output ""
    }
    
    # 采集器性能测试
    if ($TestType -in @("all", "backend", "collector")) {
        Write-ColorOutput Yellow "运行采集器性能测试..."
        go test -bench=BenchmarkCollector -benchmem -benchtime=5s ./tests/performance/... 2>&1 | Tee-Object -FilePath "$REPORT_DIR\collector_bench_$TIMESTAMP.txt"
        Write-Output ""
    }
    
    # 查询性能测试
    if ($TestType -in @("all", "backend", "query")) {
        Write-ColorOutput Yellow "运行查询性能测试..."
        go test -bench=BenchmarkQuery -benchmem -benchtime=5s ./tests/performance/... 2>&1 | Tee-Object -FilePath "$REPORT_DIR\query_bench_$TIMESTAMP.txt"
        Write-Output ""
    }
    
    # 压力测试
    if ($TestType -in @("all", "backend", "stress")) {
        Write-ColorOutput Yellow "运行压力测试..."
        go test -bench=BenchmarkAPIPressure -benchmem -benchtime=10s ./tests/performance/... 2>&1 | Tee-Object -FilePath "$REPORT_DIR\stress_test_$TIMESTAMP.txt"
        Write-Output ""
    }
    
    # 生成CPU性能分析
    Write-ColorOutput Yellow "生成CPU性能分析..."
    go test -bench=BenchmarkCollectorMillionPoints -cpuprofile="$REPORT_DIR\cpu.prof" ./tests/performance/... 2>&1 | Out-Null
    Write-ColorOutput Green "✓ CPU性能分析已生成: $REPORT_DIR\cpu.prof"
    
    # 生成内存性能分析
    Write-ColorOutput Yellow "生成内存性能分析..."
    go test -bench=BenchmarkCollectorMemoryUsage -memprofile="$REPORT_DIR\mem.prof" ./tests/performance/... 2>&1 | Out-Null
    Write-ColorOutput Green "✓ 内存性能分析已生成: $REPORT_DIR\mem.prof"
    Write-Output ""
}

# 运行前端性能测试
function Run-FrontendTests {
    Write-ColorOutput Cyan "========================================"
    Write-ColorOutput Cyan "  运行前端性能测试"
    Write-ColorOutput Cyan "========================================"
    Write-Output ""
    
    $webDir = Join-Path $PROJECT_ROOT "web"
    Set-Location $webDir
    
    # 检查是否安装了依赖
    if (-not (Test-Path "node_modules")) {
        Write-ColorOutput Yellow "安装前端依赖..."
        npm install
    }
    
    # 组件渲染性能测试
    if ($TestType -in @("all", "frontend")) {
        Write-ColorOutput Yellow "运行组件渲染性能测试..."
        npm run test:perf -- tests/performance/component-perf.test.ts 2>&1 | Tee-Object -FilePath "$REPORT_DIR\component_perf_$TIMESTAMP.txt"
        Write-Output ""
        
        Write-ColorOutput Yellow "运行状态更新性能测试..."
        npm run test:perf -- tests/performance/state-perf.test.ts 2>&1 | Tee-Object -FilePath "$REPORT_DIR\state_perf_$TIMESTAMP.txt"
        Write-Output ""
    }
    
    Set-Location $PROJECT_ROOT
}

# 生成性能报告
function Generate-Report {
    Write-ColorOutput Cyan "========================================"
    Write-ColorOutput Cyan "  生成性能测试报告"
    Write-ColorOutput Cyan "========================================"
    Write-Output ""
    
    Set-Location $PROJECT_ROOT
    
    # 运行报告生成
    go test -run=TestGeneratePerformanceReport ./tests/performance/... 2>&1 | Tee-Object -FilePath "$REPORT_DIR\report_generation_$TIMESTAMP.txt"
    
    Write-ColorOutput Green "✓ 性能测试报告已生成"
    Write-ColorOutput Green "  报告目录: $REPORT_DIR"
    Write-Output ""
    
    # 显示报告摘要
    $reportFile = Join-Path $REPORT_DIR "performance_report.json"
    if (Test-Path $reportFile) {
        Write-ColorOutput Yellow "性能测试摘要:"
        Get-Content $reportFile | ConvertFrom-Json | ConvertTo-Json -Depth 10
    }
}

# 设置性能基线
function Set-Baseline {
    Write-ColorOutput Cyan "========================================"
    Write-ColorOutput Cyan "  设置性能基线"
    Write-ColorOutput Cyan "========================================"
    Write-Output ""
    
    # 复制当前测试结果作为基线
    Copy-Item -Path "$REPORT_DIR\*" -Destination $BASELINE_DIR -Recurse -Force
    
    Write-ColorOutput Green "✓ 性能基线已设置"
    Write-ColorOutput Green "  基线目录: $BASELINE_DIR"
    Write-Output ""
}

# 清理旧报告
function Cleanup-OldReports {
    Write-ColorOutput Yellow "清理旧报告 (保留最近30天)..."
    
    $cutoffDate = (Get-Date).AddDays(-30)
    Get-ChildItem -Path $REPORT_DIR -File | Where-Object { $_.LastWriteTime -lt $cutoffDate } | Remove-Item -Force
    
    Write-ColorOutput Green "✓ 清理完成"
    Write-Output ""
}

# 主函数
function Main {
    # 显示帮助
    if ($Help) {
        Show-Help
        exit 0
    }
    
    Write-ColorOutput Cyan "========================================"
    Write-ColorOutput Cyan "  新能源监控系统 - 性能测试套件"
    Write-ColorOutput Cyan "========================================"
    Write-Output ""
    
    # 检查依赖
    Check-Dependencies
    
    # 清理旧报告
    Cleanup-OldReports
    
    # 根据测试类型运行测试
    switch ($TestType) {
        "all" {
            Run-BackendTests
            Run-FrontendTests
            Generate-Report
        }
        "backend" {
            Run-BackendTests
            Generate-Report
        }
        "frontend" {
            Run-FrontendTests
        }
        { $_ -in @("api", "db", "memory", "collector", "query", "stress") } {
            Run-BackendTests
        }
        "report" {
            Generate-Report
        }
        "baseline" {
            Set-Baseline
        }
        "compare" {
            Write-ColorOutput Yellow "对比功能开发中..."
        }
    }
    
    Write-ColorOutput Green "========================================"
    Write-ColorOutput Green "  性能测试完成!"
    Write-ColorOutput Green "========================================"
    Write-Output ""
    Write-Output "报告目录: $REPORT_DIR"
    Write-Output ""
    Write-Output "查看性能分析:"
    Write-Output "  CPU: go tool pprof -http=:8080 $REPORT_DIR\cpu.prof"
    Write-Output "  内存: go tool pprof -http=:8080 $REPORT_DIR\mem.prof"
    Write-Output ""
}

# 执行主函数
Main
