# Weiyun API V3 (MCP Wrapper)

[English](README.md) | **中文**

[![CI Build and Test](https://github.com/LiuZixia/weiyun-api-v3/actions/workflows/ci.yml/badge.svg)](https://github.com/LiuZixia/weiyun-api-v3/actions/workflows/ci.yml)

> **注意**：自动集成测试会在您网盘根目录下一个独立的 `/CI` 文件夹中执行，以防污染您的个人文件。文件传输测试采用了完全严格的验证规范（例如生成独立的 `tencent-weiyun.zip` 或 `test_{language}_upload_{timestamp}` 等验证文件进行上传下载与删除的全生命周期管理）。

腾讯微云 V3 MCP API 的全面多语言实现。该库允许开发者将微云存储无缝集成到应用、AI 智能体（Agent）和自动化工作流中。

## 🌟 核心特性

* **完整的文件管理体系**：目录拉取、批量获取直链下载、回收与彻底删除、分享外链生成。
* **高阶的稳定上传协议**：完整实现了官方的 FTN 双段异步协议（两阶段校验：预传 + 分片秒传校验）。
* **原生独家的 SHA1 算法内核**：高度还原封装了微云服务器强制要求提取的“小端序内部寄存器状态”（Little-Endian SHA1 Register State）的提取能力。
* **原生多语言驱动**：提供了独立、原生、无缝对接的 Python、Go、PHP 和 Shell 客户端实现。
* **AI 架构原生兼容**：该设计从基底上就完美对标模型上下文协议 (Model Context Protocol, MCP)，能够被诸如 Claude、GPT 等直接调用。

---

## 🛠 架构详解：上传协议的核心秘密

本 API 中最核心和复杂的部分当属 `weiyun.upload` 上传工具。微云不像那些提供通用标准 S3 协议的云盘，Weiyun 的架构有着非常底层硬核的校验机制：

1.  **块间校验 (Intermediate Blocks)**：客户端必须在计算常规文件的每 `512KB`（即 524288 字节）阶段，手动对外抛出当前哈希缓冲区的内部小端序寄存器（$h0, h1, h2, h3, h4$）状态。我们已在各大语言中自行实现了这套解析。
2.  **收尾块 (Final Block)**：整个文件的标准 Big-Endian SHA1 Hexdigest。
3.  **防抖校验块 (Check SHA)**：计算直到文末 128 字节之前的那个状态阶段结果，加上文件尾部最后一点点的原始 Base64 取证。

---

## 🚀 快速上手

### 环境准备

使用所有接口前，您必须先要拿到一个微云分配的 `mcp_token` 才可以。可以前往 [Weiyun Authorization Page](https://www.weiyun.com/act/openclaw) 扫码获取。

### 设置环境变量

为了防止密钥硬编码在代码中泄露，我们统一通过系统变量下发 Token：

```bash
export WEIYUN_MCP_TOKEN="填写您的_mcp_token"
export WEIYUN_MCP_URL="https://www.weiyun.com/api/v3/mcpserver"
```

---

## 📦 各语言快速指引

每种语言目录下都有详细独立的 `README.md` 和单独配置的 `.env.example`。建议您点进各自的文件夹中查看完整示例，这里只做技术概述：

### 1. Python 实现篇
包含了一个为了剥离 `hashlib` 封装而采用纯 Python 手写的自定义 `SHA1` 类，借此从底层释放出微云所需要的内部 Register 指针。
👉 [详情](python/README.md)

### 2. Go 高性能篇
直接打破了 Go 语言底层的封装安全限制，采用了 `reflect` 反射与 `unsafe.Pointer` 指针逃逸机制抓取了标准库 `crypto/sha1` 内隐匿的 `digest` 结构实体。性能极强。
👉 [详情](go/README.md)

### 3. PHP Web 篇
基于原生的 PHP 语法撰写了支持严格溢出边界控制（`& 0xFFFFFFFF` 位掩码）和防长数值溢出的 SHA1 旋转循环体系，专门应对 Web 服务器环境下 32 位与 64 位整型大小不一致的兼容陷阱。
👉 [详情](php/README.md)

### 4. Shell / DevOps 篇
原生打通了全局命令行的环境变量配置和与 `mcporter` NPM 核心包绑定的全栈指令方案。这套体系适合配置到 CI/CD 流程中作为自动备份等脚本工具运行。
👉 [详情](shell/README.md)

---

## 🤝 贡献说明

非常欢迎各种 Issue 及 PR。本项目自带了完善的多环境集成测试（Integration Tests），所有提至 Main 分支的代码变更都会在 GitHub Actions 的容器集群中全量模拟与云端的高强度传输认证握手。

## 📄 协议开源许可

代码遵守 [MIT License](https://opensource.org/licenses/MIT) 开源，请放心使用。
