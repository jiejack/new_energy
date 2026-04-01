#!/bin/bash

# 新能源监控系统性能测试脚本
# 用于运行所有性能测试并生成报告

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEST_DIR="$PROJECT_ROOT/tests/performance"
REPORT_DIR="$PROJECT_ROOT/reports/performance"

# 创建报告目录
mkdir -p "$REPORT_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}新能源监控系统 - 性能测试${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 函数：运行性能测试
run_benchmark() {
    local name=$1
    local pattern=$2
    local output_file="$REPORT_DIR/${name}.txt"

    echo -e "${YELLOW}运行测试: $name${NC}"
    echo "开始时间: $(date '+%Y-%m-%d %H:%M:%S')"

    # 运行benchmark
    cd "$PROJECT_ROOT"
    go test -bench="$pattern" \
        -benchmem \
        -benchtime=3s \
        -run=^$ \
        -timeout=30m \
        ./tests/performance/... | tee "$output_file"

    echo -e "${GREEN}测试完成: $name${NC}"
    echo "报告保存至: $output_file"
    echo ""
}

# 函数：生成CPU性能分析
generate_cpu_profile() {
    local name=$1
    local pattern=$2
    local profile_file="$REPORT_DIR/${name}_cpu.prof"

    echo -e "${YELLOW}生成CPU性能分析: $name${NC}"

    cd "$PROJECT_ROOT"
    go test -bench="$pattern" \
        -cpuprofile="$profile_file" \
        -run=^$ \
        -timeout=10m \
        ./tests/performance/...

    echo -e "${GREEN}CPU性能分析完成: $profile_file${NC}"
    echo ""
}

# 函数：生成内存性能分析
generate_mem_profile() {
    local name=$1
    local pattern=$2
    local profile_file="$REPORT_DIR/${name}_mem.prof"

    echo -e "${YELLOW}生成内存性能分析: $name${NC}"

    cd "$PROJECT_ROOT"
    go test -bench="$pattern" \
        -memprofile="$profile_file" \
        -run=^$ \
        -timeout=10m \
        ./tests/performance/...

    echo -e "${GREEN}内存性能分析完成: $profile_file${NC}"
    echo ""
}

# 函数：生成性能报告
generate_report() {
    echo -e "${YELLOW}生成性能测试报告...${NC}"

    cd "$PROJECT_ROOT"
    go test -v -run=TestReportGeneration ./tests/performance/...

    echo -e "${GREEN}性能测试报告生成完成${NC}"
    echo ""
}

# 主菜单
show_menu() {
    echo -e "${BLUE}请选择要运行的测试:${NC}"
    echo "1) 运行所有性能测试"
    echo "2) 采集性能测试"
    echo "3) 查询性能测试"
    echo "4) 并发压力测试"
    echo "5) 生成CPU性能分析"
    echo "6) 生成内存性能分析"
    echo "7) 生成性能测试报告"
    echo "8) 查看性能分析（需要pprof）"
    echo "0) 退出"
    echo ""
    echo -n "请输入选项: "
}

# 查看性能分析
view_profile() {
    local profile_type=$1
    local profile_file="$REPORT_DIR/${profile_type}.prof"

    if [ ! -f "$profile_file" ]; then
        echo -e "${RED}性能分析文件不存在: $profile_file${NC}"
        echo "请先运行性能分析生成"
        return
    fi

    echo -e "${YELLOW}启动pprof web界面...${NC}"
    go tool pprof -http=:8080 "$profile_file"
}

# 主循环
while true; do
    show_menu
    read choice

    case $choice in
        1)
            echo -e "${GREEN}运行所有性能测试...${NC}"
            run_benchmark "collector_bench" "BenchmarkCollector"
            run_benchmark "query_bench" "BenchmarkQuery"
            run_benchmark "stress_test" "BenchmarkAPI|BenchmarkWebSocket|BenchmarkDatabase|BenchmarkMessage"
            generate_report
            ;;
        2)
            run_benchmark "collector_bench" "BenchmarkCollector"
            ;;
        3)
            run_benchmark "query_bench" "BenchmarkQuery"
            ;;
        4)
            run_benchmark "stress_test" "BenchmarkAPI|BenchmarkWebSocket|BenchmarkDatabase|BenchmarkMessage"
            ;;
        5)
            echo -e "${YELLOW}选择CPU性能分析类型:${NC}"
            echo "1) 采集器CPU分析"
            echo "2) 查询CPU分析"
            echo "3) 压力测试CPU分析"
            read cpu_choice
            case $cpu_choice in
                1) generate_cpu_profile "collector" "BenchmarkCollectorMillionPoints" ;;
                2) generate_cpu_profile "query" "BenchmarkQueryMillionRecords" ;;
                3) generate_cpu_profile "stress" "BenchmarkAPIPressure" ;;
                *) echo -e "${RED}无效选项${NC}" ;;
            esac
            ;;
        6)
            echo -e "${YELLOW}选择内存性能分析类型:${NC}"
            echo "1) 采集器内存分析"
            echo "2) 查询内存分析"
            echo "3) 压力测试内存分析"
            read mem_choice
            case $mem_choice in
                1) generate_mem_profile "collector" "BenchmarkCollectorMemoryUsage" ;;
                2) generate_mem_profile "query" "BenchmarkQueryMemoryAllocation" ;;
                3) generate_mem_profile "stress" "BenchmarkMemoryPressure" ;;
                *) echo -e "${RED}无效选项${NC}" ;;
            esac
            ;;
        7)
            generate_report
            ;;
        8)
            echo -e "${YELLOW}选择要查看的性能分析:${NC}"
            echo "1) CPU性能分析"
            echo "2) 内存性能分析"
            read view_choice
            case $view_choice in
                1)
                    echo -e "${YELLOW}选择CPU分析文件:${NC}"
                    ls -1 "$REPORT_DIR"/*_cpu.prof 2>/dev/null || echo "没有CPU分析文件"
                    read cpu_file
                    view_profile "$cpu_file"
                    ;;
                2)
                    echo -e "${YELLOW}选择内存分析文件:${NC}"
                    ls -1 "$REPORT_DIR"/*_mem.prof 2>/dev/null || echo "没有内存分析文件"
                    read mem_file
                    view_profile "$mem_file"
                    ;;
                *) echo -e "${RED}无效选项${NC}" ;;
            esac
            ;;
        0)
            echo -e "${GREEN}退出性能测试${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}无效选项，请重新选择${NC}"
            ;;
    esac

    echo ""
    echo -e "${BLUE}按回车键继续...${NC}"
    read
done
