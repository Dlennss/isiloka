import { NextResponse } from "next/server";

export async function POST(req: Request) {
  await req.text();
  return NextResponse.json(
    {
      ok: false,
      error: "Top up provider harus dari mutasi bank rekening. Gunakan menu Mutasi Bank lalu assign debit bank ke provider.",
    },
    { status: 410 }
  );
}
