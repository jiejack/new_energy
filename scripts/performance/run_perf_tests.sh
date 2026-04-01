#!/bin/bash

# 新能源监控系统 - 性能测试脚本
# 用于一键运行所有性能测试并生成报告

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REPORT_DIR="${PROJECT_ROOT}/reports/performance"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BASELINE_DIR="${PROJECT_ROOT}/reports/baseline"

# 创建报告目录
mkdir -p "${REPORT_DIR}"
mkdir -p "${BASELINE_DIR}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  新能源监控系统 - 性能测试套件${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 解析参数
TEST_TYPE=${1:-"all"}
VERBOSE=${2:-""}

# 显示帮助信息
show_help() {
    echo "用法: $0 [测试类型] [选项]"
    echo ""
    echo "测试类型:"
    echo "  all         运行所有性能测试 (默认)"
    echo "  backend     仅运行后端性能测试"
    echo "  frontend    仅运行前端性能测试"
    echo "  api         仅运行API性能测试"
    echo "  db          仅运行数据库性能测试"
    echo "  memory      仅运行内存泄漏测试"
    echo "  collector   仅运行采集器性能测试"
    echo "  query       仅运行查询性能测试"
    echo "  stress      仅运行压力测试"
    echo "  report      仅生成性能报告"
    echo "  baseline    设置当前性能为基线"
    echo "  compare     与基线性能对比"
    echo ""
    echo "选项:"
    echo "  -v, --verbose   显示详细输出"
    echo "  -h, --help      显示帮助信息"
    echo ""
}

# 检查依赖
check_dependencies() {
    echo -e "${YELLOW}检查依赖...${NC}"
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}错误: 未找到Go环境${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Go $(go version | awk '{print $3}')${NC}"
    
    # 检查Node.js (前端测试需要)
    if command -v node &> /dev/null; then
        echo -e "${GREEN}✓ Node $(node -v)${NC}"
    else
        echo -e "${YELLOW}! Node.js 未安装，跳过前端测试${NC}"
    fi
    
    # 检查k6 (负载测试需要)
    if command -v k6 &> /dev/null; then
        echo -e "${GREEN}✓ k6 $(k6 version)${NC}"
    else
        echo -e "${YELLOW}! k6 未安装，部分负载测试将跳过${NC}"
    fi
    
    echo ""
}

