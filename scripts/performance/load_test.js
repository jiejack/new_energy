// k6 负载测试脚本
// 用于测试新能源监控系统的API性能

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// 自定义指标
const errorRate = new Rate('errors');
const apiLatency = new Trend('api_latency');
const requestCount = new Counter('requests');

// 测试配置
export const options = {
  // 阶段性负载测试
  stages: [
    { duration: '30s', target: 20 },   // 预热阶段：30秒内增加到20个虚拟用户
    { duration: '1m', target: 50 },    // 增长阶段：1分钟内增加到50个虚拟用户
    { duration: '2m', target: 100 },   // 峰值阶段：2分钟内增加到100个虚拟用户
    { duration: '1m', target: 100 },   // 稳定阶段：保持100个虚拟用户1分钟
    { duration: '30s', target: 0 },    // 降温阶段：30秒内降到0
  ],
  
  // 性能阈值
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95%请求<500ms, 99%请求<1000ms
    errors: ['rate<0.05'],                          // 错误率<5%
    http_req_failed: ['rate<0.01'],                 // 失败率<1%
  },
  
  // 并发设置
  noConnectionReuse: false,
  userAgent: 'k6-load-test/1.0',
};

// 基础URL配置
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// 测试数据
const testData = {
  stations: [],
  devices: [],
  points: [],
};

// 初始化测试数据
export function setup() {
  console.log('初始化测试数据...');
  
  // 创建测试站点
  const stations = [];
  for (let i = 0; i < 10; i++) {
    stations.push({
      name: `test-station-${i}`,
      type: ['solar', 'wind', 'hydro'][i % 3],
      capacity: Math.random() * 1000,
      region_id: 'region-001',
    });
  }
  
  // 创建测试设备
  const devices = [];
  for (let i = 0; i < 50; i++) {
    devices.push({
      name: `test-device-${i}`,
      type: ['inverter', 'meter', 'sensor'][i % 3],
      station_id: `station-${i % 10}`,
    });
  }
  
  // 创建测试测点
  const points = [];
  for (let i = 0; i < 100; i++) {
    points.push({
      code: `POINT-${String(i).padStart(4, '0')}`,
      name: `Test Point ${i}`,
      type: ['yaoc', 'yaoxin', 'yaokong'][i % 3],
      device_id: `device-${i % 50}`,
    });
  }
  
  return { stations, devices, points };
}

// 默认函数 - 每个虚拟用户执行
export default function (data) {
  // 随机选择测试场景
  const scenario = Math.floor(Math.random() * 10);
  
  switch (scenario) {
    case 0:
    case 1:
    case 2:
      // 30% - 查询站点列表
      testGetStations();
      break;
    case 3:
    case 4:
      // 20% - 查询设备列表
      testGetDevices();
      break;
    case 5:
    case 6:
      // 20% - 查询测点数据
      testGetPoints();
      break;
    case 7:
      // 10% - 查询实时数据
      testGetRealtimeData();
      break;
    case 8:
      // 10% - 查询历史数据
      testGetHistoryData();
      break;
    case 9:
      // 10% - 查询告警列表
      testGetAlarms();
      break;
  }
  
  // 随机等待时间
  sleep(Math.random() * 2 + 0.5);
}

// 测试场景函数

function testGetStations() {
  const url = `${BASE_URL}/api/v1/stations`;
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer test-token',
    },
  };
  
  const res = http.get(url, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  
  errorRate.add(res.status !== 200);
  apiLatency.add(res.timings.duration);
  requestCount.add(1);
}

function testGetDevices() {
  const stationId = `station-${Math.floor(Math.random() * 10)}`;
  const url = `${BASE_URL}/api/v1/devices?station_id=${stationId}`;
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer test-token',
    },
  };
  
  const res = http.get(url, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  errorRate.add(res.status !== 200);
  apiLatency.add(res.timings.duration);
  requestCount.add(1);
}

function testGetPoints() {
  const deviceId = `device-${Math.floor(Math.random() * 50)}`;
  const url = `${BASE_URL}/api/v1/points?device_id=${deviceId}`;
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer test-token',
    },
  };
  
  const res = http.get(url, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  errorRate.add(res.status !== 200);
  apiLatency.add(res.timings.duration);
  requestCount.add(1);
}

function testGetRealtimeData() {
  const pointIds = [];
  for (let i = 0; i < 10; i++) {
    pointIds.push(`point-${Math.floor(Math.random() * 100)}`);
  }
  
  const url = `${BASE_URL}/api/v1/data/realtime?point_ids=${pointIds.join(',')}`;
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer test-token',
    },
  };
  
  const res = http.get(url, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has realtime data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  
  errorRate.add(res.status !== 200);
  apiLatency.add(res.timings.duration);
  requestCount.add(1);
}

function testGetHistoryData() {
  const pointId = `point-${Math.floor(Math.random() * 100)}`;
  const endTime = Date.now();
  const startTime = endTime - 3600000; // 1小时前
  
  const url = `${BASE_URL}/api/v1/data/history?point_id=${pointId}&start_time=${startTime}&end_time=${endTime}`;
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer test-token',
    },
  };
  
  const res = http.get(url, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
  });
  
  errorRate.add(res.status !== 200);
  apiLatency.add(res.timings.duration);
  requestCount.add(1);
}

function testGetAlarms() {
  const url = `${BASE_URL}/api/v1/alarms?page=1&page_size=20`;
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer test-token',
    },
  };
  
  const res = http.get(url, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has alarms': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  
  errorRate.add(res.status !== 200);
  apiLatency.add(res.timings.duration);
  requestCount.add(1);
}

// 清理函数
export function teardown(data) {
  console.log('测试完成，清理资源...');
}

// 处理汇总数据
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'reports/performance/k6_summary.json': JSON.stringify(data, null, 2),
  };
}

// 文本摘要函数
function textSummary(data, options) {
  const indent = options.indent || '';
  const colors = options.enableColors || false;
  
  let summary = '\n';
  summary += indent + '========================================\n';
  summary += indent + '  k6 负载测试报告\n';
  summary += indent + '========================================\n\n';
  
  // HTTP请求统计
  if (data.metrics.http_req_duration) {
    summary += indent + 'HTTP请求延迟:\n';
    summary += indent + `  平均: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
    summary += indent + `  最小: ${data.metrics.http_req_duration.values.min.toFixed(2)}ms\n`;
    summary += indent + `  最大: ${data.metrics.http_req_duration.values.max.toFixed(2)}ms\n`;
    summary += indent + `  P95: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
    summary += indent + `  P99: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n\n`;
  }
  
  // 请求统计
  if (data.metrics.http_reqs) {
    summary += indent + '请求统计:\n';
    summary += indent + `  总请求数: ${data.metrics.http_reqs.values.count}\n`;
    summary += indent + `  请求速率: ${data.metrics.http_reqs.values.rate.toFixed(2)}/s\n\n`;
  }
  
  // 错误率
  if (data.metrics.errors) {
    summary += indent + '错误率:\n';
    summary += indent + `  错误率: ${(data.metrics.errors.values.rate * 100).toFixed(2)}%\n\n`;
  }
  
  // 数据传输
  if (data.metrics.data_received && data.metrics.data_sent) {
    summary += indent + '数据传输:\n';
    summary += indent + `  接收: ${(data.metrics.data_received.values.count / 1024 / 1024).toFixed(2)}MB\n`;
    summary += indent + `  发送: ${(data.metrics.data_sent.values.count / 1024 / 1024).toFixed(2)}MB\n\n`;
  }
  
  summary += indent + '========================================\n';
  
  return summary;
}
