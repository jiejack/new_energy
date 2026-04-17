# 新能源监控系统 - Wiki 首页

## 项目简介

新能源监控系统（New Energy Monitoring System，简称NEM）是一个基于云原生架构的分布式能源监控平台，专为光伏电站、风电场等新能源设施设计。系统采用微服务架构，支持大规模设备接入、实时数据采集、智能告警分析和可视化展示。

## 核心特性

| 特性 | 描述 |
|------|------|
| 🔌 **多协议支持** | 支持Modbus、IEC104、IEC61850等工业协议 |
| 📊 **实时监控** | 毫秒级数据采集，实时展示设备状态 |
| 🚨 **智能告警** | 基于规则的告警引擎，支持多渠道通知 |
| 📈 **数据分析** | 历史数据存储与查询，支持多种报表导出 |
| 🔐 **权限管理** | 完善的RBAC权限体系，支持细粒度控制 |
| 🐳 **云原生** | 容器化部署，支持Kubernetes编排 |

## 技术栈

### 后端
- **语言**: Go 1.24
- **框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL 15
- **缓存**: Redis 7
- **消息队列**: Kafka

### 前端
- **框架**: Vue 3 + TypeScript
- **UI组件**: Element Plus
- **状态管理**: Pinia
- **图表**: ECharts
- **构建工具**: Vite

### 基础设施
- **容器化**: Docker + Kubernetes
- **监控**: Prometheus + Grafana
- **追踪**: Jaeger
- **CI/CD**: GitHub Actions

## Wiki 目录

### 新手入门
- [安装指南](./Installation-Guide) - 环境准备与安装部署
- [快速开始](./Quick-Start) - 5分钟快速上手
- [项目结构](./Project-Structure) - 代码目录说明

### 使用文档
- [用户手册](./User-Manual) - 功能使用说明
- [功能说明](./Feature-Guide) - 详细功能模块介绍
- [API文档](./API-Documentation) - 接口调用指南
- [配置说明](./Configuration) - 系统配置详解

### 开发文档
- [开发指南](./Developer-Guide) - 开发环境搭建与规范
- [测试指南](./Testing-Guide) - 测试策略与用例编写
- [部署指南](./Deployment-Guide) - 生产环境部署

### 运维文档
- [运维手册](./Operations-Manual) - 日常运维操作
- [故障排查](./Troubleshooting) - 常见问题解决
- [性能调优](./Performance-Tuning) - 性能优化指南

### 参考文档
- [架构设计](./Architecture) - 系统架构说明
- [数据库设计](./Database-Design) - 数据模型文档
- [FAQ](./FAQ) - 常见问题解答

## 快速链接

| 链接 | 描述 |
|------|------|
| [GitHub仓库](https://github.com/jiejack/new_energy) | 源代码仓库 |
| [问题反馈](https://github.com/jiejack/new_energy/issues) | Bug报告与功能建议 |
| [更新日志](./Changelog) | 版本更新记录 |

## 项目状态

![Build Status](https://img.shields.io/github/actions/workflow/status/jiejack/new_energy/ci.yml?branch=main)
![Coverage](https://img.shields.io/codecov/c/github/jiejack/new_energy)
![License](https://img.shields.io/github/license/jiejack/new_energy)
![Go Version](https://img.shields.io/github/go-mod/go-version/jiejack/new_energy)

## 贡献指南

我们欢迎所有形式的贡献！请阅读 [贡献指南](./Contributing) 了解如何参与项目开发。

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](https://github.com/jiejack/new_energy/blob/main/LICENSE) 文件。

---

**最后更新**: 2026-04-07
