import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SystemStatus } from "../app/system-status";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("SystemStatus", () => {
  it("shows real system details and readiness from same-origin endpoints", async () => {
    const fetch = vi.fn(async (input: string) => {
      if (input === "/api/v1/system") return new Response(JSON.stringify({ service: "api", version: "dev", status: "ok" }), { status: 200 });
      if (input === "/readyz") return new Response(JSON.stringify({ status: "ready" }), { status: 200 });
      throw new Error(`Unexpected URL: ${input}`);
    });
    vi.stubGlobal("fetch", fetch);

    render(<SystemStatus />);

    expect(await screen.findByText("运行正常")).toBeInTheDocument();
    expect(screen.getByText("dev")).toBeInTheDocument();
    expect(screen.getByText("就绪")).toBeInTheDocument();
    expect(fetch).toHaveBeenCalledWith("/api/v1/system", expect.objectContaining({ cache: "no-store" }));
    expect(fetch).toHaveBeenCalledWith("/readyz", expect.objectContaining({ cache: "no-store" }));
  });

  it("reports degraded readiness and can refresh to recovered state", async () => {
    let ready = false;
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/v1/system") return new Response(JSON.stringify({ service: "api", version: "dev", status: "ok" }), { status: 200 });
      if (input === "/readyz") return new Response(JSON.stringify(ready ? { status: "ready" } : { status: "degraded", dependencies: { postgres: "down" } }), { status: ready ? 200 : 503 });
      throw new Error(`Unexpected URL: ${input}`);
    }));

    render(<SystemStatus />);
    expect(await screen.findByText("未就绪")).toBeInTheDocument();
    expect(screen.getByText("postgres")).toBeInTheDocument();

    ready = true;
    await userEvent.click(screen.getByRole("button", { name: "刷新状态" }));
    expect(await screen.findByText("就绪")).toBeInTheDocument();
    expect(screen.queryByText("postgres")).not.toBeInTheDocument();
  });

  it("reports request errors rather than claiming that the system is healthy", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => { throw new Error("Network unavailable"); }));

    render(<SystemStatus />);

    expect(await screen.findByText("无法连接")).toBeInTheDocument();
    expect(screen.getAllByText(/Network unavailable/)).toHaveLength(2);
  });
});
