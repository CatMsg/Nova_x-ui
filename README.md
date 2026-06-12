# Nova_x-ui

Nova_x-ui 是3x-ui修改版

> **免责声明：** 本项目仅供个人学习与交流使用，请勿用于非法用途。

## 主要功能

- 面板总览：查看系统、Xray、流量和连接状态
- 入站管理：新增、编辑、删除、批量导入导出、客户端管理
- 节点管理：同步、测试和节点相关配置
- Xray 管理：运行控制、配置编辑、DNS 预设、出站测试、地理文件更新
- 面板设置：账号、默认端口、订阅与基础参数管理
- 订阅输出：订阅链接、订阅页和订阅 JSON / Clash 输出

## 本地运行

本地开发、调试或验证改动时，直接运行：

```bash
./run-local.sh
```

首次运行时，如果本机没有可用的 `xray`，脚本会自动下载对应平台的官方 release，并将可执行文件缓存到本地，再由 `bin/` 目录建立链接使用。

如果你要切换 Xray 版本，可以修改 `scripts/provision_xray.py` 里的 `DEFAULT_VERSION`，或者在运行前设置 `XRAY_VERSION`。

如果你希望把所有平台的 release zip 一次性缓存下来，可以运行：

```bash
python3 scripts/provision_xray.py --all
```

## 默认配置

| 项目 | 默认值 |
| --- | --- |
| 面板端口 | `2026` |
| 订阅端口 | `2027` |
| 命令 | `Nova_x-ui` |
| 数据库文件 | `Nova_x-ui.db` |
| Go module 路径 | `github.com/CatMsg/Nova_x-ui/v3` |

## 目录说明

| 路径 | 说明 |
| --- | --- |
| `frontend/` | Web 前端源码和构建产物 |
| `web/` | HTTP 路由、控制器、服务层和页面逻辑 |
| `xray/` | Xray 进程管理、配置和 API 交互 |
| `sub/` | 订阅链接、订阅页面和导出逻辑 |
| `database/` | 本地数据库初始化和模型 |
| `scripts/` | 本地开发辅助脚本 |
| `run-local.sh` | 本地开发启动入口 |

## 致谢

- [MHSanaei](https://github.com/MHSanaei/3x-ui)
