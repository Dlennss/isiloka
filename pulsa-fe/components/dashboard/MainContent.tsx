type Props = {
  children: React.ReactNode;
};

export function MainContent({ children }: Props) {
  return (
    <main className="min-w-0 flex-1 px-2 py-5 md:px-2 md:py-4">
        {children}
    </main>
  );
}

