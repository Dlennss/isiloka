export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-svh bg-[#13464b] text-slate-950">
      {children}
    </div>
  );
}
