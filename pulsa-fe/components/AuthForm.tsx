'use client';
 
import * as React from 'react';
import { register as apiRegister, login as apiLogin } from '@/lib/auth';
import { useRouter } from 'next/navigation';
import { getErrorMessage } from '@/lib/error';

type Props = { mode: 'register' | 'login' };

export default function AuthForm({ mode }: Props) {
  const router = useRouter();
  const [email, setEmail] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [showPass, setShowPass] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [info, setInfo] = React.useState<string | null>(null);

  const title = mode === 'register' ? 'Daftar Akun' : 'Masuk';
  const buttonLabel = mode === 'register' ? 'Daftar' : 'Login';
  const oppositeHref = mode === 'register' ? '/login' : '/register';
  const oppositeLabel =
    mode === 'register' ? 'Sudah punya akun? Login' : 'Belum punya akun? Daftar';

  function isValidEmail(v: string) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v.trim());
  }

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setInfo(null);

    const em = email.trim().toLowerCase();
    const pw = password.trim();

    if (!isValidEmail(em)) { setError('Format email tidak valid'); return; }
    if (!pw) { setError('Password wajib diisi'); return; }

    setLoading(true);
    try {
      if (mode === 'register') {
        await apiRegister({ email: em, password: pw });
        setInfo('Registrasi berhasil. Silakan login.');
        setTimeout(() => router.push('/login'), 1000);
      } else {
        const { token } = await apiLogin({ email: em, password: pw });
        localStorage.setItem('auth_token', token);
        router.push('/dashboard');
      }
    } catch (err: unknown) {
      setError(getErrorMessage(err) || 'Terjadi kesalahan');
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={onSubmit} className="auth-card" aria-busy={loading}>
      <h1>{title}</h1>

      {error && <div role="alert" className="alert error">{error}</div>}
      {info && <div role="status" className="alert info">{info}</div>}

      <label>
        <span>Email</span>
        <input
          type="email"
          value={email}
          autoComplete="email"
          onChange={(e) => setEmail(e.target.value)}
          placeholder="nama@domain.com"
          required
        />
      </label>

      <label>
        <span>Password</span>
        <div className="password-field">
          <input
            type={showPass ? 'text' : 'password'}
            value={password}
            autoComplete={mode === 'register' ? 'new-password' : 'current-password'}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="********"
            required
            minLength={1}
          />
          <button
            type="button"
            className="toggle"
            aria-label={showPass ? 'Sembunyikan password' : 'Lihat password'}
            onClick={() => setShowPass((s) => !s)}
          >
            {showPass ? 'Hide' : 'Show'}
          </button>
        </div>
      </label>

      <button type="submit" disabled={loading}>
        {loading ? 'Memproses...' : buttonLabel}
      </button>

      <p className="muted">
        <a href={oppositeHref}>{oppositeLabel}</a>
      </p>
    </form>
  );
}
