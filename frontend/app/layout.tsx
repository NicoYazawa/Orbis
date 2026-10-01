import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Orbis · 工程状态",
  description: "Orbis Phase 0 工程状态与服务就绪检查",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="zh-CN"><body>{children}</body></html>;
}