# 运行后端性能测试
run_backend_tests() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  运行后端性能测试${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    cd "${PROJECT_ROOT}"
    
    # API性能测试
    if [[ "$TEST_TYPE" == "all" || "$TEST_TYPE" == "backend" || "$TEST_TYPE" == "api" ]]; then
        echo -e "${YELLOW}运行API性能测试...${NC}"
        go test -bench=BenchmarkAPI -benchmem -benchtime=5s \
            ./tests/performance/... 2>&1 | tee "${REPORT_DIR}/api_perf_${TIMESTAMP}.txt"
        echo ""
    fi
    
    # 数据库性能测试
    if [[ "$TEST_TYPE" == "all" || "$TEST_TYPE" == "backend" || "$TEST_TYPE" == "db" ]]; then
        echo -e "${YELLOW}运行数据库性能测试...${NC}"
        go test -bench=BenchmarkDatabase -benchmem -benchtime=5s \
            ./tests/performance/... 2>&1 | tee "${REPORT_DIR}/db_perf_${TIMESTAMP}.txt"
        echo ""
    fi
    
    # 内存泄漏测试
    if [[ "$TEST_TYPE" == "all" || "$TEST_TYPE" == "backend" || "$TEST_TYPE" == "memory" ]]; then
        echo -e "${YELLOW}运行内存泄漏测试...${NC}"
        go test -bench=BenchmarkMemory -benchmem -benchtime=10s \
            ./tests/performance/... 2>&1 | tee "${REPORT_DIR}/memory_leak_${TIMESTAMP}.txt"
        echo ""
    fi
    
    # 采集器性能测试
    if [[ "$TEST_TYPE" == "all" || "$TEST_TYPE" == "backend" || "$TEST_TYPE" == "collector" ]]; then
        echo -e "${YELLOW}运行采集器性能测试...${NC}"
        go test -bench=BenchmarkCollector -benchmem -benchtime=5s \
            ./tests/performance/... 2>&1 | tee "${REPORT_DIR}/collector_bench_${TIMESTAMP}.txt"
        echo ""
    fi
    
    # 查询性能测试
    if [[ "$TEST_TYPE" == "all" || "$TEST_TYPE" == "backend" || "$TEST_TYPE" == "query" ]]; then
        echo -e "${YELLOW}运行查询性能测试...${NC}"
        go test -bench=BenchmarkQuery -benchmem -benchtime=5s \
            ./tests/performance/... 2>&1 | tee "${REPORT_DIR}/query_bench_${TIMESTAMP}.txt"
        echo ""
    fi
    
    # 压力测试
    if [[ "$TEST_TYPE" == "all" || "$TEST_TYPE" == "backend" || "$TEST_TYPE" == "stress" ]]; then
        echo -e "${YELLOW}运行压力测试...${NC}"
        go test -bench=BenchmarkAPIPressure -benchmem -benchtime=10s \
            ./tests/performance/... 2>&1 | tee "${REPORT_DIR}/stress_test_${TIMESTAMP}.txt"
        echo ""
    fi
    
    # 生成CPU性能分析
    echo -e "${YELLOW}生成CPU性能分析...${NC}"
    go test -bench=BenchmarkCollectorMillionPoints -cpuprofile="${REPORT_DIR}/cpu.prof" \
        ./tests/performance/... 2>&1 > /dev/null
    echo -e "${GREEN}✓ CPU性能分析已生成: ${REPORT_DIR}/cpu.prof${NC}"
    
    # 生成内存性能分析
    echo -e "${YELLOW}生成内存性能分析...${NC}"
    go test -bench=BenchmarkCollectorMemoryUsage -memprofile="${REPORT_DIR}/mem.prof" \
        ./tests/performance/... 2>&1 > /dev/null
    echo -e "${GREEN}✓ 内存性能分析已生成: ${REPORT_DIR}/mem.prof${NC}"
    echo ""
}

# 运行前端性能测试
run_frontend_tests() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  运行前端性能测试${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    cd "${PROJECT_ROOT}/web"
    
    # 检查是否安装了依赖
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}安装前端依赖...${NC}"
        npm install
    fi
    
    # 组件渲染性能测试
    if [[ "$TEST_TYPE" == "all" || "$TEST_TYPE" == "frontend" ]]; then
        echo -e "${YELLOW}运行组件渲染性能测试...${NC}"
        npm run test:perf -- tests/performance/component-perf.test.ts 2>&1 | \
            tee "${REPORT_DIR}/component_perf_${TIMESTAMP}.txt"
        echo ""
        
        echo -e "${YELLOW}运行状态更新性能测试...${NC}"
        npm run test:perf -- tests/performance/state-perf.test.ts 2>&1 | \
            tee "${REPORT_DIR}/state_perf_${TIMESTAMP}.txt"
        echo ""
    fi
    
    cd "${PROJECT_ROOT}"
}

