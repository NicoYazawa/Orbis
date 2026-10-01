"use client";

import { useCallback, useEffect, useRef, useState } from "react";

type SystemInfo = { service: string; version: string; status: string };
type Readiness = { status: string; dependencies?: Record<string, string> };
type Result<T> =
  | { kind: "loading" }
  | { kind: "success"; data: T }
  | { kind: "failure"; message: string; data?: T };

async function readJson<T>(path: string, signal: AbortSignal): Promise<{ response: Response; data: T }> {
  const response = await fetch(path, { cache: "no-store", signal });
  const data = (await response.json()) as T;
  return { response, data };
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "请求失败";
}

export function SystemStatus() {
  const [system, setSystem] = useState<Result<SystemInfo>>({ kind: "loading" });
  const [readiness, setReadiness] = useState<Result<Readiness>>({ kind: "loading" });
  const [checkedAt, setCheckedAt] = useState<string>();
  const controller = useRef<AbortController | null>(null);

  const refresh = useCallback(async () => {
    controller.current?.abort();
    const current = new AbortController();
    controller.current = current;
    setSystem({ kind: "loading" });
    setReadiness({ kind: "loading" });

    const [systemResult, readinessResult] = await Promise.allSettled([
      readJson<SystemInfo>("/api/v1/system", current.signal),
      readJson<Readiness>("/readyz", current.signal),
    ]);
    if (current.signal.aborted) return;

    if (systemResult.status === "fulfilled") {
      const { response, data } = systemResult.value;
      setSystem(response.ok && data.status === "ok"
        ? { kind: "success", data }
        : { kind: "failure", message: `HTTP ${response.status}`, data });
    } else {
      setSystem({ kind: "failure", message: errorMessage(systemResult.reason) });
    }

    if (readinessResult.status === "fulfilled") {
      const { response, data } = readinessResult.value;
      setReadiness(response.ok && data.status === "ready"
        ? { kind: "success", data }
        : { kind: "failure", message: `HTTP ${response.status}`, data });
    } else {
      setReadiness({ kind: "failure", message: errorMessage(readinessResult.reason) });
    }
    setCheckedAt(new Date().toLocaleTimeString("zh-CN", { hour12: false }));
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => void refresh(), 0);
    return () => {
      window.clearTimeout(timer);
      controller.current?.abort();
    };
  }, [refresh]);

  const loading = system.kind === "loading" || readiness.kind === "loading";
  const unreachable = system.kind === "failure" && !system.data || readiness.kind === "failure" && !readiness.data;
  const healthy = system.kind === "success" && readiness.kind === "success";
  const overall = loading ? "正在检查" : unreachable ? "无法连接" : healthy ? "运行正常" : "需要关注";
  const tone = loading ? "pending" : healthy ? "healthy" : "warning";

  return (
    <section className="status-section" aria-label="系统状态">
      <div className="status-heading">
        <div>
          <p className="eyebrow">LIVE HEALTH CHECK</p>
          <h2>当前工程状态</h2>
          <p className="section-copy">以下数据直接来自当前环境的 API 与就绪检查。</p>
        </div>
        <button className="refresh-button" type="button" onClick={() => void refresh()} disabled={loading}>
          <span aria-hidden="true">↻</span> 刷新状态
        </button>
      </div>

      <div className={`overall-banner ${tone}`} aria-live="polite">
        <span className="status-indicator" aria-hidden="true" />
        <span>{overall}</span>
        <span className="last-checked">{checkedAt ? `上次检查 ${checkedAt}` : "正在获取最新状态"}</span>
      </div>

      <div className="status-grid">
        <article className="status-card">
          <div className="card-topline"><span className="card-number">01 / API</span><span className="card-symbol" aria-hidden="true">↗</span></div>
          <h3>服务信息</h3>
          {system.kind === "loading" && <p className="card-state">检查中…</p>}
          {system.kind === "success" && <>
            <p className="card-state good">运行中</p>
            <dl className="detail-list">
              <div><dt>服务</dt><dd>{system.data.service}</dd></div>
              <div><dt>版本</dt><dd>{system.data.version}</dd></div>
              <div><dt>状态</dt><dd>{system.data.status}</dd></div>
            </dl>
          </>}
          {system.kind === "failure" && <>
            <p className="card-state bad">读取失败</p>
            <p className="error-copy">{system.message}</p>
          </>}
          <p className="endpoint">GET /api/v1/system</p>
        </article>

        <article className="status-card">
          <div className="card-topline"><span className="card-number">02 / READINESS</span><span className="card-symbol" aria-hidden="true">◎</span></div>
          <h3>依赖就绪</h3>
          {readiness.kind === "loading" && <p className="card-state">检查中…</p>}
          {readiness.kind === "success" && <>
            <p className="card-state good">就绪</p>
            <p className="card-copy">服务依赖检查已通过，可以接收请求。</p>
          </>}
          {readiness.kind === "failure" && <>
            <p className="card-state bad">未就绪</p>
            <p className="error-copy">{readiness.message}</p>
            {readiness.data?.dependencies && <ul className="dependency-list">
              {Object.entries(readiness.data.dependencies).map(([name, status]) => <li key={name}><span>{name}</span><span>{status}</span></li>)}
            </ul>}
          </>}
          <p className="endpoint">GET /readyz</p>
        </article>
      </div>
    </section>
  );
}
