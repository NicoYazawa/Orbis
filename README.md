# Orbis

Orbis 是用于构建、运行、调试和评测 Agent 的新平台。Phase 0 工程与基础设施已完成本机验收；登录、Agent、聊天、工作流和评测尚未实现。

## 本地启动

要求 Docker Desktop、Docker Compose 和 Node.js 24 或更新版本。

```powershell
node scripts/orbis.mjs init
./scripts/start-infra.ps1
```

打开 http://127.0.0.1:18080。默认仅发布这个回环端口，其他服务走独立 Orbis Docker 内网。启动前检查端口；冲突时只调整 `.env` 的 `ORBIS_WEB_PORT`，不停止其他项目。

```powershell
./scripts/verify-infra.ps1
./scripts/stop-infra.ps1
```

停止保留 Orbis 数据卷。初始化自动生成被忽略的 `.env`，不会打印凭证。

开发文档与本机验收记录仅保留在本地，不纳入远程仓库。
