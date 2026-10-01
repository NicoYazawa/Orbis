# Orbis 本地 Docker 镜像清单

> 采集日期：2026-10-01（Asia/Shanghai）
> Docker Desktop：4.91.0；Engine/CLI：29.8.0；Compose：v5.5.1；context：desktop-linux；服务端：linux/amd64。
> 来源：docker version、docker compose version、docker image ls、docker image inspect。只读查询，未拉取、启动、停止或修改容器。

本清单证明镜像存在于当前 Docker Engine，不证明 Orbis 集成或健康检查通过。ID 与 RepoDigest 分开记录；部署锁定使用 RepoDigest。大小为 inspect 返回的 SizeBytes 换算 MiB，不代表去重后的磁盘占用。

| 本地镜像标签 | 镜像 ID | RepoDigest（部署锁定引用） | 平台 | 大小 MiB |
|---|---|---|---|---|
| postgres:16-alpine | sha256:721873c34ceb9f8d8fc265984940dc982404c105f19ad51be9fdc5970a6080ea | postgres@sha256:721873c34ceb9f8d8fc265984940dc982404c105f19ad51be9fdc5970a6080ea | linux/amd64 | 399.7 |
| redis:7.4-alpine | sha256:ff02b58f971e7d7d156a1267e283fcbbeee91773b6aa36c49dac28ecfe28eadf | redis@sha256:ff02b58f971e7d7d156a1267e283fcbbeee91773b6aa36c49dac28ecfe28eadf | linux/amd64 | 54.7 |
| ghcr.io/rustfs/rustfs:latest | sha256:8cc9801755448b71a786705ce76692c77e14936cccd87cf2fc31842e58f4d1ff | ghcr.io/rustfs/rustfs@sha256:8cc9801755448b71a786705ce76692c77e14936cccd87cf2fc31842e58f4d1ff | linux/amd64 | 381.3 |
| qdrant/qdrant:v1.19.0 | sha256:057ee3a8da769fe7310dd3537b4dc7583bf87a95ce8ac43c0af5a46bc580d1fc | qdrant/qdrant@sha256:057ee3a8da769fe7310dd3537b4dc7583bf87a95ce8ac43c0af5a46bc580d1fc | linux/amd64 | 257.4 |
| ollama/ollama:latest | sha256:8262851b2846b87c649eddf3e76beb270c52f4d1bc94559f47efde16b0841551 | ollama/ollama@sha256:8262851b2846b87c649eddf3e76beb270c52f4d1bc94559f47efde16b0841551 | linux/amd64 | 8847.4 |
| searxng/searxng:latest | sha256:eb2a404db174d33a04cd9f2e519bcae2ca722a58923510652cc1314be94e8e63 | searxng/searxng@sha256:eb2a404db174d33a04cd9f2e519bcae2ca722a58923510652cc1314be94e8e63 | linux/amd64 | 365.3 |
| jaegertracing/all-in-one:1.60 | sha256:4fd2d70fa347d6a47e79fcb06b1c177e6079f92cba88b083153d56263082135e | jaegertracing/all-in-one@sha256:4fd2d70fa347d6a47e79fcb06b1c177e6079f92cba88b083153d56263082135e | linux/amd64 | 107.3 |
| jaegertracing/all-in-one:latest | sha256:ab6f1a1f0fb49ea08bcd19f6b84f6081d0d44b364b6de148e1798eb5816bacac | jaegertracing/all-in-one@sha256:ab6f1a1f0fb49ea08bcd19f6b84f6081d0d44b364b6de148e1798eb5816bacac | linux/amd64 | 117.6 |
| prom/prometheus:v3.1.0 | sha256:6559acbd5d770b15bb3c954629ce190ac3cbbdb2b7f1c30f0385c4e05104e218 | prom/prometheus@sha256:6559acbd5d770b15bb3c954629ce190ac3cbbdb2b7f1c30f0385c4e05104e218 | linux/amd64 | 395 |
| grafana/grafana:11.3.0 | sha256:a0f881232a6fb71a0554a47d0fe2203b6888fe77f4cefb7ea62bed7eb54e13c3 | grafana/grafana@sha256:a0f881232a6fb71a0554a47d0fe2203b6888fe77f4cefb7ea62bed7eb54e13c3 | linux/amd64 | 616 |
| golang:1.27-alpine | sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 | golang@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 | linux/amd64 | 363.3 |
| golang:1.27-bookworm | sha256:69a7b9788769bec032d238959b61854e9ae87f57be9029ec04e9885fabf99195 | golang@sha256:69a7b9788769bec032d238959b61854e9ae87f57be9029ec04e9885fabf99195 | linux/amd64 | 1184.8 |
| node:24-alpine | sha256:ebfe2f90462722a7a4de65e91990e97fe0d401c70e0e762c5b53302f905ec1c1 | node@sha256:ebfe2f90462722a7a4de65e91990e97fe0d401c70e0e762c5b53302f905ec1c1 | linux/amd64 | 230.3 |
| alpine:3.23.5 | sha256:fd791d74b68913cbb027c6546007b3f0d3bc45125f797758156952bc2d6daf40 | alpine@sha256:fd791d74b68913cbb027c6546007b3f0d3bc45125f797758156952bc2d6daf40 | linux/amd64 | 12.3 |
| nginx:1.31.5-alpine | sha256:72ba65eb42c10344912a84ff42408db7d34f2feb642204570ab8fc5ffd29f1d3 | nginx@sha256:72ba65eb42c10344912a84ff42408db7d34f2feb642204570ab8fc5ffd29f1d3 | linux/amd64 | 97.3 |
| minio/mc:latest | sha256:a7fe349ef4bd8521fb8497f55c6042871b2ae640607cf99d9bede5e9bdf11727 | minio/mc@sha256:a7fe349ef4bd8521fb8497f55c6042871b2ae640607cf99d9bede5e9bdf11727 | linux/amd64 | 111.8 |
| testcontainers/ryuk:0.14.0 | sha256:7c1a8a9a47c780ed0f983770a662f80deb115d95cce3e2daa3d12115b8cd28f0 | testcontainers/ryuk@sha256:7c1a8a9a47c780ed0f983770a662f80deb115d95cce3e2daa3d12115b8cd28f0 | linux/amd64 | 4.6 |
| python:3.11-slim | sha256:e41613d42d4891e4930f79523f93f81bbc7632584ec65e36ab055f41a800b41e | python@sha256:e41613d42d4891e4930f79523f93f81bbc7632584ec65e36ab055f41a800b41e | linux/amd64 | 178.3 |

