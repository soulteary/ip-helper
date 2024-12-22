# IP Helper

一个功能强大的多协议 IP 信息查询服务器。

## 主要特性

- **多协议支持**
  - HTTP/HTTPS Web 界面访问
  - CURL/Wget 命令行工具查询
  - Telnet 协议查询
  - FTP 协议查询
- **丰富功能**
  - IP 地理位置查询
  - Token 认证机制
  - Gzip 压缩支持
  - 健康检查接口
  - 自定义域名和端口

## 快速开始

### 环境要求

- Go 1.x 或更高版本

### 安装部署

```bash
git clone https://github.com/soulteary/ip-helper.git
cd ip-helper
go build
```

### 启动服务

```bash
./ip-helper
```

## 配置说明

支持通过环境变量或命令行参数进行配置:

| 配置项 | 环境变量 | 命令行参数 | 默认值 | 说明 |
|--------|----------|------------|---------|------|
| 调试模式 | DEBUG | -debug | false | 是否启用调试日志 |
| 服务端口 | SERVER_PORT | -port | 8080 | HTTP 服务监听端口 |
| 服务域名 | SERVER_DOMAIN | -domain | http://localhost:8080 | 服务访问域名 |
| 访问令牌 | TOKEN | -token | "" | API 访问认证令牌 |

## API 使用说明

### Web 界面查询

直接访问 `http://localhost:8080` 使用 Web 界面进行查询。

### 命令行查询

提供多种命令行查询方式:

```bash
# curl 方式
curl http://localhost:8080

# telnet 方式  
telnet localhost 8080

# ftp 方式
ftp localhost 8080
```

### API 认证

如果配置了访问令牌,可通过以下方式携带:

```bash
# URL 参数方式
curl http://localhost:8080?token=your_token

# Header 方式
curl -H "X-Token: your_token" http://localhost:8080
```

## 开源协议

本项目采用 MIT 开源协议。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来帮助改进项目。

## 相关链接

- [项目主页](https://github.com/soulteary/ip-helper)
- [问题反馈](https://github.com/soulteary/ip-helper/issues)
- [开发文档](https://github.com/soulteary/ip-helper/wiki)