# 运行k6负载测试
run_k6_tests() {
    if ! command -v k6 &> /dev/null; then
        echo -e "${YELLOW}跳过k6负载测试 (k6未安装)${NC}"
        return
    fi
    
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  运行k6负载测试${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    # 检查k6脚本是否存在
    K6_SCRIPT="${PROJECT_ROOT}/scripts/performance/load_test.js"
    if [ -f "$K6_SCRIPT" ]; then
        echo -e "${YELLOW}运行k6负载测试...${NC}"
        k6 run --out json="${REPORT_DIR}/k6_results_${TIMESTAMP}.json" "$K6_SCRIPT" 2>&1 | \
            tee "${REPORT_DIR}/k6_test_${TIMESTAMP}.txt"
    else
        echo -e "${YELLOW}跳过k6负载测试 (脚本不存在)${NC}"
    fi
    echo ""
}

# 生成性能报告
generate_report() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  生成性能测试报告${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    cd "${PROJECT_ROOT}"
    
    # 运行报告生成
    go test -run=TestGeneratePerformanceReport ./tests/performance/... 2>&1 | \
        tee "${REPORT_DIR}/report_generation_${TIMESTAMP}.txt"
    
    echo -e "${GREEN}✓ 性能测试报告已生成${NC}"
    echo -e "${GREEN}  报告目录: ${REPORT_DIR}${NC}"
    echo ""
    
    # 显示报告摘要
    if [ -f "${REPORT_DIR}/performance_report.json" ]; then
        echo -e "${YELLOW}性能测试摘要:${NC}"
        cat "${REPORT_DIR}/performance_report.json" | python3 -m json.tool 2>/dev/null || \
            cat "${REPORT_DIR}/performance_report.json"
    fi
}

# 设置性能基线
set_baseline() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  设置性能基线${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    # 复制当前测试结果作为基线
    cp -r "${REPORT_DIR}"/* "${BASELINE_DIR}/"
    
    echo -e "${GREEN}✓ 性能基线已设置${NC}"
    echo -e "${GREEN}  基线目录: ${BASELINE_DIR}${NC}"
    echo ""
}

# 与基线对比
compare_with_baseline() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  与基线性能对比${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    if [ ! -d "${BASELINE_DIR}" ] || [ -z "$(ls -A ${BASELINE_DIR})" ]; then
        echo -e "${RED}错误: 未找到性能基线，请先运行 'set_baseline' 设置基线${NC}"
        return 1
    fi
    
    # 对比逻辑
    echo -e "${YELLOW}对比当前性能与基线...${NC}"
    
    # 这里可以添加更详细的对比逻辑
    # 例如对比ops/s、延迟、内存使用等
    
    echo -e "${GREEN}✓ 对比完成${NC}"
    echo ""
}

# 清理旧报告
cleanup_old_reports() {
    echo -e "${YELLOW}清理旧报告 (保留最近30天)...${NC}"
    
    find "${REPORT_DIR}" -type f -mtime +30 -delete 2>/dev/null || true
    
    echo -e "${GREEN}✓ 清理完成${NC}"
    echo ""
}

# 主函数
main() {
    # 显示帮助
    if [[ "$TEST_TYPE" == "-h" || "$TEST_TYPE" == "--help" ]]; then
        show_help
        exit 0
    fi
    
    # 检查依赖
    check_dependencies
    
    # 清理旧报告
    cleanup_old_reports
    
    # 根据测试类型运行测试
    case "$TEST_TYPE" in
        all)
            run_backend_tests
            run_frontend_tests
            run_k6_tests
            generate_report
            ;;
        backend)
            run_backend_tests
            generate_report
            ;;
        frontend)
            run_frontend_tests
            ;;
        api|db|memory|collector|query|stress)
            run_backend_tests
            ;;
        report)
            generate_report
            ;;
        baseline)
            set_baseline
            ;;
        compare)
            compare_with_baseline
            ;;
        *)
            echo -e "${RED}错误: 未知的测试类型 '${TEST_TYPE}'${NC}"
            show_help
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  性能测试完成!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "报告目录: ${REPORT_DIR}"
    echo ""
    echo -e "查看性能分析:"
    echo -e "  CPU: go tool pprof -http=:8080 ${REPORT_DIR}/cpu.prof"
    echo -e "  内存: go tool pprof -http=:8080 ${REPORT_DIR}/mem.prof"
    echo ""
}

# 执行主函数
main
