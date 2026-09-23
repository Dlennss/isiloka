import path from "path";
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: process.platform === "win32" ? undefined : "standalone",
  turbopack: {
    root: path.resolve(__dirname),
  },
  async redirects() {
    return [
      {
        source: "/admin",
        destination: "/dashboard/admin",
        permanent: false,
      },
      {
        source: "/admin/:path*",
        destination: "/dashboard/admin/:path*",
        permanent: false,
      },
    ];
  },
};

export default nextConfig;
