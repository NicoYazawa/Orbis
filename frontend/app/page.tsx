import { SystemStatus } from "./system-status";

export default function Home() {
  return (
    <main className="site-shell">
      <header className="site-header">
        <div className="brand"><span className="brand-mark" aria-hidden="true">✳</span><span>ORBIS</span></div>
        <div className="header-label"><span className="header-dot" /> 工程控制台 · Phase 0</div>
      </header>

      <section className="hero">
        <div className="hero-copy">
          <p className="eyebrow">FOUNDATION / 00</p>
          <h1>Orbis <span>工程已启动。</span></h1>
          <p className="hero-description">构建、运行与观测 Agent 的平台正在搭建中。这里展示当前环境的实时服务状态，作为后续功能的可靠起点。</p>
        </div>
        <div className="hero-art" aria-hidden="true"><div className="orb orb-one" /><div className="orb orb-two" /><div className="orb orb-three" /><div className="orb-core" /></div>
      </section>

      <SystemStatus />

      <footer className="site-footer">
        <span>ORBIS / PHASE 0</span>
        <span>基础设施与工程验证阶段</span>
      </footer>
    </main>
  );
}
