# 🚀 Go Online Editor


<div align="center">
  <a href="https://github.com/HHU-47133/go-online-editor">
    <img src="./static/assets/Go-Online-Editor.png" alt="go-online-editor Logo" width="200"/>
  </a>

一个轻量级在线 Golang 代码编辑与运行平台

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/HHU-47133/go-online-editor)](https://goreportcard.com/report/github.com/HHU-47133/go-online-editor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

## 💡 为什么需要它？
在大多数正式的 ACM/ICPC 或各类编程竞赛中，Go 往往不在官方支持之列。但对于热爱 Go 的开发者来说，其简洁的语法与卓越的性能，本该是算法竞技中的神兵利器。

为此，我们专为 Gopher 选手量身打造了这款标准 ACM 模式编辑平台。它将 Vim 的指尖律动与极致的赛场反馈精准耦合，助你在算法的荒原里，以 Go 之名，行猎杀之实。

## ✨ 功能特性

- ⚡ 极速体验：后端完全基于 Go 标准库构建，无任何外部框架依赖，编译与执行响应控制在毫秒级。

- 🏆 专为竞赛设计
  - ACM 模式模拟：完美模拟标准输入（`Stdin`）与标准输出（`Stdout/Stderr`）交互，支持大数据量输入测试。

  - 算法模板集成：通过 `git submodule` 深度整合 `goacm` 模板库，支持常见数据结构与算法一键预览、一键插入。

- ⌨️ 编辑器增强：

  - Vim Mode：集成 `Monaco Editor` 并原生支持 `Vim` 模式，满足算法竞赛玩家对输入速度的极致追求。

  - 智能增强：内置代码高亮、自动对齐、`Ctrl + S` 实时格式化及 `Ctrl + Enter` 快捷运行。

- 🛡️ 安全与自愈：
  - 异常清理：当用户刷新页面或断开连接时，后端通过 Context 级联取消 自动清理整个进程组，严防死循环耗尽服务器资源。

  - 手动终止：提供停止按钮，可强行中断无限循环等异常程序。

  - 资源限制：支持 `Docker` 容器化部署，默认限制 `512M` 内存。

  - 路径沙箱：模板读取经过路径清洗（`filepath.Clean`），彻底杜绝目录穿越攻击。


## 📂 项目结构

```plaintext
.
├── main.go                # 核心后端逻辑：编译、运行及进程管理  
├── goacm/                 # (Submodule) 丰富的 Go 算法模板库  
├── static/                
│   ├── assets             # 资源文件夹  
│   ├── index.html         # 前端单页 UI  
│   └── default.tpl        # 默认快速启动代码模板   
├── Dockerfile             # 多阶段构建容器镜像  
└── docker-compose.yml     # 一键服务编排  
```

## 🚀 快速开始

**方式一：使用 Docker (推荐)**

```bash
docker compose up --build -d
```

访问：`http://localhost:2009`

**方式二：本地运行**

- 环境要求：`Go 1.25+`

```bash
git submodule update --init --remote --recursive
go mod tidy
go run main.go
```

## 🤝 贡献指南
我们非常欢迎社区的贡献：

- 新增模板：请在 `goacm` 子模块中按类别提交 PR。

- 改进前端：保持纯原生 `JavaScript/HTML` 风格，不引入重型框架。

- 代码风格：遵循标准 `gofmt`。

## 📜 开源许可

本项目采用 [MIT License](LICENSE) 开源。