## 使用范围与版本解释

- PostgreSQL 16、Redis 7.4、Qdrant v1.19.0 为当前方案的本地候选；PG 18、Redis 8 等其他项目版本不自动替代它们。
- RustFS、Ollama、SearXNG、mc 的 latest 是可变标签，按清单 RepoDigest 固定内容。初次清单采集未启动二进制；后续 Phase 0 实测 RustFS1.0.0。Ollama/SearXNG/mc 的未知应用版本不作推断。
- SearXNG 的 OCI version 标签为 2026.9.29-f5035873a，可作为镜像元数据佐证。Ollama 的 OCI version 为 24.04，可能来自基础镜像，不能当作 Ollama 版本。
- Jaeger 1.60 与 latest 是两个不同 digest；先保留 1.60 为开发候选，生产选型需单独验证维护状态和 OTLP 集成。
- Go 1.27-alpine 用于普通构建；1.27-bookworm 可用于需要 CGO/竞态检测的 Linux 测试环境。
- Node 24-alpine、Alpine 3.23.5、Nginx 1.31.5-alpine 为构建/运行/代理候选，不代表应用镜像已构建。
- minio/mc 用于 S3 手工诊断；testcontainers/ryuk:0.14.0 是本地清理辅助镜像，实际 Testcontainers 依赖可能要求不同版本，应按其依赖锁定，不强制覆盖。

## Phase 0 后续实测

2026-10-01：以上固定digest对应的PostgreSQL16.15、Redis7.4.11、RustFS1.0.0、Qdrant1.19.0、Go1.27.1、Node24.21.0、Prometheus3.1.0、Grafana11.3.0通过镜像内命令或服务API确认。基础/rag/obs/app集成与SDK验收通过，详细验收记录仅保留在本地。Ollama与SearXNG未启动；其模型和搜索业务留待后续阶段。
- python:3.11-slim 为可选契约测试运行环境，安装 Schemathesis 前验证 Python 兼容性并锁定依赖。
- 本次清单未发现 k6、Playwright 或 Toxiproxy 专用镜像；对应工具规划使用 CLI/npm/Go 依赖或独立测试镜像，未声明已安装。

原始本机选取字段保留在本地，不提交远程；再次查询时更新公开版本清单与开发计划，并将原始运行记录留在被忽略的目录。
