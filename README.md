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

## ✨ 功能特性
- ⚡ 极速体验：后端仅使用 Go 标准库构建，无任何外部框架依赖，响应极快。

- ⌨️ 编辑器增强：
    - 集成 Monaco Editor，支持 `Vim` 模式。

    - 内置代码高亮、自动对齐及 `Ctrl + Enter` 快捷运行。

- 🛠️ 算法库集成：通过 `git submodule` 引入 `goacm` 模板库，支持一键预览并插入代码。

- 🛡️ 安全与自愈：

    - 异常清理：绑定请求上下文，当用户刷新或断开连接时，自动清理整个进程组，严防僵尸进程。

    - 手动终止：提供停止按钮，可强行中断无限循环等异常程序。

    - 资源限制：Docker 部署环境下严格限制内存（默认 512M），确保宿主机稳定。

- ⌨️ 标准 IO：完美模拟标准输入（Stdin）与标准输出（Stdout/Stderr）交互。

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

## 🔒 安全说明

- 路径隔离：模板读取接口使用 `filepath.Clean` 并校验 `goacm` 前缀，彻底杜绝 `..` 目录穿越攻击。

- 进程组管理：运行任务使用 `syscall.SysProcAttr{Setpgid: true}`，确保 Kill 信号能覆盖所有子进程。

## 🤝 贡献指南
我们非常欢迎社区的贡献：

- 新增模板：请在 `goacm` 子模块中按类别提交 PR。

- 改进前端：保持纯原生 `JavaScript/HTML` 风格，不引入重型框架。

- 代码风格：遵循标准 `gofmt`。

## 📜 开源许可

本项目采用 [MIT License](LICENSE) 开源。