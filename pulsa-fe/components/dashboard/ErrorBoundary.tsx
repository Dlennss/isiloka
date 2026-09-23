"use client";

import { Component, type ReactNode } from "react";

interface Props { children: ReactNode; }
interface State { hasError: boolean; error?: Error; }

export class DashboardErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-[50vh] items-center justify-center p-8">
          <div className="max-w-md space-y-4 text-center">
            <div className="text-lg font-semibold text-white">Terjadi Kesalahan</div>
            <div className="text-sm text-white/60">
              Halaman mengalami error. Coba muat ulang halaman.
            </div>
            <button
              onClick={() => {
                this.setState({ hasError: false, error: undefined });
                window.location.reload();
              }}
              className="rounded-xl border border-white/10 bg-white/[0.04] px-6 py-2 text-sm text-white hover:bg-white/10"
            >
              Muat Ulang
            </button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
