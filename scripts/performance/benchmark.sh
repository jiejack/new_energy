#!/bin/bash
# 性能基准测试脚本
# 用于验证系统性能指标

set -e

echo "=========================================="
echo "  新能源监控系统 - 性能基准测试"
echo "=========================================="

API_URL="${API_URL:-http://localhost:8080}"
RESULTS_DIR="scripts/performance/results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_FILE="$RESULTS_DIR/benchmark_$TIMESTAMP.json"

mkdir -p $RESULTS_DIR

echo ""
echo "测试配置:"
echo "  API URL: $API_URL"
echo "  结果文件: $RESULT_FILE"
echo ""

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "错误: 未找到命令 '$1'"
        exit 1
    fi
}

check_command curl
check_command jq

API_RESPONSE_TIME=""
API_P95=""
API_P99=""

test_api_health() {
    echo ">>> 测试 API 健康检查..."
    START=$(date +%s%N)
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/health")
    END=$(date +%s%N)
    
    DURATION_MS=$(( (END - START) / 1000000 ))
    
    if [ "$RESPONSE" = "200" ]; then
        echo "  ✓ 健康检查通过 (${DURATION_MS}ms)"
        API_RESPONSE_TIME=$DURATION_MS
        return 0
    else
        echo "  ✗ 健康检查失败 (HTTP $RESPONSE)"
        return 1
    fi
}

test_api_latency() {
    echo ">>> 测试 API 延迟..."
    
    TOTAL=0
    COUNT=20
    TIMES=()
    
    for i in $(seq 1 $COUNT); do
        START=$(date +%s%N)
        curl -s -o /dev/null "$API_URL/health"
        END=$(date +%s%N)
        DURATION_MS=$(( (END - START) / 1000000 ))
        TIMES+=($DURATION_MS)
        TOTAL=$((TOTAL + DURATION_MS))
    done
    
    AVG=$((TOTAL / COUNT))
    
    SORTED=($(printf '%s\n' "${TIMES[@]}" | sort -n))
    P95_IDX=$((COUNT * 95 / 100 - 1))
    P99_IDX=$((COUNT * 99 / 100 - 1))
    
    P95=${SORTED[$P95_IDX]}
    P99=${SORTED[$P99_IDX]}
    
    echo "  平均延迟: ${AVG}ms"
    echo "  P95 延迟: ${P95}ms"
    echo "  P99 延迟: ${P99}ms"
    
    API_P95=$P95
    API_P99=$P99
    
    if [ $P95 -lt 200 ]; then
        echo "  ✓ P95 延迟符合要求 (<200ms)"
    else
        echo "  ✗ P95 延迟超标 (>200ms)"
    fi
}

test_concurrent_requests() {
    echo ">>> 测试并发请求..."
    
    CONCURRENT=100
    SUCCESS=0
    FAIL=0
    
    for i in $(seq 1 $CONCURRENT); do
        if curl -s -o /dev/null -w "%{http_code}" "$API_URL/health" | grep -q "200"; then
            SUCCESS=$((SUCCESS + 1))
        else
            FAIL=$((FAIL + 1))
        fi &
    done
    wait
    
    echo "  成功: $SUCCESS"
    echo "  失败: $FAIL"
    
    if [ $FAIL -eq 0 ]; then
        echo "  ✓ 并发测试通过"
    else
        echo "  ✗ 并发测试有失败"
    fi
}

test_memory_usage() {
    echo ">>> 测试内存使用..."
    
    if command -v docker &> /dev/null; then
        MEMORY=$(docker stats --no-stream --format "{{.MemUsage}}" 2>/dev/null | head -1)
        if [ -n "$MEMORY" ]; then
            echo "  容器内存: $MEMORY"
        fi
    fi
    
    if command -v ps &> /dev/null; then
        PID=$(pgrep -f "api-server" | head -1)
        if [ -n "$PID" ]; then
            MEM_MB=$(ps -o rss= -p $PID | awk '{print int($1/1024)}')
            echo "  进程内存: ${MEM_MB}MB"
            
            if [ $MEM_MB -lt 512 ]; then
                echo "  ✓ 内存使用正常 (<512MB)"
            else
                echo "  ⚠ 内存使用较高 (>512MB)"
            fi
        fi
    fi
}

generate_report() {
    echo ""
    echo "=========================================="
    echo "  性能测试报告"
    echo "=========================================="
    
    cat > $RESULT_FILE << EOF
{
  "timestamp": "$(date -Iseconds)",
  "api_url": "$API_URL",
  "metrics": {
    "response_time_ms": $API_RESPONSE_TIME,
    "p95_latency_ms": $API_P95,
    "p99_latency_ms": $API_P99
  },
  "thresholds": {
    "p95_latency_ms": 200,
    "p99_latency_ms": 500
  },
  "passed": $([ $API_P95 -lt 200 ] && echo "true" || echo "false")
}
EOF
    
    echo "报告已保存: $RESULT_FILE"
    echo ""
    
    if [ $API_P95 -lt 200 ]; then
        echo "✓ 所有性能测试通过"
        return 0
    else
        echo "✗ 部分性能测试未通过"
        return 1
    fi
}

main() {
    test_api_health
    test_api_latency
    test_concurrent_requests
    test_memory_usage
    generate_report
}

main
