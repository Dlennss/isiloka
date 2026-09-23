type Props = { children: React.ReactNode };
export function BackgroundAuth({ children }: Props) {
  return <main className="brand-auth-shell flex min-h-svh items-center justify-center px-4 py-8"><div className="w-full max-w-[390px]">{children}</div></main>;
}
