export const API_BASE = process.env.NEXT_PUBLIC_API_BASE?.replace(/\/+$/, '') || '';
if (!API_BASE) {
  // Biarkan terlihat saat dev agar tidak bingung ketika lupa set env
  // (Tidak melempar error, hanya warning di console)
  console.warn('NEXT_PUBLIC_API_BASE is not set');
}
