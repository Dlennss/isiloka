function normalizeSiteUrl(raw?: string) {
  if (!raw) return null;
  if (raw.startsWith("https://") || raw.startsWith("http://")) return raw.replace(/\/$/, "");
  if (raw.startsWith("https:")) return raw.replace(/^https:/, "https://").replace(/\/$/, "");
  if (raw.startsWith("http:")) return raw.replace(/^http:/, "http://").replace(/\/$/, "");
  return `https://${raw.replace(/^\/+/, "").replace(/\/$/, "")}`;
}

export const SITE_URL =
  normalizeSiteUrl(process.env.NEXT_PUBLIC_SITE_URL) ??
  normalizeSiteUrl(process.env.NEXTAUTH_URL) ??
  "http://localhost:3102";

export const CANONICAL_SITE_URL = SITE_URL;
export const DEFAULT_OG_IMAGE_URL = `${CANONICAL_SITE_URL}/images/logo-pulsakilat.svg`;

export type ArticleSection = {
  heading: string;
  body: string[];
};

export type ArticleFaq = {
  question: string;
  answer: string;
};

export type SeoArticle = {
  slug: string;
  title: string;
  description: string;
  excerpt: string;
  category: string;
  readTime: string;
  keywords: string[];
  hero: string;
  products: string[];
  ctaLabel: string;
  ctaHref: string;
  publishedAt: string;
  updatedAt: string;
  sections: ArticleSection[];
  faqs: ArticleFaq[];
};

export type QuickLink = {
  label: string;
  href: string;
};

const publishedAt = "2026-03-31T00:00:00.000Z";
const updatedAt = "2026-03-31T00:00:00.000Z";

export const suggestedTopics: string[] = [
  "Cara membangun katalog produk digital yang mudah dipahami pelanggan baru",
  "Cara mengatur layanan pelanggan untuk usaha pulsa dan PPOB",
  "Cara menggabungkan pulsa, paket data, dan e-wallet dalam satu usaha",
  "Panduan mengelola pelanggan rutin untuk listrik, BPJS, dan PDAM",
  "Cara membangun usaha top up game yang cepat dan rapi",
  "Cara memanfaatkan deposit dan saldo agar operasional lancar",
  "Tips memilih kategori produk paling cocok untuk toko digital Anda",
  "Cara memperluas usaha dari jualan manual ke website sendiri",
  "Checklist kesiapan bisnis sebelum masuk ke integrasi H2H",
  "Keuntungan menjual produk digital lengkap dalam satu platform",
  "Cara menyusun layanan untuk pelanggan retail, warung, dan komunitas",
  "Panduan menentukan fokus usaha: transaksi harian atau bisnis integrasi",
];

export const seoArticles: SeoArticle[] = [
  {
    slug: "cara-isi-pulsa-online-semua-operator",
    title: "Cara Isi Pulsa Online Semua Operator dengan Cepat dan Aman",
    description: "Panduan isi pulsa online Telkomsel, Indosat, XL, Axis, Smartfren, dan Tri lewat Isiloka dengan alur yang cepat dan mudah dipahami.",
    excerpt: "Isi pulsa online paling nyaman jika operator, nominal, dan pembayaran sudah tersusun jelas sejak awal.",
    category: "Pulsa",
    readTime: "6 menit",
    keywords: ["cara isi pulsa online", "isi pulsa telkomsel", "isi pulsa indosat", "pulsa cepat"],
    hero: "Pulsa tetap jadi produk digital harian yang paling mudah dijual dan paling sering dicari pelanggan.",
    products: ["Pulsa Telkomsel", "Pulsa Indosat", "Pulsa XL", "Pulsa Smartfren"],
    ctaLabel: "Isi Pulsa Sekarang",
    ctaHref: "/pulsa",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa isi pulsa online tetap relevan",
        body: [
          "Pulsa masih dipakai untuk kebutuhan telepon, SMS, pembelian paket, dan banyak transaksi kecil lain. Karena itu, layanan ini tetap jadi kebutuhan harian yang paling mudah dijual.",
          "Isiloka cocok untuk kebutuhan ini karena pengguna bisa langsung masuk ke kategori pulsa tanpa harus melewati langkah yang rumit.",
        ],
      },
      {
        heading: "Langkah isi pulsa di Isiloka",
        body: [
          "Masuk ke halaman pulsa, pilih operator sesuai nomor tujuan, masukkan nomor HP, lalu pilih nominal yang dibutuhkan. Setelah itu lanjut ke pembayaran dan pantau status transaksi.",
          "Jika pembeli sudah tahu operator seperti Telkomsel atau Indosat, halaman operator khusus akan mempercepat pencarian nominal yang tepat.",
        ],
      },
      {
        heading: "Kenapa pulsa jadi produk dasar untuk member",
        body: [
          "Pulsa adalah produk yang paling mudah dipahami pelanggan baru. Karena itu, kategori ini cocok dijadikan titik awal untuk calon member yang baru membangun usaha.",
          "Dari pulsa, member bisa berkembang ke paket data, e-wallet, token listrik, dan kategori lain tanpa harus mengubah pola layanan dari nol.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah semua operator besar tersedia?", answer: "Ya. Isiloka mendukung operator utama seperti Telkomsel, Indosat, XL, Axis, Smartfren, Tri, dan brand terkait yang aktif di katalog." },
      { question: "Apakah isi pulsa harus login dulu?", answer: "Tidak selalu. Banyak alur pembelian bisa dimulai dari halaman guest, meski akun tetap lebih baik untuk riwayat dan saldo." },
    ],
  },
  {
    slug: "cara-beli-paket-data-harian-dan-bulanan",
    title: "Cara Beli Paket Data Harian dan Bulanan yang Lebih Hemat",
    description: "Panduan membeli paket data harian, mingguan, dan bulanan melalui Isiloka dengan grup produk yang lebih jelas dan mudah dipilih.",
    excerpt: "Paket data yang tepat bukan hanya soal kuota besar, tetapi juga soal masa aktif dan kejelasan pilihan produk.",
    category: "Paket Data",
    readTime: "7 menit",
    keywords: ["cara beli paket data", "paket data telkomsel", "paket internet murah", "top up kuota"],
    hero: "Kategori paket data membutuhkan tampilan yang rapi karena variasi produknya jauh lebih banyak daripada pulsa.",
    products: ["Paket Data Telkomsel", "Paket Data Indosat", "Paket Data XL", "Paket Data by.U"],
    ctaLabel: "Buka Paket Data",
    ctaHref: "/paket-data",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Cara membaca grup paket data",
        body: [
          "Paket data perlu dipisah per grup seperti promo, bulanan, mini, regional, atau nama khusus operator. Ini membantu pembeli menyaring pilihan sebelum membandingkan paket.",
          "Di Isiloka, grup produk dan judul kartu yang lebih ringkas membuat pembeli lebih cepat memahami isi paket seperti 1 GB 7 Hari atau 12 GB 30 Hari.",
        ],
      },
      {
        heading: "Langkah memilih paket yang tepat",
        body: [
          "Masukkan nomor tujuan, pastikan operator sudah sesuai, lalu buka grup paket yang relevan. Setelah itu pilih paket berdasarkan kuota, masa aktif, dan harga final.",
          "Strategi ini lebih aman daripada hanya mengejar nama paket yang terdengar besar tetapi belum tentu cocok dengan pola pemakaian.",
        ],
      },
      {
        heading: "Kenapa paket data penting untuk member",
        body: [
          "Paket data memberi peluang repeat order yang tinggi karena kebutuhan internet pelanggan terus berulang. Ini membuat kategori kuota sangat penting untuk member yang ingin transaksi stabil.",
          "Karena pilihannya banyak, tampilan grup dan judul produk yang jelas akan membantu calon pembeli lebih cepat menentukan paket yang cocok.",
        ],
      },
    ],
    faqs: [
      { question: "Kenapa judul paket data harus ringkas?", answer: "Agar pembeli langsung paham kuota dan masa aktif tanpa terganggu pengulangan nama grup yang terlalu panjang." },
      { question: "Apakah grup produk penting untuk paket data?", answer: "Ya. Grup memudahkan pencarian dan mengurangi kebingungan saat pilihan produknya banyak." },
    ],
  },
  {
    slug: "cara-top-up-ewallet-dana-ovo-gopay-linkaja",
    title: "Cara Top Up DANA, OVO, GoPay, dan LinkAja di Isiloka",
    description: "Panduan top up e-wallet melalui Isiloka, termasuk nominal tetap, bebas nominal, dan cara membaca biaya secara lebih jelas.",
    excerpt: "Top up e-wallet akan lebih nyaman jika nominal, biaya, dan varian produk ditampilkan dengan jelas sejak awal.",
    category: "E-Wallet",
    readTime: "7 menit",
    keywords: ["top up dana", "top up ovo", "top up gopay", "top up linkaja"],
    hero: "Kategori e-wallet sangat cocok untuk transaksi cepat dan repeat order tinggi.",
    products: ["DANA", "OVO", "GoPay", "LinkAja", "ShopeePay"],
    ctaLabel: "Top Up E-Wallet",
    ctaHref: "/ewallet",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Nominal tetap dan bebas nominal",
        body: [
          "Nominal tetap cocok untuk transaksi cepat karena pembeli tinggal memilih 10.000, 20.000, dan seterusnya. Bebas nominal lebih fleksibel, tetapi judul produknya tidak boleh dibuat terlalu umum.",
          "Varian seperti Bebas Nominal Direct harus tetap bisa dibedakan agar pembeli tidak tertukar.",
        ],
      },
      {
        heading: "Cara top up yang lebih aman",
        body: [
          "Pilih brand e-wallet, pastikan nomor atau akun tujuan benar, lalu pilih nominal yang paling sesuai. Setelah itu cek ringkasan pembayaran dan lanjutkan checkout.",
          "Kartu e-wallet yang hanya menonjolkan nominal dan menaruh harga di bawah membuat keputusan pengguna lebih cepat.",
        ],
      },
      {
        heading: "Kenapa e-wallet penting untuk usaha member",
        body: [
          "Top up DANA, OVO, GoPay, dan LinkAja dibutuhkan sangat sering oleh pelanggan harian. Karena itu, kategori e-wallet cocok dijadikan produk inti bersama pulsa dan paket data.",
          "Bagi member, e-wallet membantu menarik pelanggan yang ingin transaksi cepat dan membuka peluang repeat order yang tinggi.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah produk bebas nominal harus dibedakan dari nominal tetap?", answer: "Ya. Varian bebas nominal seperti LinkAja atau DANA dengan nominal terbuka harus dibedakan dari nominal tetap agar pembeli tidak salah pilih." },
      { question: "Brand e-wallet apa yang paling sering dicari?", answer: "Biasanya DANA, OVO, GoPay, dan LinkAja paling sering dicari karena dipakai untuk belanja, transfer, dan top up kebutuhan harian." },
    ],
  },
  {
    slug: "cara-top-up-game-mobile-legends-free-fire-pubg",
    title: "Cara Top Up Game Mobile Legends, Free Fire, dan PUBG yang Praktis",
    description: "Panduan top up game melalui Isiloka untuk pemain dan penjual, dengan alur yang lebih ringkas dan fokus ke produk inti.",
    excerpt: "Top up game yang baik harus fokus pada ID akun, pilihan item, dan pembayaran tanpa gangguan elemen yang tidak relevan.",
    category: "Game",
    readTime: "6 menit",
    keywords: ["top up game", "top up mobile legends", "top up free fire", "top up pubg"],
    hero: "Kategori game perlu serba ringkas karena pembeli biasanya ingin transaksi secepat mungkin.",
    products: ["Mobile Legends", "Free Fire", "PUBG", "Top up game lainnya"],
    ctaLabel: "Buka Kategori Game",
    ctaHref: "/game",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa top up game perlu alur yang sederhana",
        body: [
          "Pemain tidak ingin terganggu popup atau ringkasan yang tidak relevan saat memilih produk game. Mereka hanya ingin memastikan ID, nominal item, dan pembayaran.",
          "Karena itu, struktur checkout game yang bersih sangat penting untuk kenyamanan dan konversi.",
        ],
      },
      {
        heading: "Cara top up game di Isiloka",
        body: [
          "Masuk ke kategori game, pilih judul game, isi data akun yang diperlukan, lalu pilih nominal item atau diamond. Setelah itu cek pembayaran dan lanjutkan transaksi.",
          "Untuk penjual, validasi ID pelanggan adalah langkah paling penting sebelum transaksi dikirim.",
        ],
      },
      {
        heading: "Peluang bisnis top up game",
        body: [
          "Kategori game sangat menarik karena pembeli biasanya ingin transaksi cepat dan sering melakukan pembelian ulang. Ini menjadikannya produk yang bagus untuk member yang menargetkan pasar anak muda.",
          "Top up game juga mudah dipadukan dengan e-wallet dan paket data agar layanan terasa lebih lengkap.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah kategori game cocok untuk reseller?", answer: "Ya. Mobile Legends, Free Fire, dan PUBG punya pasar aktif dan repeat order tinggi jika alur transaksi cepat dan jelas." },
      { question: "Apa yang paling penting saat top up game?", answer: "Pastikan ID akun benar, pilih judul game yang tepat seperti Mobile Legends, Free Fire, atau PUBG, lalu cocokkan nominal sebelum bayar." },
    ],
  },
  {
    slug: "cara-beli-token-listrik-pln-online",
    title: "Cara Beli Token Listrik PLN Online dengan Pilihan Produk Lebih Jelas",
    description: "Panduan membeli token listrik PLN melalui Isiloka, termasuk cara membaca grup promo, standar reply, dan full reply.",
    excerpt: "Token listrik lebih mudah dicari jika produk dipisahkan per grup dan nominal populer tampil jelas di satu halaman.",
    category: "Listrik",
    readTime: "6 menit",
    keywords: ["beli token listrik", "token listrik pln", "isi token listrik online", "cara beli listrik online"],
    hero: "Kategori token listrik sangat cocok untuk pelanggan rumah tangga dan agen PPOB karena kebutuhannya stabil.",
    products: ["Token Listrik Promo", "Token Listrik Standar Reply", "Token Listrik Full Reply"],
    ctaLabel: "Beli Token Listrik",
    ctaHref: "/listrik/token",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Peran grup produk pada token listrik",
        body: [
          "Token listrik akan lebih mudah dipilih jika produk dipisah per grup seperti promo, standar reply, dan full reply. Pembeli tidak perlu membaca daftar panjang yang bercampur.",
          "Ini membantu pengalaman pengguna sekaligus membuat pilihan produk lebih mudah dipahami pembeli baru.",
        ],
      },
      {
        heading: "Langkah membeli token listrik",
        body: [
          "Masukkan nomor meter, pilih grup produk yang ingin dilihat, lalu pilih nominal token seperti 20.000, 50.000, 100.000, hingga nominal besar lain yang tersedia.",
          "Selalu cek ulang nomor meter sebelum melanjutkan pembayaran agar transaksi tidak salah arah.",
        ],
      },
      {
        heading: "Kenapa token listrik penting untuk member",
        body: [
          "Token listrik adalah produk rumah tangga yang dibutuhkan terus-menerus. Karena itu, kategori ini sangat baik untuk membangun repeat order yang stabil.",
          "Untuk agen, token listrik menjadi pelengkap yang kuat di samping pulsa, paket data, dan layanan PPOB lain.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah token listrik sekarang sudah punya tab grup produk?", answer: "Ya. Grup seperti promo, standar reply, dan full reply membantu pembeli menemukan produk lebih cepat." },
      { question: "Apakah token listrik bagus untuk dijual ulang?", answer: "Ya. Kebutuhan rumah tangga yang rutin membuat kategori ini sangat stabil untuk penjualan." },
    ],
  },
  {
    slug: "cara-bayar-tagihan-listrik-online",
    title: "Cara Bayar Tagihan Listrik Online dengan Alur yang Lebih Praktis",
    description: "Panduan membayar tagihan listrik atau PLN pascabayar melalui Isiloka dengan proses yang mudah dipahami pelanggan rumah tangga maupun agen PPOB.",
    excerpt: "Tagihan listrik termasuk produk PPOB bulanan yang sangat kuat untuk repeat order dan sangat penting untuk layanan rumah tangga.",
    category: "Listrik",
    readTime: "5 menit",
    keywords: ["bayar tagihan listrik", "pln pascabayar online", "bayar listrik online", "ppob listrik"],
    hero: "Pelanggan rumah tangga selalu membutuhkan kanal pembayaran listrik yang cepat, jelas, dan bisa diakses dari HP.",
    products: ["Tagihan Listrik", "Token Listrik Promo"],
    ctaLabel: "Bayar Tagihan Listrik",
    ctaHref: "/listrik/tagihan",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa tagihan listrik penting untuk layanan member",
        body: [
          "Tagihan listrik punya pola pembayaran bulanan yang stabil. Ini menjadikannya kategori yang sangat baik untuk repeat order dan loyalitas pelanggan.",
          "Ketika digabung dengan token listrik, kategori listrik menjadi layanan rumah tangga yang lengkap dan berguna untuk member.",
        ],
      },
      {
        heading: "Cara bayar tagihan listrik",
        body: [
          "Masuk ke halaman tagihan listrik, isi ID pelanggan atau nomor meter, lalu cek nominal tagihan yang muncul. Setelah data sesuai, lanjutkan pembayaran dan simpan bukti transaksi.",
          "Agen sebaiknya mencatat pelanggan yang rutin membayar agar pelayanan bulan berikutnya lebih cepat.",
        ],
      },
      {
        heading: "Peluang repeat order untuk agen",
        body: [
          "Kategori ini sangat cocok dipadukan dengan token listrik, BPJS, PDAM, dan internet pascabayar agar agen punya layanan rumah tangga yang lebih lengkap.",
          "Semakin lengkap layanannya, semakin besar peluang pelanggan kembali tiap bulan ke member yang sama.",
        ],
      },
    ],
    faqs: [
      { question: "Apa beda token listrik dan tagihan listrik?", answer: "Token listrik dipakai pelanggan prabayar yang membeli nominal listrik, sedangkan tagihan listrik untuk pelanggan pascabayar yang membayar pemakaian bulanan." },
      { question: "Layanan rumah tangga apa yang paling cocok dipasangkan dengan tagihan listrik?", answer: "Token listrik, BPJS Kesehatan, PDAM, dan internet pascabayar adalah pasangan paling masuk akal untuk membangun repeat order bulanan." },
    ],
  },
  {
    slug: "cara-bayar-bpjs-kesehatan-online",
    title: "Cara Bayar BPJS Kesehatan Online dengan Mudah",
    description: "Panduan membayar BPJS Kesehatan online lewat Isiloka untuk kebutuhan pribadi, keluarga, dan layanan PPOB.",
    excerpt: "BPJS Kesehatan adalah produk PPOB penting untuk kebutuhan keluarga dan sangat baik untuk layanan rutin bulanan.",
    category: "BPJS",
    readTime: "5 menit",
    keywords: ["bayar bpjs online", "bpjs kesehatan online", "cara bayar bpjs", "ppob bpjs"],
    hero: "Kategori BPJS memperluas layanan Isiloka dari produk harian ke kebutuhan bulanan yang lebih wajib.",
    products: ["BPJS Kesehatan", "Tagihan Listrik", "PDAM"],
    ctaLabel: "Bayar BPJS Kesehatan",
    ctaHref: "/bpjs/kesehatan",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Mengapa BPJS penting untuk member",
        body: [
          "BPJS dibayar rutin oleh banyak keluarga. Karena itu, kategori ini sangat bagus untuk member yang ingin punya pelanggan bulanan dengan kebutuhan yang jelas.",
          "Layanan BPJS juga memperluas citra Isiloka dari top up harian menjadi platform pembayaran yang lebih lengkap.",
        ],
      },
      {
        heading: "Cara bayar BPJS di Isiloka",
        body: [
          "Masuk ke kategori BPJS, isi nomor peserta atau data yang dibutuhkan, lalu cek nominal tagihan. Setelah datanya benar, lanjutkan pembayaran.",
          "Biasakan menyimpan data peserta keluarga atau pelanggan agar pembayaran berikutnya lebih cepat.",
        ],
      },
      {
        heading: "Nilai bisnis untuk agen",
        body: [
          "BPJS memberi agen akses ke pelanggan yang rutin bertransaksi bulanan. Ini bagus untuk stabilitas usaha dan loyalitas pelanggan.",
          "Selain itu, kategori ini melengkapi layanan rumah tangga bersama listrik, PDAM, dan internet pascabayar.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah BPJS cocok untuk dijual di website PPOB?", answer: "Ya. BPJS Kesehatan adalah layanan rutin dengan intent transaksi tinggi dan sangat relevan untuk website produk digital." },
      { question: "Layanan apa yang paling cocok dipasangkan dengan BPJS?", answer: "Tagihan listrik, PDAM, dan internet pascabayar cocok dipasangkan dengan BPJS untuk membangun paket layanan rumah tangga yang lengkap." },
    ],
  },
  {
    slug: "cara-bayar-pdam-online",
    title: "Cara Bayar PDAM Online untuk Pelanggan Rumah Tangga dan Agen PPOB",
    description: "Panduan membayar tagihan PDAM secara online melalui Isiloka untuk pelanggan rumah tangga dan usaha PPOB.",
    excerpt: "PDAM adalah produk rumah tangga yang stabil dan sangat cocok menjadi kategori pendamping untuk usaha pulsa dan PPOB.",
    category: "PDAM",
    readTime: "5 menit",
    keywords: ["bayar pdam online", "tagihan pdam online", "ppob pdam", "cara bayar air online"],
    hero: "Layanan PDAM membantu Isiloka menjangkau kebutuhan rumah tangga yang rutin dan sangat dekat dengan repeat order.",
    products: ["PDAM", "Tagihan Listrik", "BPJS Kesehatan"],
    ctaLabel: "Bayar PDAM",
    ctaHref: "/pdam",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa PDAM penting untuk layanan digital",
        body: [
          "Tagihan air dibayar rutin oleh banyak rumah tangga. Karena itu, kategori PDAM sangat bagus untuk melengkapi website produk digital dan menarik pelanggan bulanan.",
          "PDAM juga membantu toko Anda terlihat lebih lengkap dibanding hanya menjual pulsa dan kuota.",
        ],
      },
      {
        heading: "Cara pembayaran PDAM",
        body: [
          "Buka kategori PDAM, masukkan nomor pelanggan, lalu cek detail tagihan yang muncul. Setelah data sesuai, lanjutkan ke pembayaran dan simpan bukti transaksi.",
          "Jika Anda mengelola banyak pelanggan, pencatatan nomor pelanggan yang rapi akan sangat membantu operasional.",
        ],
      },
      {
        heading: "Kenapa PDAM bagus untuk usaha PPOB",
        body: [
          "PDAM memperkuat layanan rumah tangga yang dibayar rutin setiap bulan. Ini sangat membantu member yang ingin punya basis pelanggan tetap.",
          "Jika dipadukan dengan listrik, BPJS, dan internet pascabayar, layanan toko digital akan terasa jauh lebih lengkap.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah PDAM bagus untuk repeat order?", answer: "Ya. Tagihan air bersifat rutin sehingga pelanggan punya alasan kuat untuk kembali setiap bulan." },
      { question: "Layanan apa yang paling cocok dipasangkan dengan PDAM?", answer: "Tagihan listrik, BPJS Kesehatan, dan internet pascabayar adalah pasangan yang paling masuk akal untuk memperkuat layanan rumah tangga." },
    ],
  },
  {
    slug: "cara-bayar-internet-pascabayar-online",
    title: "Cara Bayar Internet Pascabayar Online tanpa Repot",
    description: "Panduan membayar internet pascabayar secara online melalui Isiloka untuk kebutuhan pribadi, keluarga, dan bisnis PPOB.",
    excerpt: "Internet pascabayar adalah kategori rumah tangga modern yang penting untuk melengkapi layanan pembayaran bulanan.",
    category: "Internet Pascabayar",
    readTime: "5 menit",
    keywords: ["bayar internet pascabayar", "tagihan internet online", "ppob internet rumah", "internet pascabayar"],
    hero: "Kategori internet pascabayar membantu Isiloka menjangkau kebutuhan digital rumah tangga yang terus tumbuh.",
    products: ["Internet Pascabayar", "Tagihan Listrik", "TV Berlangganan"],
    ctaLabel: "Bayar Internet Pascabayar",
    ctaHref: "/internet-pascabayar",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Mengapa internet pascabayar penting untuk layanan member",
        body: [
          "Pengguna internet rumah membayar tagihan rutin setiap bulan. Ini membuat kategori internet pascabayar sangat baik untuk repeat order dan layanan pelanggan jangka panjang.",
          "Kategori ini juga memperkuat citra Isiloka sebagai platform kebutuhan rumah tangga digital, bukan hanya top up harian.",
        ],
      },
      {
        heading: "Cara pembayaran internet pascabayar",
        body: [
          "Masuk ke kategori internet pascabayar, isi nomor pelanggan yang diminta, lalu cek tagihan. Setelah hasilnya benar, lanjutkan pembayaran dan simpan referensi transaksi.",
          "Bagi agen, layanan ini bagus untuk memperluas repeat order dari segmen rumah tangga atau usaha kecil.",
        ],
      },
      {
        heading: "Potensi bisnis untuk agen",
        body: [
          "Internet pascabayar sangat cocok dipadukan dengan listrik, PDAM, BPJS, dan TV berlangganan agar layanan rumah tangga member lebih lengkap.",
          "Untuk bisnis, kategori ini membantu menambah nilai layanan dan peluang loyalitas pelanggan.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah kategori internet pascabayar cocok untuk agen PPOB?", answer: "Ya. Ini adalah kebutuhan bulanan yang sangat relevan untuk rumah tangga dan usaha kecil." },
      { question: "Layanan apa yang paling cocok dipasangkan dengan internet pascabayar?", answer: "Tagihan listrik dan TV berlangganan paling cocok dipasangkan karena sama-sama mewakili kebutuhan rutin rumah tangga modern." },
    ],
  },
  {
    slug: "cara-bayar-hp-pascabayar-online",
    title: "Cara Bayar HP Pascabayar Online untuk Kebutuhan Pribadi dan Bisnis",
    description: "Panduan membayar tagihan HP pascabayar melalui Isiloka untuk pelanggan pribadi, agen, dan website PPOB.",
    excerpt: "HP pascabayar adalah kategori niche yang tetap penting karena membantu layanan terlihat lebih lengkap dan profesional.",
    category: "HP Pascabayar",
    readTime: "5 menit",
    keywords: ["bayar hp pascabayar", "tagihan pascabayar online", "kartu halo online", "ppob hp pascabayar"],
    hero: "Kategori niche seperti HP pascabayar tetap penting untuk membangun layanan digital yang lengkap.",
    products: ["HP Pascabayar", "Pulsa Telkomsel", "Paket Telepon Telkomsel"],
    ctaLabel: "Bayar HP Pascabayar",
    ctaHref: "/hp-pascabayar",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa kategori ini masih dibutuhkan",
        body: [
          "Sebagian pelanggan tetap memakai layanan pascabayar karena kebutuhan kerja, prioritas jaringan, atau kebiasaan penggunaan. Mereka membutuhkan kanal pembayaran yang cepat dan mudah diakses.",
          "Bagi website produk digital, kategori ini memberi sinyal bahwa layanan Anda lengkap dan tidak hanya fokus pada produk populer.",
        ],
      },
      {
        heading: "Cara bayar HP pascabayar",
        body: [
          "Masuk ke kategori HP pascabayar, isi nomor pelanggan, lalu cek tagihan yang muncul. Setelah datanya benar, lanjutkan pembayaran dan simpan bukti transaksi.",
          "Ketelitian pada nomor pelanggan sangat penting karena produk pascabayar sangat bergantung pada data tujuan yang benar.",
        ],
      },
      {
        heading: "Nilai untuk bisnis member",
        body: [
          "Kategori niche membantu layanan member terlihat lebih lengkap karena tidak hanya fokus pada produk paling umum.",
          "Di sisi bisnis, kategori seperti ini membantu membedakan layanan Anda dari kompetitor yang hanya bermain di produk populer.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah kategori niche masih penting?", answer: "Ya. Kategori niche membantu layanan terlihat lebih lengkap dan menjangkau segmen pelanggan yang lebih spesifik." },
      { question: "Produk apa yang paling relevan dipasangkan dengan HP pascabayar?", answer: "Pulsa Telkomsel dan paket telepon Telkomsel biasanya paling relevan sebagai pelengkap untuk pelanggan yang masih aktif memakai layanan pascabayar." },
    ],
  },
  {
    slug: "cara-beli-voucher-tv-digital-dan-berlangganan",
    title: "Cara Beli Voucher TV Digital dan Bayar TV Berlangganan Online",
    description: "Panduan membeli voucher TV digital dan membayar layanan TV berlangganan melalui Isiloka untuk pelanggan rumah tangga dan agen PPOB.",
    excerpt: "TV digital dan TV berlangganan adalah kategori rumah tangga yang bagus untuk melengkapi layanan digital keluarga.",
    category: "TV",
    readTime: "5 menit",
    keywords: ["voucher tv digital", "bayar tv berlangganan online", "ppob tv", "tv digital online"],
    hero: "Kategori TV membuat layanan rumah tangga di Isiloka terasa lebih lengkap dan lebih relevan untuk repeat order.",
    products: ["TV Berlangganan", "Internet Pascabayar", "Tagihan Listrik"],
    ctaLabel: "Buka Kategori TV",
    ctaHref: "/tv",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Mengapa kategori TV layak dibahas",
        body: [
          "Banyak pelanggan rumah tangga membayar layanan hiburan secara rutin. Karena itu, kategori TV layak menjadi bagian dari layanan PPOB yang lengkap.",
          "Topik ini juga memperluas layanan rumah tangga di luar listrik dan air.",
        ],
      },
      {
        heading: "Cara bayar atau beli voucher TV",
        body: [
          "Masuk ke kategori TV, pilih operator atau produk yang tersedia, lalu isi nomor pelanggan atau data yang diminta. Setelah hasil tagihan atau voucher tampil, lanjutkan pembayaran.",
          "Simpan data pelanggan yang sering membayar agar layanan bulan berikutnya bisa diproses lebih cepat.",
        ],
      },
      {
        heading: "Hubungan TV dan layanan rumah tangga",
        body: [
          "Kategori TV membantu layanan member terlihat lebih lengkap untuk kebutuhan keluarga di rumah.",
          "Pelanggan yang datang untuk TV sering juga membutuhkan listrik, BPJS, atau internet pascabayar.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah kategori TV cocok untuk toko PPOB?", answer: "Ya. TV berlangganan dan voucher TV memberi variasi layanan yang relevan untuk rumah tangga." },
      { question: "Layanan apa yang paling cocok dipasangkan dengan TV berlangganan?", answer: "Internet pascabayar dan tagihan listrik paling cocok dipasangkan karena ketiganya sama-sama dekat dengan kebutuhan hiburan dan operasional rumah." },
    ],
  },
  {
    slug: "cara-perpanjang-masa-aktif-nomor-hp",
    title: "Cara Perpanjang Masa Aktif Nomor HP dengan Cepat",
    description: "Panduan memperpanjang masa aktif nomor HP melalui Isiloka agar nomor tetap aktif tanpa harus membeli produk yang tidak diperlukan.",
    excerpt: "Masa aktif sering diabaikan sampai nomor hampir hangus. Karena itu, topik ini sangat relevan untuk pencarian yang butuh solusi cepat.",
    category: "Masa Aktif",
    readTime: "4 menit",
    keywords: ["perpanjang masa aktif", "masa aktif nomor hp", "beli masa aktif", "masa aktif telkomsel"],
    hero: "Kategori masa aktif membantu pengguna menjaga nomor tetap hidup dengan langkah yang lebih jelas.",
    products: ["Masa Aktif"],
    ctaLabel: "Perpanjang Masa Aktif",
    ctaHref: "/masa-aktif",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa masa aktif penting",
        body: [
          "Nomor cadangan, nomor usaha, atau nomor yang dipakai untuk OTP harus tetap aktif. Begitu masa aktif habis, dampaknya bisa lebih besar daripada yang terlihat.",
          "Karena masalahnya nyata dan mendesak, kategori masa aktif sangat berguna untuk membantu pelanggan yang butuh solusi cepat.",
        ],
      },
      {
        heading: "Cara beli masa aktif",
        body: [
          "Masuk ke kategori masa aktif, pilih operator atau brand yang sesuai, lalu pilih produk yang tersedia. Setelah itu lanjutkan pembayaran seperti transaksi digital lain.",
          "Flow ini sederhana, tetapi sangat berguna untuk pengguna yang hanya ingin menjaga nomor tetap hidup.",
        ],
      },
      {
        heading: "Nilai bisnis untuk agen",
        body: [
          "Masa aktif cocok sebagai produk pendamping untuk pulsa dan paket data. Pelanggan yang datang untuk satu kebutuhan sering juga membutuhkan kategori ini.",
          "Karena relatif niche, topik ini juga membantu website menangkap keyword yang sering diabaikan pesaing.",
        ],
      },
    ],
    faqs: [
      { question: "Siapa yang paling butuh layanan masa aktif?", answer: "Pengguna nomor cadangan, nomor bisnis, modem, atau nomor yang dipakai untuk OTP dan komunikasi penting." },
      { question: "Apakah masa aktif cocok dijual bersama pulsa?", answer: "Ya. Masa aktif sangat cocok sebagai produk pendamping untuk pengguna operator seluler." },
    ],
  },
  {
    slug: "cara-aktivasi-perdana-dan-topup-awal",
    title: "Cara Aktivasi Perdana dan Top Up Awal untuk Nomor Baru",
    description: "Panduan aktivasi perdana dan top up awal nomor baru lewat Isiloka untuk pengguna baru, agen, dan penjual starter pack.",
    excerpt: "Aktivasi perdana adalah pintu masuk yang bagus untuk memperkenalkan pelanggan baru ke kategori pulsa, kuota, dan masa aktif.",
    category: "Aktivasi Perdana",
    readTime: "5 menit",
    keywords: ["aktivasi perdana", "starter pack internet", "nomor baru", "perdana data"],
    hero: "Nomor baru membutuhkan alur awal yang jelas agar pengguna cepat sampai ke kebutuhan transaksi berikutnya.",
    products: ["Aktivasi Perdana", "Pulsa", "Paket Data", "Masa Aktif"],
    ctaLabel: "Lihat Aktivasi Perdana",
    ctaHref: "/aktivasi-perdana",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Mengapa aktivasi perdana relevan",
        body: [
          "Pengguna nomor baru biasanya membutuhkan panduan yang sangat jelas. Mereka belum tentu tahu perbedaan antara aktivasi kartu, isi pulsa, beli kuota, dan perpanjangan masa aktif.",
          "Karena itu, layanan aktivasi perdana sangat berguna untuk membantu pengguna baru memahami langkah awal yang benar sebelum masuk ke kebutuhan berikutnya.",
        ],
      },
      {
        heading: "Langkah awal setelah kartu aktif",
        body: [
          "Setelah kartu aktif, pengguna biasanya butuh pulsa awal, paket data pertama, atau masa aktif tambahan. Tiga kebutuhan ini adalah langkah lanjutan yang paling sering dibutuhkan pengguna baru.",
          "Untuk agen, starter pack juga bisa menjadi pintu masuk ke repeat order berikutnya.",
        ],
      },
      {
        heading: "Peluang penjualan lanjutan",
        body: [
          "Topik nomor baru, aktivasi perdana, dan paket awal sangat dekat dengan kebutuhan pengguna baru atau penjual kartu perdana.",
          "Dari aktivasi perdana, pelanggan bisa diarahkan ke pulsa, paket data, dan masa aktif sebagai kebutuhan lanjutan yang paling dekat.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah aktivasi perdana relevan untuk website jualan pulsa?", answer: "Ya. Aktivasi perdana membantu menangkap pengguna baru dan mengarahkannya ke produk lanjutan." },
      { question: "Kategori apa yang paling dekat dengan kebutuhan pengguna baru?", answer: "Pulsa, paket data, dan masa aktif adalah tiga kategori paling dekat dengan kebutuhan pengguna baru." },
    ],
  },
  {
    slug: "cara-jualan-pulsa-dari-rumah-untuk-pemula",
    title: "Cara Jualan Pulsa dari Rumah untuk Pemula",
    description: "Panduan memulai usaha jualan pulsa dari rumah dengan produk awal yang tepat, langkah sederhana, dan arah tumbuh bersama Isiloka.",
    excerpt: "Jualan pulsa dari rumah tetap menarik karena modalnya fleksibel, pasarnya luas, dan bisa berkembang menjadi usaha produk digital yang lengkap.",
    category: "Bisnis Agen",
    readTime: "7 menit",
    keywords: ["cara jualan pulsa", "usaha pulsa rumahan", "bisnis pulsa pemula", "jualan produk digital"],
    hero: "Bisnis pulsa rumahan paling sehat jika dibangun bertahap dengan produk, pelayanan, dan pencatatan yang rapi.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "DANA", "Token Listrik Promo"],
    ctaLabel: "Mulai dari Pulsa",
    ctaHref: "/pulsa",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa usaha pulsa masih relevan",
        body: [
          "Pulsa, paket data, e-wallet, dan token listrik tetap dibutuhkan setiap hari. Ini membuat usaha produk digital masih sangat relevan meskipun cara jualannya bergeser ke online.",
          "Untuk pemula, bisnis ini menarik karena bisa dimulai dari rumah tanpa harus langsung membuka toko fisik.",
        ],
      },
      {
        heading: "Produk pertama yang paling aman dijual",
        body: [
          "Mulailah dari pulsa semua operator, paket data populer, dan e-wallet. Setelah alur dasar stabil, barulah tambahkan token listrik dan PPOB lain.",
          "Pendekatan bertahap akan membuat operasional lebih mudah dikontrol dan kesalahan lebih sedikit.",
        ],
      },
      {
        heading: "Arah berkembang setelah tahap awal",
        body: [
          "Setelah pelanggan mulai rutin, usaha bisa diarahkan ke model agen yang lebih lengkap, website sendiri, atau H2H jika volume transaksi sudah lebih besar.",
          "Dengan memahami tahapan ini sejak awal, pemula bisa menyiapkan usaha yang lebih rapi dan tidak bingung saat transaksi mulai bertambah.",
        ],
      },
    ],
    faqs: [
      { question: "Produk apa yang paling aman untuk jualan pertama?", answer: "Mulailah dari pulsa operator besar, paket data yang paling sering dicari, top up DANA atau GoPay, lalu token listrik sebagai layanan rumah tangga pertama." },
      { question: "Apakah harus punya toko fisik untuk mulai jualan?", answer: "Tidak. Banyak member memulai dari rumah lewat WhatsApp, lalu baru menambah katalog online atau website saat transaksi sudah lebih rutin." },
    ],
  },
  {
    slug: "cara-menjadi-agen-pulsa-di-pulsakilat",
    title: "Cara Menjadi Agen Pulsa di Isiloka",
    description: "Panduan menjadi agen Isiloka dari pendaftaran, pilihan produk awal, sampai cara membangun layanan yang rapi dan repeat order.",
    excerpt: "Menjadi agen pulsa bukan hanya soal saldo. Yang lebih penting adalah memahami produk, alur transaksi, dan pengalaman pelanggan.",
    category: "Bisnis Agen",
    readTime: "7 menit",
    keywords: ["cara menjadi agen pulsa", "daftar agen pulsa", "agen pulsakilat", "reseller produk digital"],
    hero: "Agen yang tumbuh cepat biasanya bukan yang paling murah, tetapi yang paling jelas dan paling konsisten melayani pelanggan.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "DANA", "Token Listrik Promo", "Mobile Legends"],
    ctaLabel: "Daftar Jadi Member",
    ctaHref: "/register",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Modal utama menjadi agen",
        body: [
          "Modal utama agen adalah kombinasi saldo, pengetahuan produk, dan disiplin operasional. Tanpa tiga hal ini, bisnis mudah kacau saat order mulai ramai.",
          "Isiloka memberi fondasi yang baik karena kategori produk utama sudah tersedia dalam satu platform.",
        ],
      },
      {
        heading: "Langkah memulai sebagai agen",
        body: [
          "Buat akun, pahami kategori produk, lalu pilih produk inti yang paling mudah dijual. Untuk tahap awal, pulsa, paket data, dan e-wallet adalah kombinasi yang paling aman.",
          "Setelah itu, tambahkan layanan rumah tangga dan game agar pelanggan punya alasan lebih kuat untuk repeat order.",
        ],
      },
      {
        heading: "Jalur naik level",
        body: [
          "Agen yang sudah rutin bertransaksi bisa berkembang ke katalog online, website jualan sendiri, atau H2H. Ini membantu usaha tidak berhenti di level manual.",
          "Dengan jalur seperti ini, agen tidak perlu pindah platform saat bisnis mulai tumbuh. Isiloka tetap bisa dipakai dari tahap awal sampai tahap usaha yang lebih serius.",
        ],
      },
    ],
    faqs: [
      { question: "Produk apa yang paling cocok untuk agen baru?", answer: "Pulsa semua operator, paket data populer, top up DANA, dan token listrik adalah kombinasi paling aman untuk tahap awal karena cepat dipahami pelanggan." },
      { question: "Kapan agen perlu menambah website atau H2H?", answer: "Saat order mulai rutin dari banyak pelanggan, katalog makin luas, dan proses manual mulai menyita waktu atau rawan salah input." },
    ],
  },
  {
    slug: "cara-menjadi-member-h2h-pulsakilat",
    title: "Cara Menjadi Member H2H Isiloka untuk Bisnis yang Lebih Besar",
    description: "Panduan menjadi member H2H Isiloka untuk website, aplikasi, atau panel reseller yang ingin transaksi lebih otomatis dan rapi.",
    excerpt: "Model H2H cocok untuk bisnis yang ingin mengotomatisasi transaksi, membangun brand sendiri, dan naik ke level operasional yang lebih efisien.",
    category: "H2H",
    readTime: "8 menit",
    keywords: ["member h2h", "host to host pulsa", "api pulsa", "pulsakilat h2h"],
    hero: "H2H adalah langkah logis ketika bisnis digital tidak lagi nyaman dijalankan sepenuhnya secara manual.",
    products: ["API H2H", "Pulsa Telkomsel", "Paket Data Telkomsel", "BPJS Kesehatan", "Mobile Legends", "DANA"],
    ctaLabel: "Buka Dokumentasi H2H",
    ctaHref: "/docs",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Siapa yang cocok memakai H2H",
        body: [
          "H2H cocok untuk pemilik website, aplikasi, panel reseller, atau pelaku bisnis yang sudah punya volume transaksi cukup besar. Model ini membantu operasional jadi lebih otomatis.",
          "Jika order masih sedikit, retail biasa mungkin cukup. Tetapi saat transaksi mulai padat, H2H memberi efisiensi yang jauh lebih tinggi.",
        ],
      },
      {
        heading: "Keuntungan dibanding proses manual",
        body: [
          "Dengan H2H, Anda bisa menerima order dari banyak kanal lalu meneruskannya ke sistem secara otomatis. Ini mengurangi salah input dan mempercepat operasional.",
          "Selain itu, pelanggan berinteraksi dengan brand Anda sendiri, bukan hanya dengan chat manual.",
        ],
      },
      {
        heading: "Kapan bisnis siap naik level",
        body: [
          "Bisnis siap naik ke H2H ketika volume transaksi mulai menyulitkan proses manual, ketika laporan dan log mulai penting, dan ketika Anda ingin membangun sistem sendiri yang lebih serius.",
          "Isiloka memberi jalur ini untuk member yang ingin bergerak dari layanan manual ke sistem yang lebih otomatis tanpa harus mengganti basis produk.",
        ],
      },
    ],
    faqs: [
      { question: "Apa bedanya agen biasa dengan member H2H?", answer: "Agen biasa biasanya menerima order lewat chat lalu memproses manual, sedangkan H2H menghubungkan website, aplikasi, atau panel reseller langsung ke sistem transaksi." },
      { question: "Produk apa yang cocok dibawa ke H2H lebih dulu?", answer: "Pulsa, paket data, e-wallet, game, dan PPOB dasar seperti BPJS adalah kategori yang paling masuk akal untuk mulai diotomatisasi." },
    ],
  },
  {
    slug: "cara-membuat-website-jualan-pulsa-sendiri",
    title: "Cara Membuat Website Jualan Pulsa Sendiri yang Siap Tumbuh",
    description: "Panduan membuat website jualan pulsa dengan halaman produk yang jelas, struktur layanan yang rapi, dan arah tumbuh ke H2H Isiloka.",
    excerpt: "Website jualan pulsa yang kuat harus punya halaman produk, informasi layanan, dan arah pertumbuhan yang jelas ke integrasi yang lebih serius.",
    category: "Website Jualan",
    readTime: "9 menit",
    keywords: ["website jualan pulsa", "cara buat website pulsa", "jualan pulsa online", "website ppob"],
    hero: "Website produk digital terbaik bukan hanya katalog, tetapi mesin trafik dan transaksi yang saling terhubung.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "DANA", "Token Listrik Promo", "Mobile Legends", "API H2H"],
    ctaLabel: "Daftar dan Siapkan Website Jualan",
    ctaHref: "/register",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Halaman wajib yang harus ada",
        body: [
          "Website jualan pulsa setidaknya harus punya halaman kategori utama, halaman panduan, halaman tentang, FAQ, dan kontak. Ini membantu calon pelanggan memahami struktur layanan Anda.",
          "Tanpa struktur ini, website hanya terlihat seperti daftar produk tanpa konteks yang kuat.",
        ],
      },
      {
        heading: "Peran halaman panduan",
        body: [
          "Halaman panduan membantu website menjelaskan cara isi pulsa, cara bayar listrik, atau cara menjadi agen dengan bahasa yang mudah dipahami calon pelanggan dan calon member.",
          "Dari halaman inilah pengunjung diarahkan ke kategori produk dan layanan yang paling relevan.",
        ],
      },
      {
        heading: "Kapan perlu H2H",
        body: [
          "Saat website mulai ramai dan transaksi tidak nyaman lagi diproses manual, H2H menjadi tahap berikutnya. Pada fase ini, bisnis mulai membutuhkan integrasi yang lebih rapi agar order bisa diproses otomatis.",
          "Dengan model seperti ini, website jualan tidak berhenti sebagai katalog, tetapi bisa berkembang menjadi mesin transaksi yang lebih serius.",
        ],
      },
    ],
    faqs: [
      { question: "Halaman produk apa yang paling penting untuk website awal?", answer: "Mulailah dari pulsa operator utama, paket data, e-wallet, token listrik, dan satu kategori game agar website langsung terlihat hidup dan mudah dipakai." },
      { question: "Kapan website perlu H2H?", answer: "Saat order mulai ramai, pembayaran perlu dipantau lebih rapi, dan Anda ingin order dari website atau panel masuk otomatis ke sistem transaksi." },
    ],
  },
  {
    slug: "keuntungan-berlangganan-dan-bertumbuh-bersama-pulsakilat",
    title: "Keuntungan Berlangganan dan Bertumbuh Bersama Isiloka",
    description: "Keuntungan memakai Isiloka untuk agen, reseller, dan bisnis yang ingin tumbuh dari transaksi harian ke website atau H2H.",
    excerpt: "Keuntungan platform yang baik tidak hanya pada produk, tetapi pada seberapa mudah platform itu membantu bisnis bertumbuh.",
    category: "Bisnis Agen",
    readTime: "6 menit",
    keywords: ["keuntungan pulsakilat", "platform jualan pulsa", "reseller produk digital", "agen pulsakilat"],
    hero: "Isiloka relevan bukan hanya untuk transaksi harian, tetapi juga untuk pertumbuhan bisnis jangka panjang.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "Token Listrik Promo", "DANA", "Mobile Legends", "API H2H"],
    ctaLabel: "Lihat Produk untuk Calon Member",
    ctaHref: "/pulsa",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Produk yang saling melengkapi",
        body: [
          "Isiloka tidak berhenti di satu kategori. Pulsa, paket data, e-wallet, token listrik, game, dan PPOB saling mendukung untuk membentuk layanan yang lebih lengkap.",
          "Kelengkapan seperti ini membantu agen dan pemilik website membangun repeat order yang lebih stabil.",
        ],
      },
      {
        heading: "Ruang untuk naik level",
        body: [
          "Seorang pengguna bisa mulai dari retail, lalu berkembang ke agen, lalu ke website sendiri, bahkan ke H2H. Jalur bertumbuh ini membuat platform lebih relevan untuk jangka panjang.",
          "Artinya, calon member tidak perlu mencari sistem baru setiap kali bisnis masuk ke tahap yang lebih besar.",
        ],
      },
      {
        heading: "Kenapa jalur ini penting untuk calon member",
        body: [
          "Calon member biasanya ingin tahu apakah sebuah platform hanya cocok untuk transaksi kecil atau juga bisa dipakai untuk bertumbuh. Isiloka memberi jawaban lewat jalur yang jelas dari transaksi harian sampai integrasi yang lebih serius.",
          "Dengan memahami jalur retail, agen, website sendiri, dan H2H, calon member bisa memilih langkah yang paling sesuai dengan tahap usahanya.",
        ],
      },
    ],
    faqs: [
      { question: "Apa keuntungan utama Isiloka untuk agen?", answer: "Agen bisa menjual pulsa, paket data, e-wallet, token listrik, game, dan layanan rumah tangga tanpa pindah-pindah platform saat usaha mulai tumbuh." },
      { question: "Apakah Isiloka cocok untuk pemula dan bisnis bertumbuh?", answer: "Ya. Pemula bisa mulai dari pulsa, kuota, dan e-wallet, sementara bisnis yang lebih siap bisa naik ke website jualan dan H2H." },
    ],
  },
  {
    slug: "cara-mengelola-markup-dan-profit-produk-digital",
    title: "Cara Mengelola Markup dan Profit Produk Digital",
    description: "Panduan mengatur markup dan profit pulsa, paket data, e-wallet, game, dan PPOB agar usaha agen tetap sehat dan kompetitif.",
    excerpt: "Markup yang baik tidak hanya menjaga profit, tetapi juga membantu layanan tetap kompetitif dan dipercaya pelanggan.",
    category: "Bisnis Agen",
    readTime: "7 menit",
    keywords: ["markup pulsa", "profit jualan pulsa", "margin produk digital", "harga jual agen pulsa"],
    hero: "Margin yang sehat lebih penting daripada sekadar menjadi yang paling murah.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "DANA", "Mobile Legends", "Token Listrik Promo"],
    ctaLabel: "Lihat Produk dengan Margin Aman",
    ctaHref: "/pulsa",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Jangan hanya mengejar harga termurah",
        body: [
          "Perang harga tanpa strategi akan membuat bisnis cepat lelah. Pelanggan juga menilai kecepatan, kejelasan, dan kenyamanan, bukan hanya harga terendah.",
          "Markup yang sehat memberi ruang untuk pelayanan yang lebih baik dan usaha yang lebih tahan lama.",
        ],
      },
      {
        heading: "Atur markup per kategori",
        body: [
          "Pulsa dan paket data biasanya bermain di volume tinggi, sedangkan PPOB, game, atau e-wallet bisa punya ruang margin yang berbeda. Karena itu, semua kategori tidak harus diperlakukan sama.",
          "Gunakan data transaksi untuk melihat kategori mana yang paling sering dibeli ulang dan kategori mana yang bisa menjaga profit lebih baik.",
        ],
      },
      {
        heading: "Kenapa margin perlu dihitung sejak awal",
        body: [
          "Markup, margin, dan profit adalah dasar penting bagi agen yang ingin usahanya sehat dalam jangka panjang.",
          "Dengan memahami perbedaan karakter tiap kategori produk, member bisa menyusun harga jual yang lebih aman dan lebih terukur.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah markup kecil selalu lebih baik?", answer: "Tidak. Markup harus disesuaikan dengan volume, persaingan, dan nilai layanan yang Anda berikan." },
      { question: "Kategori apa yang paling baik untuk menjaga profit?", answer: "Biasanya kombinasi pulsa, paket data, e-wallet, token listrik, dan PPOB dasar memberi keseimbangan yang baik antara volume dan repeat order." },
    ],
  },
  {
    slug: "cara-rekonsiliasi-transaksi-produk-digital",
    title: "Cara Rekonsiliasi Transaksi Produk Digital",
    description: "Panduan rekonsiliasi transaksi untuk agen, admin, dan pemilik website agar pembayaran, status, dan catatan tetap sinkron.",
    excerpt: "Saat bisnis transaksi mulai ramai, rekonsiliasi menjadi pembeda antara usaha yang tumbuh sehat dan usaha yang penuh kebingungan.",
    category: "Operasional",
    readTime: "7 menit",
    keywords: ["rekonsiliasi transaksi", "catatan transaksi digital", "operasional agen pulsa", "status transaksi produk digital"],
    hero: "Pencatatan yang rapi menjaga kepercayaan pelanggan dan membantu bisnis bertahan dalam jangka panjang.",
    products: ["Pulsa Telkomsel", "BPJS Kesehatan", "DANA", "Mobile Legends", "Deposit QRIS"],
    ctaLabel: "Masuk untuk Cek Saldo dan Mutasi",
    ctaHref: "/login",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa rekonsiliasi wajib",
        body: [
          "Saat transaksi masih sedikit, banyak kesalahan tertutup karena masih mudah diingat. Namun saat volume naik, status pending, pembayaran ganda, atau catatan yang hilang akan cepat menjadi masalah.",
          "Rekonsiliasi membantu memastikan pembayaran, status, dan referensi transaksi tetap sinkron.",
        ],
      },
      {
        heading: "Data minimum yang harus dicatat",
        body: [
          "Simpan waktu transaksi, nomor tujuan atau akun, produk, nominal, status akhir, dan referensi transaksi. Untuk bisnis lebih besar, log pembayaran dan log status juga sangat penting.",
          "Tanpa data minimum ini, penelusuran masalah akan selalu lambat dan memakan energi.",
        ],
      },
      {
        heading: "Kenapa operasional rapi penting untuk member",
        body: [
          "Saat transaksi makin ramai, agen dan admin tidak cukup hanya mengandalkan ingatan atau chat. Operasional yang rapi menjadi syarat agar usaha tetap sehat dan pelanggan tetap percaya.",
          "Segmen ini membutuhkan panduan yang lebih serius karena mereka sedang membangun sistem kerja jangka panjang, bukan sekadar transaksi sesaat.",
        ],
      },
    ],
    faqs: [
      { question: "Apa manfaat terbesar rekonsiliasi transaksi?", answer: "Mengurangi kebocoran operasional, mempercepat penelusuran masalah, dan menjaga kepercayaan pelanggan saat transaksi mulai ramai." },
      { question: "Data apa yang paling penting saat mulai rekonsiliasi?", answer: "Catat waktu transaksi, produk, nomor tujuan atau akun, nominal, status akhir, dan referensi pembayaran agar penelusuran masalah tidak berputar-putar." },
    ],
  },
  {
    slug: "cara-beli-paket-telepon-semua-operator",
    title: "Cara Beli Paket Telepon Semua Operator di Isiloka",
    description: "Panduan membeli paket telepon semua operator di Isiloka untuk pelanggan harian dan member yang ingin menambah layanan pelengkap.",
    excerpt: "Paket telepon tetap penting untuk pelanggan yang butuh nelpon rutin, terutama untuk keluarga, kerja, dan usaha kecil.",
    category: "Paket Telepon",
    readTime: "5 menit",
    keywords: ["paket telepon", "paket nelpon", "paket telepon telkomsel", "paket telepon murah"],
    hero: "Paket telepon adalah produk pendamping yang kuat karena masih dibutuhkan oleh pelanggan yang aktif berkomunikasi lewat panggilan suara.",
    products: ["Paket Telepon Telkomsel", "Paket Telepon Indosat", "Paket Telepon XL", "Paket Telepon Smartfren"],
    ctaLabel: "Buka Paket Telepon",
    ctaHref: "/paket-telepon",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa paket telepon masih dibutuhkan",
        body: [
          "Tidak semua pelanggan hanya mencari kuota internet. Banyak pengguna masih membutuhkan paket telepon untuk keluarga, pekerjaan, dan komunikasi dengan pelanggan.",
          "Karena itu, kategori paket telepon tetap penting untuk membuat layanan Isiloka terasa lebih lengkap.",
        ],
      },
      {
        heading: "Cara membeli paket telepon di Isiloka",
        body: [
          "Masukkan nomor tujuan, pilih operator yang sesuai, lalu pilih paket telepon berdasarkan durasi, jumlah menit, atau masa aktif yang tersedia.",
          "Setelah itu lanjutkan ke pembayaran seperti kategori produk digital lainnya dan cek status transaksi hingga selesai.",
        ],
      },
      {
        heading: "Nilainya untuk member dan agen",
        body: [
          "Paket telepon sangat cocok dijadikan produk pelengkap untuk pelanggan yang sudah sering membeli pulsa dan kuota.",
          "Dengan menambahkan kategori ini, member bisa menjangkau kebutuhan komunikasi yang lebih luas tanpa harus mengubah pola layanan yang sudah berjalan.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah paket telepon masih relevan di era internet?", answer: "Masih. Paket telepon Telkomsel, Indosat, XL, dan Smartfren tetap dibutuhkan untuk kerja, keluarga, dan komunikasi penting yang lebih nyaman lewat panggilan suara." },
      { question: "Apakah paket telepon cocok dijual bersama pulsa dan paket data?", answer: "Ya. Paket telepon sangat cocok dipasangkan dengan pulsa reguler dan kuota operator yang sama agar pelanggan lebih mudah memilih." },
    ],
  },
  {
    slug: "cara-beli-pulsa-data-dan-kombinasi-pulsa-kuota",
    title: "Cara Beli Pulsa Data dan Kombinasi Pulsa Kuota di Isiloka",
    description: "Panduan membeli pulsa data dan produk kombinasi pulsa-kuota di Isiloka untuk pelanggan yang butuh pilihan lebih fleksibel.",
    excerpt: "Pulsa data cocok untuk pelanggan yang butuh kombinasi nominal, kuota, atau produk operator tertentu dalam satu alur yang lebih praktis.",
    category: "Pulsa Data",
    readTime: "5 menit",
    keywords: ["pulsa data", "beli pulsa data", "pulsa kuota", "produk operator data"],
    hero: "Pulsa data membantu Isiloka menjangkau kebutuhan operator yang tidak selalu pas jika hanya dibagi antara pulsa reguler dan paket data biasa.",
    products: ["Pulsa Data", "Pulsa Telkomsel", "Paket Data Telkomsel"],
    ctaLabel: "Buka Pulsa Data",
    ctaHref: "/pulsa-data",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Apa itu pulsa data di Isiloka",
        body: [
          "Pulsa data adalah kategori yang menjembatani kebutuhan pelanggan yang mencari produk operator dengan karakter lebih khusus daripada pulsa reguler.",
          "Kategori ini membantu pembeli yang ingin pilihan lebih fleksibel sesuai produk yang tersedia dari operator atau provider.",
        ],
      },
      {
        heading: "Cara memilih produk yang tepat",
        body: [
          "Mulailah dari nomor tujuan dan operator yang benar, lalu bandingkan apakah kebutuhan pelanggan lebih cocok ke pulsa reguler, paket data, atau pulsa data.",
          "Jika pelanggan membutuhkan produk operator tertentu yang tidak nyaman dicari di kategori biasa, pulsa data bisa menjadi jalur yang lebih jelas.",
        ],
      },
      {
        heading: "Kenapa kategori ini penting untuk layanan lengkap",
        body: [
          "Pulsa data membantu Isiloka terlihat lebih lengkap karena mampu menutup celah antara pulsa reguler dan kuota biasa.",
          "Bagi member, kategori ini berguna untuk menjaga agar pelanggan tidak pindah ke tempat lain hanya karena mencari produk operator yang lebih spesifik.",
        ],
      },
    ],
    faqs: [
      { question: "Apa beda pulsa data dengan paket data?", answer: "Paket data biasanya fokus ke kuota dan masa aktif tertentu, sedangkan pulsa data lebih cocok untuk produk operator yang lebih khusus atau kebutuhan campuran antara pulsa dan kuota." },
      { question: "Kapan pelanggan lebih cocok diarahkan ke pulsa data?", answer: "Saat pelanggan tidak menemukan kebutuhan yang pas di pulsa reguler atau paket data biasa, lalu butuh produk operator yang lebih spesifik." },
    ],
  },
  {
    slug: "cara-bayar-pgn-online-dan-layanan-energi-rumah-tangga",
    title: "Cara Bayar PGN Online dan Melengkapi Layanan Energi Rumah Tangga",
    description: "Panduan membayar PGN online melalui Isiloka untuk pelanggan rumah tangga dan agen yang ingin layanan energinya lebih lengkap.",
    excerpt: "PGN adalah kategori yang membuat layanan rumah tangga di Isiloka terasa lebih lengkap bersama listrik, PDAM, dan internet pascabayar.",
    category: "PGN",
    readTime: "5 menit",
    keywords: ["bayar pgn online", "tagihan gas online", "pgn rumah tangga", "ppob pgn"],
    hero: "PGN memperluas layanan Isiloka ke kebutuhan energi rumah tangga yang lebih lengkap dan lebih bernilai untuk pelanggan bulanan.",
    products: ["PGN", "Listrik", "PDAM", "Internet Pascabayar"],
    ctaLabel: "Bayar PGN",
    ctaHref: "/pgn",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa PGN penting untuk layanan rumah tangga",
        body: [
          "PGN memang tidak dicari sebanyak pulsa atau kuota, tetapi tetap penting untuk pelanggan rumah tangga yang ingin semua pembayaran bulanan ada di satu tempat.",
          "Kategori seperti ini memberi kesan bahwa Isiloka bukan hanya platform top up, tetapi juga pusat pembayaran digital keluarga.",
        ],
      },
      {
        heading: "Cara bayar PGN di Isiloka",
        body: [
          "Masuk ke kategori PGN, isi nomor pelanggan atau data yang dibutuhkan, lalu cek detail tagihan yang muncul sebelum melanjutkan pembayaran.",
          "Setelah transaksi selesai, simpan referensi pembayaran agar lebih mudah saat pelanggan menanyakan riwayatnya.",
        ],
      },
      {
        heading: "Keuntungan untuk member dan masa depan layanan",
        body: [
          "Semakin lengkap kategori rumah tangga yang dijual member, semakin besar peluang pelanggan menjadikan layanan itu sebagai tempat pembayaran bulanan utama.",
          "Inilah salah satu masa depan Isiloka: tumbuh dari kebutuhan harian seperti pulsa dan kuota ke kebutuhan rumah tangga yang lebih menyeluruh.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah PGN layak disediakan meski pasarnya lebih kecil?", answer: "Ya. Justru kategori yang lebih spesifik membantu layanan terlihat lebih lengkap dan profesional." },
      { question: "Apakah PGN cocok dipadukan dengan produk lain?", answer: "Ya. PGN sangat cocok dipadukan dengan listrik, PDAM, internet pascabayar, dan TV berlangganan." },
    ],
  },
  {
    slug: "cara-top-up-saldo-member-dengan-qris-dan-bank",
    title: "Cara Top Up Saldo Member dengan QRIS dan Bank di Isiloka",
    description: "Panduan top up saldo member Isiloka lewat QRIS dan bank agar transaksi harian dan operasional agen tetap lancar.",
    excerpt: "Saldo adalah bahan bakar utama member. Karena itu, alur top up yang jelas sangat penting untuk menjaga transaksi tetap jalan tanpa hambatan.",
    category: "Deposit",
    readTime: "6 menit",
    keywords: ["top up saldo member", "deposit qris", "saldo pulsakilat", "isi saldo agen"],
    hero: "Top up saldo yang jelas dan cepat membantu member menjaga operasional tetap stabil, terutama saat transaksi sedang ramai.",
    products: ["Deposit QRIS", "Transfer Bank", "Saldo Member"],
    ctaLabel: "Masuk untuk Isi Saldo Member",
    ctaHref: "/login",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa saldo member sangat penting",
        body: [
          "Saldo adalah fondasi transaksi untuk member retail maupun agen. Tanpa saldo yang cukup, pelayanan bisa terhenti saat pelanggan justru sedang ramai bertransaksi.",
          "Karena itu, top up saldo bukan sekadar fitur tambahan, tetapi bagian inti dari operasional bisnis di Isiloka.",
        ],
      },
      {
        heading: "Cara top up saldo di Isiloka",
        body: [
          "Member bisa memilih metode QRIS atau transfer bank sesuai kebutuhan. Setelah memilih nominal, cek ringkasan pembayaran, selesaikan prosesnya, lalu pantau status hingga saldo benar-benar masuk.",
          "Untuk QRIS, pembagian nominal, fee admin, dan total bayar harus terlihat jelas agar member mudah memverifikasi pembayaran yang sedang berjalan.",
        ],
      },
      {
        heading: "Hubungannya dengan masa depan usaha member",
        body: [
          "Usaha yang ingin tumbuh harus punya ritme operasional yang stabil, dan itu dimulai dari manajemen saldo yang rapi. Top up saldo yang mudah membantu member melayani pelanggan tanpa jeda yang merugikan.",
          "Dalam jangka panjang, alur saldo yang sehat mendukung pertumbuhan member dari transaksi manual ke operasional yang lebih besar, termasuk website sendiri dan H2H.",
        ],
      },
    ],
    faqs: [
      { question: "Lebih baik top up saldo lewat QRIS atau bank?", answer: "QRIS cocok untuk isi saldo cepat saat transaksi sedang berjalan, sedangkan transfer bank cocok untuk pola deposit yang lebih terencana." },
      { question: "Kenapa saldo penting untuk calon member?", answer: "Karena ketersediaan saldo menentukan apakah pulsa, kuota, e-wallet, dan token listrik bisa langsung diproses saat pelanggan datang." },
    ],
  },
  {
    slug: "perbedaan-member-retail-agen-dan-h2h-di-pulsakilat",
    title: "Perbedaan Member Retail, Agen, dan H2H di Isiloka",
    description: "Panduan memahami beda retail, agen, dan H2H di Isiloka agar calon member bisa memilih model usaha yang paling sesuai.",
    excerpt: "Tidak semua member harus langsung masuk ke H2H. Yang penting adalah memahami tahap usaha dan memilih model yang tepat sejak awal.",
    category: "Panduan Member",
    readTime: "7 menit",
    keywords: ["member retail", "agen pulsa", "h2h pulsakilat", "perbedaan retail agen h2h"],
    hero: "Isiloka bisa dipakai dari level transaksi harian sampai integrasi bisnis. Bedanya ada pada cara kerja, skala, dan kebutuhan operasionalnya.",
    products: ["Pulsa Telkomsel", "DANA", "Token Listrik Promo", "BPJS Kesehatan", "API H2H"],
    ctaLabel: "Lihat Jalur Retail, Agen, dan H2H",
    ctaHref: "/artikel/cara-bertumbuh-dari-retail-ke-agen-dan-h2h-bersama-pulsakilat",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kapan retail sudah cukup",
        body: [
          "Model retail cocok untuk pengguna yang ingin transaksi langsung untuk diri sendiri atau untuk melayani pelanggan sekitar dalam volume kecil sampai menengah.",
          "Pada tahap ini, fokus utamanya adalah memahami produk, menjaga saldo, dan membangun kepercayaan pelanggan dari transaksi yang rapi.",
        ],
      },
      {
        heading: "Kapan masuk ke level agen",
        body: [
          "Model agen cocok ketika transaksi mulai rutin dan pelanggan mulai datang berulang. Pada tahap ini, member biasanya mulai memikirkan margin, repeat order, dan perluasan kategori produk.",
          "Agen membutuhkan disiplin operasional yang lebih tinggi karena layanan sudah menyangkut banyak pelanggan, bukan hanya transaksi sesekali.",
        ],
      },
      {
        heading: "Kapan H2H jadi langkah yang masuk akal",
        body: [
          "H2H tepat ketika bisnis mulai butuh otomasi, website sendiri, panel reseller, atau integrasi sistem agar order tidak lagi dikelola sepenuhnya secara manual.",
          "Dengan memahami perbedaan tiga model ini, calon member bisa merancang masa depan usahanya di Isiloka dengan lebih realistis.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah member baru harus langsung masuk H2H?", answer: "Tidak. Lebih sehat jika memulai dari produk dasar seperti pulsa, kuota, e-wallet, dan token listrik sampai alurnya benar-benar dipahami." },
      { question: "Apa penanda bahwa bisnis sudah siap ke H2H?", answer: "Biasanya saat transaksi rutin mulai tinggi, order datang dari banyak kanal, dan Anda mulai butuh website, panel, atau integrasi agar proses tidak lagi manual." },
    ],
  },
  {
    slug: "cara-bertumbuh-dari-retail-ke-agen-dan-h2h-bersama-pulsakilat",
    title: "Cara Bertumbuh dari Retail ke Agen dan H2H Bersama Isiloka",
    description: "Panduan bertahap untuk tumbuh dari retail ke agen lalu ke H2H di Isiloka tanpa kehilangan kendali operasional.",
    excerpt: "Masa depan usaha digital tidak dibangun dengan loncat tahap. Yang lebih sehat adalah bertumbuh bertahap dari retail, lalu agen, lalu H2H saat memang siap.",
    category: "Bisnis Agen",
    readTime: "8 menit",
    keywords: ["naik dari retail ke agen", "bertumbuh ke h2h", "masa depan usaha pulsa", "pulsakilat bisnis"],
    hero: "Isiloka paling bernilai ketika dipakai sebagai jalur pertumbuhan usaha, bukan hanya sebagai tempat transaksi harian.",
    products: ["Pulsa Telkomsel", "DANA", "Token Listrik Promo", "Website Jualan", "Saldo Member"],
    ctaLabel: "Daftar Jadi Member",
    ctaHref: "/register",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Bangun pondasi dari retail yang rapi",
        body: [
          "Tahap retail adalah fase membangun dasar: memahami kategori produk, mengelola saldo, mencatat transaksi, dan belajar pola kebutuhan pelanggan.",
          "Pada tahap ini, tujuan utamanya bukan besar dulu, tetapi membuat alur layanan stabil dan minim kesalahan.",
        ],
      },
      {
        heading: "Naik ke agen saat repeat order mulai terbentuk",
        body: [
          "Saat pelanggan mulai rutin membeli pulsa, kuota, e-wallet, listrik, atau PPOB, member sudah memasuki fase agen. Di sini fokus bergeser ke katalog produk, margin, dan pelayanan yang konsisten.",
          "Semakin lengkap produk yang dijual, semakin besar peluang pelanggan menjadikan usaha Anda sebagai tempat transaksi utama.",
        ],
      },
      {
        heading: "Masuk ke H2H untuk masa depan usaha yang lebih besar",
        body: [
          "Ketika order makin ramai, website sendiri mulai dibutuhkan, atau integrasi sistem mulai terasa penting, H2H menjadi langkah logis berikutnya.",
          "Masa depan usaha bersama Isiloka adalah kemampuan untuk naik tahap tanpa harus mengganti fondasi produk yang sudah Anda bangun sejak awal.",
        ],
      },
    ],
    faqs: [
      { question: "Apa kesalahan paling umum saat ingin cepat naik level?", answer: "Melompat ke tahap yang lebih kompleks sebelum pondasi retail dan agen benar-benar rapi." },
      { question: "Apakah semua usaha harus punya website dan H2H?", answer: "Tidak selalu. Tetapi untuk usaha yang ingin tumbuh lebih besar, dua hal itu adalah jalur yang sangat masuk akal." },
    ],
  },
  {
    slug: "produk-pulsakilat-yang-paling-cocok-untuk-calon-member",
    title: "Produk Isiloka yang Paling Cocok untuk Calon Member",
    description: "Panduan memilih produk Isiloka untuk calon member, dari pulsa, paket data, e-wallet, game, listrik, sampai PPOB rumah tangga.",
    excerpt: "Calon member tidak harus langsung menjual semua kategori. Yang penting adalah memilih produk awal yang paling mudah dijalankan dan paling sering dibutuhkan pelanggan.",
    category: "Panduan Member",
    readTime: "8 menit",
    keywords: ["produk pulsakilat", "calon member pulsakilat", "produk paling laris", "usaha produk digital"],
    hero: "Calon member akan lebih mudah tumbuh jika memulai dari kategori produk yang paling dekat dengan kebutuhan pasar di sekitarnya.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "Paket Telepon Telkomsel", "Pulsa Data", "DANA", "Mobile Legends", "Token Listrik Promo", "BPJS Kesehatan", "PGN"],
    ctaLabel: "Buka Produk Awal yang Paling Aman",
    ctaHref: "/pulsa",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Mulai dari produk yang paling mudah dijual",
        body: [
          "Pulsa, paket data, dan e-wallet adalah tiga kategori yang paling mudah dipahami pelanggan dan paling cepat dipasarkan oleh calon member.",
          "Tiga kategori ini cocok untuk tahap awal karena kebutuhan pasarnya luas dan proses transaksinya relatif sederhana.",
        ],
      },
      {
        heading: "Tambah kategori rumah tangga setelah tahap awal stabil",
        body: [
          "Setelah layanan dasar berjalan rapi, calon member bisa menambah token listrik, tagihan listrik, BPJS, PDAM, internet pascabayar, TV, PGN, paket telepon, dan pulsa data.",
          "Kategori rumah tangga ini sangat penting untuk repeat order karena banyak pelanggan membayarnya rutin setiap bulan.",
        ],
      },
      {
        heading: "Naik level saat transaksi mulai ramai",
        body: [
          "Jika transaksi sudah makin ramai, calon member bisa mulai memikirkan website sendiri, pengelolaan markup yang lebih rapi, dan integrasi H2H untuk otomasi.",
          "Dengan jalur bertumbuh seperti ini, Isiloka tidak hanya membantu transaksi harian, tetapi juga perkembangan usaha jangka panjang.",
        ],
      },
    ],
    faqs: [
      { question: "Produk apa yang paling cocok untuk calon member baru?", answer: "Pulsa operator besar, paket data, dan top up e-wallet adalah titik awal terbaik karena paling cepat dipahami dan paling sering dicari pelanggan." },
      { question: "Kapan calon member sebaiknya menambah layanan rumah tangga?", answer: "Saat transaksi dasar sudah stabil, lalu tambahkan token listrik, BPJS, PDAM, dan PGN untuk membangun repeat order bulanan." },
    ],
  },
  {
    slug: "cara-membangun-warung-pulsa-dan-ppob-yang-lengkap",
    title: "Cara Membangun Warung Pulsa dan PPOB yang Lengkap",
    description: "Panduan membangun warung pulsa dan PPOB dengan produk yang tepat, layanan rumah tangga yang lengkap, dan arah tumbuh bersama Isiloka.",
    excerpt: "Warung pulsa yang kuat bukan hanya menjual nominal kecil, tetapi juga menyediakan layanan rumah tangga yang dicari pelanggan setiap bulan.",
    category: "Bisnis Agen",
    readTime: "7 menit",
    keywords: ["warung pulsa", "warung ppob", "jualan pulsa dan ppob", "konter pulsa lengkap"],
    hero: "Warung yang lengkap lebih mudah menjadi tempat transaksi utama pelanggan dibanding warung yang hanya mengandalkan pulsa reguler.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "Token Listrik Promo", "BPJS Kesehatan", "DANA"],
    ctaLabel: "Buka Produk Warung dan PPOB",
    ctaHref: "/listrik",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Produk dasar yang wajib ada di warung",
        body: [
          "Pulsa, paket data, e-wallet, dan token listrik adalah empat kategori yang paling aman untuk tahap awal. Kebutuhannya tinggi dan mudah dipahami pelanggan warung.",
          "Begitu alur transaksi dasar sudah rapi, warung bisa menambah BPJS, PDAM, internet pascabayar, dan tagihan rumah tangga lain.",
        ],
      },
      {
        heading: "Kenapa layanan rumah tangga penting",
        body: [
          "Pelanggan yang datang untuk kebutuhan bulanan seperti listrik, BPJS, dan PDAM cenderung kembali ke tempat yang sama jika layanannya mudah dan jelas.",
          "Inilah alasan warung PPOB yang lengkap punya peluang repeat order lebih kuat daripada warung yang hanya fokus pada top up harian.",
        ],
      },
      {
        heading: "Arah tumbuh setelah warung mulai ramai",
        body: [
          "Saat pelanggan mulai rutin dan produk makin banyak, pemilik warung perlu memikirkan saldo, margin, pencatatan, dan kemungkinan memperluas layanan ke kanal online.",
          "Dari titik ini, Isiloka bisa dipakai bukan hanya untuk melayani warung harian, tetapi juga untuk menyiapkan langkah usaha yang lebih besar.",
        ],
      },
    ],
    faqs: [
      { question: "Produk apa yang paling penting untuk warung pulsa baru?", answer: "Pulsa operator besar, paket data, top up e-wallet, dan token listrik adalah kombinasi awal paling aman untuk warung yang ingin cepat dipakai pelanggan." },
      { question: "Kenapa warung pulsa perlu menambah PPOB rumah tangga?", answer: "Karena BPJS, PDAM, listrik, dan PGN memberi repeat order bulanan dan membuat warung lebih mudah jadi tempat pembayaran utama di sekitar rumah." },
    ],
  },
  {
    slug: "cara-menyusun-katalog-produk-digital-yang-mudah-dipahami",
    title: "Cara Menyusun Katalog Produk Digital yang Mudah Dipahami",
    description: "Panduan menyusun katalog produk digital agar pembeli cepat paham, mudah memilih, dan lebih nyaman bertransaksi di layanan Isiloka Anda.",
    excerpt: "Katalog yang rapi membuat pembeli lebih cepat memilih produk dan mengurangi kebingungan saat pilihan nominal atau kategori semakin banyak.",
    category: "Operasional",
    readTime: "6 menit",
    keywords: ["katalog produk digital", "susun produk pulsa", "kategori produk ppob", "layout produk digital"],
    hero: "Katalog yang mudah dipahami adalah salah satu pembeda terbesar antara layanan yang terasa profesional dan layanan yang terasa membingungkan.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "DANA", "Mobile Legends", "Token Listrik Promo"],
    ctaLabel: "Lihat Susunan Kategori Produk",
    ctaHref: "/pulsa",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Mulai dari kategori yang paling familiar",
        body: [
          "Pulsa, paket data, e-wallet, dan listrik sebaiknya ditempatkan sebagai kategori yang paling mudah ditemukan. Ini membantu pembeli baru memahami isi layanan sejak awal.",
          "Setelah kategori utama jelas, barulah produk pelengkap seperti game, paket telepon, pulsa data, dan layanan PPOB lain ditambahkan dengan urutan yang rapi.",
        ],
      },
      {
        heading: "Pisahkan produk yang memang butuh grup",
        body: [
          "Paket data, token listrik, atau kategori lain yang variasinya banyak perlu dipisah dengan grup atau label yang mudah dipahami. Tujuannya agar pembeli tidak tenggelam dalam daftar panjang yang campur aduk.",
          "Judul produk juga harus langsung menunjukkan isi penting seperti nominal, kuota, atau masa aktif, bukan mengulang nama kategori yang sama.",
        ],
      },
      {
        heading: "Dampaknya ke konversi dan layanan pelanggan",
        body: [
          "Katalog yang rapi membuat pembeli lebih cepat sampai ke keputusan. Ini juga mengurangi pertanyaan berulang yang sebenarnya bisa dicegah lewat struktur halaman yang jelas.",
          "Untuk member, katalog yang baik membantu layanan terasa lebih profesional meski bisnis masih dijalankan dari rumah atau warung kecil.",
        ],
      },
    ],
    faqs: [
      { question: "Kategori apa yang harus ditaruh paling depan?", answer: "Biasanya pulsa, paket data, e-wallet, dan listrik karena paling sering dicari pelanggan harian." },
      { question: "Kenapa judul produk harus ringkas?", answer: "Agar pembeli langsung paham isi produk tanpa terganggu pengulangan nama kategori atau label yang tidak penting." },
    ],
  },
  {
    slug: "cara-mengelola-pelanggan-rutin-token-listrik-bpjs-dan-pdam",
    title: "Cara Mengelola Pelanggan Rutin Token Listrik, BPJS, dan PDAM",
    description: "Panduan mengelola pelanggan rutin untuk token listrik, BPJS, PDAM, dan tagihan rumah tangga agar repeat order lebih stabil di Isiloka.",
    excerpt: "Pelanggan rutin adalah fondasi usaha PPOB yang sehat karena memberi arus transaksi berulang dan lebih mudah dipertahankan.",
    category: "Operasional",
    readTime: "7 menit",
    keywords: ["pelanggan rutin ppob", "pelanggan token listrik", "pelanggan bpjs pdam", "repeat order ppob"],
    hero: "Layanan rumah tangga yang dibayar rutin memberi peluang paling besar untuk membangun hubungan pelanggan jangka panjang.",
    products: ["Token Listrik Promo", "BPJS Kesehatan", "PDAM", "Internet Pascabayar"],
    ctaLabel: "Buka Layanan Rumah Tangga",
    ctaHref: "/listrik",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa pelanggan rutin sangat berharga",
        body: [
          "Pelanggan yang datang tiap bulan untuk listrik, BPJS, atau PDAM lebih mudah dipertahankan daripada pelanggan yang hanya membeli sesekali.",
          "Jika layanannya cepat dan data pelanggan tersimpan rapi, mereka cenderung kembali tanpa banyak pertimbangan.",
        ],
      },
      {
        heading: "Data apa yang perlu dijaga",
        body: [
          "Catat nomor pelanggan, jenis layanan, jadwal pembayaran yang biasa dilakukan, dan preferensi kanal komunikasi. Data sederhana seperti ini sangat membantu pelayanan bulan berikutnya.",
          "Dengan pencatatan yang rapi, agen tidak perlu mengulang proses dari nol setiap kali pelanggan datang lagi.",
        ],
      },
      {
        heading: "Cara membuat repeat order lebih stabil",
        body: [
          "Gabungkan layanan token listrik, tagihan listrik, BPJS, PDAM, internet pascabayar, dan TV agar pelanggan melihat usaha Anda sebagai pusat pembayaran rumah tangga.",
          "Semakin lengkap layanannya, semakin besar peluang pelanggan menjadikan Anda pilihan utama setiap bulan.",
        ],
      },
    ],
    faqs: [
      { question: "Kategori apa yang paling kuat untuk pelanggan rutin?", answer: "Token listrik, BPJS, PDAM, internet pascabayar, dan layanan rumah tangga lain adalah kategori paling kuat untuk repeat order bulanan." },
      { question: "Data apa yang paling penting disimpan untuk pelanggan rutin?", answer: "Minimal simpan nomor pelanggan, jenis layanan, kebiasaan waktu bayar, dan catatan transaksi terakhir agar pelayanan bulan berikutnya lebih cepat." },
    ],
  },
  {
    slug: "cara-membangun-panel-reseller-dengan-h2h-pulsakilat",
    title: "Cara Membangun Panel Reseller dengan H2H Isiloka",
    description: "Panduan membangun panel reseller dengan H2H Isiloka untuk bisnis yang ingin menerima order otomatis dan menata operasional lebih rapi.",
    excerpt: "Panel reseller butuh produk yang stabil, alur order yang jelas, dan fondasi integrasi yang siap menampung transaksi dari banyak pengguna.",
    category: "H2H",
    readTime: "8 menit",
    keywords: ["panel reseller", "h2h pulsakilat", "api panel pulsa", "reseller otomatis"],
    hero: "Panel reseller adalah langkah lanjutan yang masuk akal ketika bisnis tidak lagi nyaman dikelola dengan order manual.",
    products: ["API H2H", "Pulsa Telkomsel", "Paket Data Telkomsel", "BPJS Kesehatan", "DANA"],
    ctaLabel: "Buka Dokumentasi H2H",
    ctaHref: "/docs",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kapan panel reseller mulai dibutuhkan",
        body: [
          "Panel reseller dibutuhkan saat order mulai datang dari banyak agen, komunitas, atau kanal sekaligus. Dalam kondisi ini, proses manual akan cepat melelahkan dan rawan salah input.",
          "Dengan panel, bisnis bisa menata produk, harga, dan order dari satu sistem yang lebih mudah diawasi.",
        ],
      },
      {
        heading: "Produk yang paling aman dijadikan fondasi",
        body: [
          "Pulsa, paket data, e-wallet, dan PPOB rumah tangga adalah fondasi terbaik karena pasarnya luas dan repeat order-nya kuat. Produk seperti game juga bisa ditambahkan untuk memperluas segmen.",
          "Semakin seimbang komposisi produknya, semakin sehat peluang transaksi reseller berjalan konsisten.",
        ],
      },
      {
        heading: "Apa yang harus disiapkan selain integrasi",
        body: [
          "Panel reseller tidak cukup hanya dengan API. Bisnis juga perlu memikirkan saldo, pencatatan, dukungan pelanggan, dan aturan harga agar operasional tetap sehat.",
          "Isiloka cocok dipakai di tahap ini karena memberi jalur dari member biasa ke integrasi yang lebih serius tanpa mengganti fondasi produk.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah panel reseller harus langsung dibuat sejak awal?", answer: "Tidak. Panel reseller lebih tepat saat volume transaksi sudah cukup ramai dan proses manual mulai terasa berat." },
      { question: "Produk apa yang paling aman untuk panel reseller?", answer: "Pulsa, paket data, e-wallet, dan PPOB rumah tangga adalah fondasi terbaik karena pasarnya luas dan repeat order-nya kuat." },
    ],
  },
  {
    slug: "cara-menjual-top-up-game-untuk-komunitas-dan-anak-muda",
    title: "Cara Menjual Top Up Game untuk Komunitas dan Anak Muda",
    description: "Panduan menjual top up game untuk komunitas dan anak muda dengan layanan yang cepat, pilihan produk yang tepat, dan dukungan pembayaran yang nyaman.",
    excerpt: "Pasar top up game bergerak cepat. Karena itu, layanan yang ringkas dan mudah dipahami lebih penting daripada tampilan yang ramai tetapi membingungkan.",
    category: "Game",
    readTime: "6 menit",
    keywords: ["jualan top up game", "komunitas game", "top up game anak muda", "bisnis game digital"],
    hero: "Komunitas gamer mencari layanan yang cepat, jelas, dan mudah dipakai berulang kali tanpa banyak langkah tambahan.",
    products: ["Mobile Legends", "DANA", "Paket Data Telkomsel"],
    ctaLabel: "Buka Kategori Game",
    ctaHref: "/game",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa pasar game menarik untuk member",
        body: [
          "Pelanggan game biasanya membeli cepat dan berulang, terutama saat ada event, turnamen kecil, atau kebutuhan item harian. Ini membuat kategori game menarik untuk pasar anak muda.",
          "Jika flow transaksi rapi, produk game bisa menjadi pintu masuk yang kuat untuk membangun pelanggan loyal.",
        ],
      },
      {
        heading: "Produk pelengkap yang paling masuk akal",
        body: [
          "Selain game, e-wallet dan paket data sangat cocok dijadikan pelengkap. Banyak pelanggan game membutuhkan dua kategori ini untuk mendukung kebiasaan bermain mereka.",
          "Kombinasi ini membuat layanan Anda terasa lebih lengkap tanpa harus membuka terlalu banyak kategori sejak awal.",
        ],
      },
      {
        heading: "Cara mendekati komunitas",
        body: [
          "Komunitas lebih mudah percaya pada layanan yang cepat, konsisten, dan responsif. Karena itu, kejelasan produk, ketepatan transaksi, dan komunikasi yang rapi jauh lebih penting daripada promosi berlebihan.",
          "Jika komunitas mulai rutin membeli, kategori game bisa berkembang menjadi salah satu penggerak repeat order yang paling aktif.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah top up game cocok untuk member pemula?", answer: "Ya, terutama jika digabung dengan e-wallet dan paket data agar layanannya terasa lebih lengkap." },
      { question: "Kenapa komunitas game penting untuk repeat order?", answer: "Karena kebutuhan item, diamond, dan top up biasanya berulang, terutama saat ada event atau rutinitas bermain." },
    ],
  },
  {
    slug: "cara-mengelola-layanan-pelanggan-usaha-produk-digital",
    title: "Cara Mengelola Layanan Pelanggan untuk Usaha Produk Digital",
    description: "Panduan mengelola layanan pelanggan usaha produk digital agar transaksi lebih rapi, keluhan lebih cepat selesai, dan pelanggan lebih mudah kembali.",
    excerpt: "Layanan pelanggan yang rapi membuat produk yang sama terasa lebih meyakinkan dibanding pesaing yang hanya mengandalkan harga murah.",
    category: "Operasional",
    readTime: "7 menit",
    keywords: ["layanan pelanggan pulsa", "customer service produk digital", "pelayanan agen pulsa", "usaha produk digital"],
    hero: "Dalam usaha produk digital, cara melayani pelanggan sering menjadi pembeda utama saat harga antar penjual terlihat mirip.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "DANA", "Mobile Legends", "Token Listrik Promo"],
    ctaLabel: "Buka Produk untuk Layanan Pelanggan",
    ctaHref: "/pulsa",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Apa yang paling dicari pelanggan",
        body: [
          "Pelanggan mencari layanan yang cepat, jelas, dan mudah dihubungi saat ada masalah. Mereka tidak hanya menilai harga, tetapi juga bagaimana transaksi ditangani saat ada kendala.",
          "Semakin rapi komunikasi Anda, semakin besar peluang pelanggan percaya untuk kembali bertransaksi.",
        ],
      },
      {
        heading: "Cara menangani pertanyaan dan keluhan",
        body: [
          "Jawab pertanyaan dengan singkat, tunjukkan status transaksi dengan jelas, dan simpan referensi penting agar penelusuran masalah lebih cepat. Pola ini membantu mengurangi emosi yang tidak perlu saat ada kendala.",
          "Untuk usaha yang makin ramai, pencatatan transaksi dan saldo harus sejalan dengan pelayanan agar jawaban ke pelanggan tetap akurat.",
        ],
      },
      {
        heading: "Hubungannya dengan pertumbuhan usaha",
        body: [
          "Pelayanan yang baik membuat pelanggan lebih mudah kembali untuk produk lain seperti listrik, BPJS, PDAM, atau game. Inilah cara layanan pelanggan membantu penjualan lintas kategori.",
          "Saat usaha tumbuh ke website sendiri atau H2H, kebiasaan pelayanan yang rapi tetap menjadi fondasi yang tidak boleh hilang.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah layanan pelanggan penting untuk usaha kecil?", answer: "Sangat penting, karena kecepatan dan kejelasan layanan sering menjadi alasan pelanggan kembali meski selisih harga kecil." },
      { question: "Apa kaitan layanan pelanggan dengan repeat order?", answer: "Pelanggan lebih mudah kembali jika pengalaman transaksi sebelumnya terasa jelas, aman, dan mudah ditelusuri saat ada masalah." },
    ],
  },
  {
    slug: "cara-membangun-warung-pulsa-desa-yang-tetap-ramai",
    title: "Cara Membangun Warung Pulsa Desa yang Tetap Ramai",
    description: "Panduan membangun warung pulsa desa dengan produk yang dibutuhkan warga, layanan rumah tangga, dan pola repeat order yang lebih stabil.",
    excerpt: "Warung pulsa desa akan lebih kuat jika melayani kebutuhan harian dan bulanan warga sekaligus, bukan hanya pulsa reguler.",
    category: "Bisnis Agen",
    readTime: "7 menit",
    keywords: ["warung pulsa desa", "jualan pulsa desa", "ppob desa", "usaha pulsa kampung"],
    hero: "Di area desa, layanan yang lengkap dan konsisten sering lebih penting daripada promosi yang besar tetapi tidak berkelanjutan.",
    products: ["Pulsa Telkomsel", "Token Listrik Promo", "BPJS Kesehatan", "Paket Data Telkomsel", "DANA"],
    ctaLabel: "Buka Produk Warung Desa",
    ctaHref: "/listrik",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Produk yang paling dekat dengan kebutuhan warga",
        body: [
          "Pulsa, paket data, token listrik, tagihan listrik, BPJS, dan PDAM adalah kategori yang paling mudah diterima di lingkungan desa karena langsung berhubungan dengan kebutuhan rumah tangga.",
          "Jika warung hanya menjual pulsa reguler, pelanggan akan lebih mudah berpindah ke tempat lain saat membutuhkan layanan bulanan.",
        ],
      },
      {
        heading: "Kenapa kedekatan layanan lebih penting",
        body: [
          "Di lingkungan desa, pelanggan sering kembali ke tempat yang cepat, jelas, dan mudah ditanya saat ada masalah. Faktor kedekatan dan kepercayaan sangat besar.",
          "Karena itu, pelayanan yang rapi dan pencatatan pelanggan rutin akan memberi dampak yang lebih nyata daripada sekadar perang harga.",
        ],
      },
      {
        heading: "Arah tumbuh setelah pelanggan mulai tetap",
        body: [
          "Saat warga mulai rutin membayar listrik, BPJS, atau PDAM di warung Anda, usaha sudah punya pondasi repeat order yang bagus. Pada titik ini, pengelolaan saldo dan kategori produk jadi semakin penting.",
          "Isiloka cocok untuk pola seperti ini karena kategori rumah tangganya bisa melengkapi top up harian dalam satu layanan.",
        ],
      },
    ],
    faqs: [
      { question: "Produk apa yang paling cocok untuk warung pulsa desa?", answer: "Pulsa, paket data, token listrik, BPJS, dan PDAM adalah kombinasi yang paling dekat dengan kebutuhan warga sehari-hari." },
      { question: "Kenapa PPOB penting untuk warung di desa?", answer: "Karena layanan rumah tangga memberi alasan kuat bagi pelanggan untuk kembali setiap bulan dan membangun kebiasaan transaksi yang stabil." },
    ],
  },
  {
    slug: "cara-jadi-reseller-produk-digital-lewat-whatsapp",
    title: "Cara Jadi Reseller Produk Digital lewat WhatsApp",
    description: "Panduan menjadi reseller produk digital lewat WhatsApp dengan alur order yang rapi, produk yang tepat, dan layanan yang mudah diulang pelanggan.",
    excerpt: "Banyak usaha produk digital tumbuh dari chat WhatsApp. Yang penting bukan sekadar cepat membalas, tetapi membuat order mudah dan minim salah.",
    category: "Bisnis Agen",
    readTime: "6 menit",
    keywords: ["reseller whatsapp", "jualan pulsa whatsapp", "reseller produk digital", "order pulsa via whatsapp"],
    hero: "WhatsApp tetap menjadi jalur penjualan yang kuat untuk usaha pulsa, PPOB, dan top up digital selama alur layanannya dibuat jelas.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "DANA", "Token Listrik Promo"],
    ctaLabel: "Daftar Jadi Member",
    ctaHref: "/register",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa WhatsApp masih efektif",
        body: [
          "Banyak pelanggan lebih nyaman memesan lewat chat karena terasa cepat dan akrab. Ini membuat WhatsApp tetap relevan untuk reseller yang baru mulai.",
          "Namun ketika transaksi mulai bertambah, format order dan pencatatan harus dibuat rapi agar tidak mudah tertukar.",
        ],
      },
      {
        heading: "Produk awal yang paling aman dijual",
        body: [
          "Pulsa, paket data, e-wallet, dan token listrik adalah kombinasi awal yang paling mudah dijual lewat chat. Nama produknya familiar dan kebutuhan pasarnya luas.",
          "Jika pelanggan sudah mulai tetap, reseller bisa menambah layanan rumah tangga dan game untuk memperluas repeat order.",
        ],
      },
      {
        heading: "Batas WhatsApp dan kapan harus naik level",
        body: [
          "Saat order mulai ramai, reseller butuh pencatatan transaksi, pengelolaan saldo, dan struktur produk yang lebih rapi. Pada titik ini, usaha perlu mulai memikirkan katalog online atau website sederhana.",
          "Isiloka cocok untuk tahap ini karena bisa dipakai dari alur chat manual sampai ke jalur usaha yang lebih besar.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah WhatsApp cukup untuk memulai jualan produk digital?", answer: "Cukup untuk tahap awal, selama format order, pencatatan, dan pelayanan dibuat rapi." },
      { question: "Kapan reseller WhatsApp perlu naik level?", answer: "Saat order mulai ramai, produk makin banyak, dan pencatatan manual mulai rawan salah atau memakan waktu." },
    ],
  },
  {
    slug: "cara-membangun-usaha-pulsa-keluarga-dari-rumah",
    title: "Cara Membangun Usaha Pulsa Keluarga dari Rumah",
    description: "Panduan membangun usaha pulsa keluarga dari rumah dengan pembagian peran yang rapi, produk yang mudah dijual, dan layanan yang bisa tumbuh bertahap.",
    excerpt: "Usaha keluarga akan lebih sehat jika semua anggota memahami produk, peran kerja, dan cara melayani pelanggan dengan pola yang sama.",
    category: "Bisnis Agen",
    readTime: "7 menit",
    keywords: ["usaha pulsa keluarga", "bisnis pulsa rumahan", "jualan pulsa keluarga", "usaha digital keluarga"],
    hero: "Usaha keluarga bisa tumbuh kuat jika dimulai dari layanan yang sederhana, pencatatan yang jelas, dan pembagian tugas yang tidak tumpang tindih.",
    products: ["Pulsa", "Paket Data", "E-Wallet", "PPOB"],
    ctaLabel: "Mulai dari Pulsa",
    ctaHref: "/pulsa",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Mulai dari produk yang mudah dipahami semua anggota",
        body: [
          "Pulsa, paket data, e-wallet, dan token listrik adalah kategori awal yang paling aman untuk usaha keluarga karena cepat dipahami dan sering diminta pelanggan.",
          "Jika semua anggota keluarga memahami produk inti ini, pelayanan akan terasa lebih konsisten walaupun order datang di waktu yang berbeda.",
        ],
      },
      {
        heading: "Pentingnya pembagian peran sederhana",
        body: [
          "Satu orang bisa fokus menerima order, satu orang memantau saldo dan transaksi, dan satu orang membantu pencatatan pelanggan rutin. Pola sederhana seperti ini sudah cukup untuk tahap awal.",
          "Tanpa pembagian peran, usaha keluarga mudah kacau saat transaksi mulai ramai atau pelanggan datang bersamaan.",
        ],
      },
      {
        heading: "Arah tumbuh setelah usaha mulai stabil",
        body: [
          "Begitu pelanggan mulai rutin, usaha keluarga bisa menambah kategori rumah tangga, game, atau layanan bulanan lain. Ini membantu transaksi tidak hanya bergantung pada pulsa harian.",
          "Dalam jangka panjang, usaha keluarga juga bisa berkembang ke katalog online atau website sendiri jika operasionalnya sudah lebih siap.",
        ],
      },
    ],
    faqs: [
      { question: "Produk apa yang paling aman untuk usaha keluarga?", answer: "Pulsa, paket data, e-wallet, dan token listrik adalah kombinasi paling aman karena mudah dipahami dan paling sering dicari pelanggan." },
      { question: "Kenapa pembagian peran penting di usaha keluarga?", answer: "Karena pembagian peran mengurangi kebingungan, mempercepat pelayanan, dan membantu pencatatan transaksi tetap rapi." },
    ],
  },
  {
    slug: "cara-membuka-konter-pulsa-untuk-anak-kampus-dan-kos",
    title: "Cara Membuka Konter Pulsa untuk Anak Kampus dan Kos",
    description: "Panduan membuka konter pulsa untuk pasar kampus dan kos dengan produk yang paling sering dicari mahasiswa, anak kos, dan pelanggan harian.",
    excerpt: "Pasar kampus bergerak cepat dan sensitif pada kemudahan. Produk yang tepat dan layanan yang ringkas akan lebih efektif daripada katalog yang terlalu ramai.",
    category: "Bisnis Agen",
    readTime: "6 menit",
    keywords: ["konter pulsa kampus", "jualan pulsa anak kos", "usaha pulsa mahasiswa", "top up area kampus"],
    hero: "Lingkungan kampus dan kos punya ritme transaksi yang cepat, sehingga kategori produk harus disusun untuk kebutuhan yang paling sering muncul.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "DANA", "Mobile Legends", "Paket Telepon Telkomsel"],
    ctaLabel: "Buka Produk untuk Area Kampus",
    ctaHref: "/paket-data",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Produk yang paling sering dicari pasar kampus",
        body: [
          "Paket data, pulsa, e-wallet, dan game biasanya jadi kategori paling aktif di area kampus dan kos. Kebutuhannya cepat, sering berulang, dan mudah dibeli impulsif.",
          "Paket telepon juga tetap relevan untuk sebagian pelanggan yang masih aktif menelepon keluarga atau urusan kerja.",
        ],
      },
      {
        heading: "Kenapa layanan harus ringkas",
        body: [
          "Pelanggan kampus cenderung memilih layanan yang cepat dan tidak bertele-tele. Mereka ingin langsung tahu nominal, harga, dan langkah pembayaran tanpa perlu membaca terlalu banyak.",
          "Karena itu, struktur produk dan respon layanan yang ringkas akan membantu konversi jauh lebih besar.",
        ],
      },
      {
        heading: "Cara memperluas pasar setelah tahap awal",
        body: [
          "Jika pelanggan mahasiswa sudah mulai rutin, konter bisa menambah kategori rumah tangga ringan seperti token listrik atau layanan lain untuk penghuni kos dan keluarga mahasiswa.",
          "Dengan begitu, usaha tidak hanya bergantung pada satu pola transaksi saja dan bisa bertahan lebih stabil.",
        ],
      },
    ],
    faqs: [
      { question: "Produk apa yang paling cocok untuk area kampus?", answer: "Paket data operator besar, pulsa, top up DANA atau GoPay, dan game adalah kombinasi paling kuat untuk area kampus dan kos." },
      { question: "Kenapa layanan ringkas penting untuk pasar mahasiswa?", answer: "Karena pelanggan kampus bergerak cepat dan biasanya ingin langsung melihat nominal, harga, serta langkah bayar tanpa harus membaca terlalu banyak." },
    ],
  },
  {
    slug: "cara-memanfaatkan-promo-operator-tanpa-bikin-katalog-acak",
    title: "Cara Memanfaatkan Promo Operator tanpa Bikin Katalog Acak",
    description: "Panduan memanfaatkan promo operator sambil tetap menjaga katalog produk rapi, mudah dipahami, dan nyaman dipakai pelanggan Isiloka.",
    excerpt: "Promo operator bisa menarik pembeli, tetapi kalau penempatannya berantakan justru membuat katalog terasa penuh dan membingungkan.",
    category: "Operasional",
    readTime: "6 menit",
    keywords: ["promo operator", "katalog promo pulsa", "promo paket data", "susun promo operator"],
    hero: "Promo yang baik harus membantu pembeli mengambil keputusan lebih cepat, bukan membuat pilihan produk terasa campur aduk.",
    products: ["Pulsa Telkomsel", "Paket Data Telkomsel", "Paket Telepon Telkomsel", "Pulsa Data"],
    ctaLabel: "Lihat Kategori Promo Operator",
    ctaHref: "/paket-data",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Kenapa promo perlu ditata dengan benar",
        body: [
          "Promo memang menarik, tetapi jika dicampur tanpa struktur yang jelas, pembeli sulit membedakan mana produk reguler dan mana produk khusus. Ini membuat katalog terasa penuh dan melelahkan.",
          "Katalog yang rapi akan membantu promo bekerja lebih efektif karena pembeli bisa langsung melihat nilai tambahnya.",
        ],
      },
      {
        heading: "Pisahkan promo tanpa memecah alur belanja",
        body: [
          "Produk promo sebaiknya tetap berada di dalam grup yang mudah dipahami, bukan dilempar ke kategori acak. Dengan begitu, pelanggan tetap bisa membandingkan pilihan tanpa kehilangan konteks.",
          "Judul produk promo juga harus ringkas dan tetap menunjukkan nominal, kuota, atau masa aktif sebagai informasi utama.",
        ],
      },
      {
        heading: "Dampaknya ke kepercayaan pelanggan",
        body: [
          "Pelanggan lebih nyaman membeli jika bisa melihat bahwa promo ditempatkan secara wajar dan tidak membuat halaman terasa kacau. Ini memberi kesan bahwa layanan Anda tertata dan profesional.",
          "Untuk member, penataan promo yang rapi juga memudahkan penjelasan saat pelanggan bertanya lewat chat atau datang langsung ke warung.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah promo operator harus dipisah dari produk reguler?", answer: "Ya, tetapi tetap di dalam struktur grup yang mudah dipahami agar pembeli tidak bingung saat membandingkan pilihan." },
      { question: "Kenapa katalog promo bisa terasa kacau?", answer: "Karena terlalu banyak label dan variasi yang dicampur tanpa urutan yang jelas, sehingga pembeli sulit menemukan informasi inti produk." },
    ],
  },
  {
    slug: "cara-menyiapkan-modal-awal-dan-deposit-untuk-jualan-produk-digital",
    title: "Cara Menyiapkan Modal Awal dan Deposit untuk Jualan Produk Digital",
    description: "Panduan menyiapkan modal awal dan deposit agar usaha produk digital bisa mulai jalan dengan aman tanpa membuat operasional cepat seret.",
    excerpt: "Modal awal bukan soal harus besar, tetapi soal cukup untuk menjaga transaksi berjalan dan tidak berhenti saat pelanggan mulai datang.",
    category: "Deposit",
    readTime: "6 menit",
    keywords: ["modal awal pulsa", "deposit agen pulsa", "modal jualan produk digital", "top up saldo awal"],
    hero: "Usaha digital yang sehat dimulai dari saldo yang cukup, pengelolaan deposit yang rapi, dan pemahaman produk yang dijual.",
    products: ["Saldo Member", "Deposit QRIS", "Transfer Bank", "Pulsa Telkomsel", "Paket Data Telkomsel"],
    ctaLabel: "Masuk untuk Top Up Saldo",
    ctaHref: "/login",
    publishedAt,
    updatedAt,
    sections: [
      {
        heading: "Berapa besar modal awal yang masuk akal",
        body: [
          "Modal awal tidak harus besar, tetapi harus cukup untuk menutup transaksi dasar yang paling sering muncul seperti pulsa, paket data, e-wallet, dan token listrik.",
          "Yang penting adalah modal tidak terlalu kecil sampai membuat transaksi sering tertahan ketika pelanggan mulai datang berturut-turut.",
        ],
      },
      {
        heading: "Kenapa deposit harus dikelola sejak awal",
        body: [
          "Banyak usaha kecil tersendat bukan karena kurang pelanggan, tetapi karena saldo cepat habis dan tidak segera dipantau. Deposit yang rapi membantu pelayanan tetap berjalan lancar.",
          "Dengan pola top up yang jelas, member bisa menyesuaikan ritme deposit dengan volume transaksi nyata, bukan sekadar perkiraan kasar.",
        ],
      },
      {
        heading: "Hubungannya dengan pertumbuhan usaha",
        body: [
          "Saat usaha mulai ramai, kebutuhan saldo akan ikut naik. Karena itu, kebiasaan mengelola deposit sejak awal sangat penting untuk menyiapkan tahap usaha yang lebih besar.",
          "Isiloka mendukung pola ini karena member bisa memulai dari usaha kecil lalu menyesuaikan modal, produk, dan ritme transaksi secara bertahap.",
        ],
      },
    ],
    faqs: [
      { question: "Apakah modal awal harus besar untuk mulai jualan produk digital?", answer: "Tidak. Yang penting cukup untuk menutup transaksi dasar seperti pulsa, paket data, e-wallet, dan token listrik tanpa terlalu sering kehabisan saldo." },
      { question: "Kenapa deposit harus dipantau sejak awal?", answer: "Karena saldo yang tidak dipantau akan cepat menghambat pelayanan, terutama saat beberapa order masuk berurutan dan pelanggan ingin diproses cepat." },
    ],
  },
];

export function getSeoArticle(slug: string) {
  return seoArticles.find((article) => article.slug === slug);
}

export function getSeoArticleUrl(slug: string) {
  return `${SITE_URL}/artikel/${slug}`;
}

export function getCanonicalArticleUrl(slug: string) {
  return `${CANONICAL_SITE_URL}/artikel/${slug}`;
}

export function getCanonicalUrl(path = "") {
  return `${CANONICAL_SITE_URL}${path}`;
}

const productQuickLinkMap: Record<string, QuickLink> = {
  Pulsa: { label: "Semua Pulsa", href: "/pulsa" },
  Listrik: { label: "Kategori Listrik", href: "/listrik" },
  "Paket Data": { label: "Semua Paket Data", href: "/paket-data" },
  "Pulsa Data": { label: "Pulsa Data", href: "/pulsa-data" },
  "E-Wallet": { label: "Semua E-Wallet", href: "/ewallet" },
  Game: { label: "Semua Game", href: "/game" },
  PPOB: { label: "Layanan Rumah Tangga", href: "/listrik" },
  PGN: { label: "Bayar PGN", href: "/pgn" },
  Retail: { label: "Panduan Mulai Jualan", href: "/artikel/cara-jualan-pulsa-dari-rumah-untuk-pemula" },
  Agen: { label: "Panduan Menjadi Agen", href: "/artikel/cara-menjadi-agen-pulsa-di-pulsakilat" },
  H2H: { label: "Panduan Member H2H", href: "/artikel/cara-menjadi-member-h2h-pulsakilat" },
  "Website Jualan": { label: "Panduan Website Jualan", href: "/artikel/cara-membuat-website-jualan-pulsa-sendiri" },
  "Saldo Member": { label: "Panduan Top Up Saldo", href: "/artikel/cara-top-up-saldo-member-dengan-qris-dan-bank" },
  "API H2H": { label: "Dokumentasi H2H", href: "/docs" },
  API: { label: "Dokumentasi H2H", href: "/docs" },
  "Pulsa Telkomsel": { label: "Pulsa Telkomsel", href: "/pulsa/telkomsel" },
  "Pulsa Indosat": { label: "Pulsa Indosat", href: "/pulsa/indosat" },
  "Pulsa XL": { label: "Pulsa XL", href: "/pulsa/xl" },
  "Pulsa Smartfren": { label: "Pulsa Smartfren", href: "/pulsa/smartfren" },
  "Paket Data Telkomsel": { label: "Paket Data Telkomsel", href: "/paket-data/telkomsel" },
  "Paket Data Indosat": { label: "Paket Data Indosat", href: "/paket-data/indosat" },
  "Paket Data XL": { label: "Paket Data XL", href: "/paket-data/xl" },
  "Paket Data by.U": { label: "Paket Data by.U", href: "/paket-data/byu" },
  DANA: { label: "Top Up Dana", href: "/ewallet/dana" },
  OVO: { label: "Top Up OVO", href: "/ewallet/ovo" },
  GoPay: { label: "Top Up GoPay", href: "/ewallet/gopay" },
  LinkAja: { label: "Top Up LinkAja", href: "/ewallet/linkaja" },
  ShopeePay: { label: "Top Up ShopeePay", href: "/ewallet/shopeepay" },
  "Mobile Legends": { label: "Top Up Mobile Legends", href: "/game/mobilelegends" },
  "Free Fire": { label: "Top Up Free Fire", href: "/game/freefire" },
  PUBG: { label: "Top Up PUBG", href: "/game/pubgm" },
  "Top up game lainnya": { label: "Kategori Game", href: "/game" },
  "Token Listrik Promo": { label: "Token Listrik", href: "/listrik/token" },
  "Token Listrik Standar Reply": { label: "Token Listrik", href: "/listrik/token" },
  "Token Listrik Full Reply": { label: "Token Listrik", href: "/listrik/token" },
  "Tagihan Listrik": { label: "Tagihan Listrik", href: "/listrik/tagihan" },
  "PLN Pascabayar": { label: "Tagihan Listrik", href: "/listrik/tagihan" },
  "BPJS Kesehatan": { label: "BPJS Kesehatan", href: "/bpjs/kesehatan" },
  "BPJS Ketenagakerjaan": { label: "BPJS Ketenagakerjaan", href: "/bpjs/ketenagakerjaan" },
  PDAM: { label: "Tagihan PDAM", href: "/pdam" },
  "Internet Pascabayar": { label: "Internet Pascabayar", href: "/internet-pascabayar" },
  "HP Pascabayar": { label: "HP Pascabayar", href: "/hp-pascabayar" },
  TV: { label: "TV Berlangganan", href: "/tv" },
  "Voucher TV": { label: "TV Berlangganan", href: "/tv" },
  "TV Berlangganan": { label: "TV Berlangganan", href: "/tv" },
  "Masa Aktif": { label: "Masa Aktif", href: "/masa-aktif" },
  "Aktivasi Perdana": { label: "Aktivasi Perdana", href: "/aktivasi-perdana" },
  "Deposit QRIS": { label: "Panduan Top Up Saldo", href: "/artikel/cara-top-up-saldo-member-dengan-qris-dan-bank" },
  "Transfer Bank": { label: "Panduan Top Up Saldo", href: "/artikel/cara-top-up-saldo-member-dengan-qris-dan-bank" },
};

export function getArticleProductLinks(article: SeoArticle): QuickLink[] {
  return article.products
    .map((product) => productQuickLinkMap[product])
    .filter(Boolean)
    .filter((item, index, all) => all.findIndex((entry) => entry.href === item.href) === index);
}

const relatedArticleMap: Record<string, string[]> = {
  "cara-isi-pulsa-online-semua-operator": [
    "cara-beli-paket-data-harian-dan-bulanan",
    "cara-top-up-ewallet-dana-ovo-gopay-linkaja",
    "cara-jualan-pulsa-dari-rumah-untuk-pemula",
  ],
  "cara-beli-paket-data-harian-dan-bulanan": [
    "cara-isi-pulsa-online-semua-operator",
    "cara-beli-paket-telepon-semua-operator",
    "cara-membuka-konter-pulsa-untuk-anak-kampus-dan-kos",
  ],
  "cara-top-up-ewallet-dana-ovo-gopay-linkaja": [
    "cara-isi-pulsa-online-semua-operator",
    "cara-top-up-game-mobile-legends-free-fire-pubg",
    "produk-pulsakilat-yang-paling-cocok-untuk-calon-member",
  ],
  "cara-top-up-game-mobile-legends-free-fire-pubg": [
    "cara-top-up-ewallet-dana-ovo-gopay-linkaja",
    "cara-beli-paket-data-harian-dan-bulanan",
    "cara-menjual-top-up-game-untuk-komunitas-dan-anak-muda",
  ],
  "cara-beli-token-listrik-pln-online": [
    "cara-bayar-tagihan-listrik-online",
    "cara-bayar-pdam-online",
    "cara-bayar-bpjs-kesehatan-online",
  ],
  "cara-bayar-tagihan-listrik-online": [
    "cara-beli-token-listrik-pln-online",
    "cara-bayar-pdam-online",
    "cara-bayar-internet-pascabayar-online",
  ],
  "cara-bayar-bpjs-kesehatan-online": [
    "cara-bayar-tagihan-listrik-online",
    "cara-bayar-pdam-online",
    "cara-menjadi-agen-pulsa-di-pulsakilat",
  ],
  "cara-bayar-pdam-online": [
    "cara-bayar-tagihan-listrik-online",
    "cara-bayar-bpjs-kesehatan-online",
    "cara-bayar-internet-pascabayar-online",
  ],
  "cara-bayar-internet-pascabayar-online": [
    "cara-bayar-tagihan-listrik-online",
    "cara-bayar-pdam-online",
    "cara-beli-voucher-tv-digital-dan-berlangganan",
  ],
  "cara-bayar-hp-pascabayar-online": [
    "cara-bayar-internet-pascabayar-online",
    "cara-bayar-tagihan-listrik-online",
    "cara-menjadi-agen-pulsa-di-pulsakilat",
  ],
  "cara-beli-voucher-tv-digital-dan-berlangganan": [
    "cara-bayar-internet-pascabayar-online",
    "cara-bayar-tagihan-listrik-online",
    "cara-bayar-pdam-online",
  ],
  "cara-perpanjang-masa-aktif-nomor-hp": [
    "cara-isi-pulsa-online-semua-operator",
    "cara-beli-paket-data-harian-dan-bulanan",
    "cara-aktivasi-perdana-dan-topup-awal",
  ],
  "cara-aktivasi-perdana-dan-topup-awal": [
    "cara-perpanjang-masa-aktif-nomor-hp",
    "cara-isi-pulsa-online-semua-operator",
    "cara-beli-paket-data-harian-dan-bulanan",
  ],
  "cara-jualan-pulsa-dari-rumah-untuk-pemula": [
    "cara-menjadi-agen-pulsa-di-pulsakilat",
    "cara-jadi-reseller-produk-digital-lewat-whatsapp",
    "cara-membangun-warung-pulsa-dan-ppob-yang-lengkap",
  ],
  "cara-menjadi-agen-pulsa-di-pulsakilat": [
    "cara-jualan-pulsa-dari-rumah-untuk-pemula",
    "cara-membangun-usaha-pulsa-keluarga-dari-rumah",
    "cara-membangun-warung-pulsa-dan-ppob-yang-lengkap",
  ],
  "cara-menjadi-member-h2h-pulsakilat": [
    "perbedaan-member-retail-agen-dan-h2h-di-pulsakilat",
    "cara-bertumbuh-dari-retail-ke-agen-dan-h2h-bersama-pulsakilat",
    "cara-membuat-website-jualan-pulsa-sendiri",
  ],
  "cara-membuat-website-jualan-pulsa-sendiri": [
    "cara-menjadi-agen-pulsa-di-pulsakilat",
    "cara-menjadi-member-h2h-pulsakilat",
    "cara-membangun-panel-reseller-dengan-h2h-pulsakilat",
  ],
  "keuntungan-berlangganan-dan-bertumbuh-bersama-pulsakilat": [
    "produk-pulsakilat-yang-paling-cocok-untuk-calon-member",
    "cara-bertumbuh-dari-retail-ke-agen-dan-h2h-bersama-pulsakilat",
    "cara-top-up-saldo-member-dengan-qris-dan-bank",
  ],
  "cara-mengelola-markup-dan-profit-produk-digital": [
    "cara-jualan-pulsa-dari-rumah-untuk-pemula",
    "cara-top-up-saldo-member-dengan-qris-dan-bank",
    "cara-rekonsiliasi-transaksi-produk-digital",
  ],
  "cara-rekonsiliasi-transaksi-produk-digital": [
    "cara-mengelola-markup-dan-profit-produk-digital",
    "cara-top-up-saldo-member-dengan-qris-dan-bank",
    "cara-mengelola-layanan-pelanggan-usaha-produk-digital",
  ],
  "cara-beli-paket-telepon-semua-operator": [
    "cara-beli-paket-data-harian-dan-bulanan",
    "cara-isi-pulsa-online-semua-operator",
    "cara-perpanjang-masa-aktif-nomor-hp",
  ],
  "cara-beli-pulsa-data-dan-kombinasi-pulsa-kuota": [
    "cara-isi-pulsa-online-semua-operator",
    "cara-beli-paket-data-harian-dan-bulanan",
    "cara-beli-paket-telepon-semua-operator",
  ],
  "cara-bayar-pgn-online-dan-layanan-energi-rumah-tangga": [
    "cara-bayar-tagihan-listrik-online",
    "cara-bayar-pdam-online",
    "cara-bayar-internet-pascabayar-online",
  ],
  "cara-top-up-saldo-member-dengan-qris-dan-bank": [
    "cara-rekonsiliasi-transaksi-produk-digital",
    "cara-mengelola-markup-dan-profit-produk-digital",
    "perbedaan-member-retail-agen-dan-h2h-di-pulsakilat",
  ],
  "perbedaan-member-retail-agen-dan-h2h-di-pulsakilat": [
    "cara-menjadi-agen-pulsa-di-pulsakilat",
    "cara-menjadi-member-h2h-pulsakilat",
    "cara-bertumbuh-dari-retail-ke-agen-dan-h2h-bersama-pulsakilat",
  ],
  "cara-bertumbuh-dari-retail-ke-agen-dan-h2h-bersama-pulsakilat": [
    "perbedaan-member-retail-agen-dan-h2h-di-pulsakilat",
    "keuntungan-berlangganan-dan-bertumbuh-bersama-pulsakilat",
    "cara-menjadi-member-h2h-pulsakilat",
  ],
  "produk-pulsakilat-yang-paling-cocok-untuk-calon-member": [
    "cara-membangun-warung-pulsa-dan-ppob-yang-lengkap",
    "cara-top-up-saldo-member-dengan-qris-dan-bank",
    "perbedaan-member-retail-agen-dan-h2h-di-pulsakilat",
  ],
  "cara-membangun-warung-pulsa-dan-ppob-yang-lengkap": [
    "cara-jualan-pulsa-dari-rumah-untuk-pemula",
    "produk-pulsakilat-yang-paling-cocok-untuk-calon-member",
    "cara-mengelola-pelanggan-rutin-token-listrik-bpjs-dan-pdam",
  ],
  "cara-menyusun-katalog-produk-digital-yang-mudah-dipahami": [
    "cara-beli-paket-data-harian-dan-bulanan",
    "cara-memanfaatkan-promo-operator-tanpa-bikin-katalog-acak",
    "cara-mengelola-layanan-pelanggan-usaha-produk-digital",
  ],
  "cara-mengelola-pelanggan-rutin-token-listrik-bpjs-dan-pdam": [
    "cara-bayar-tagihan-listrik-online",
    "cara-bayar-bpjs-kesehatan-online",
    "cara-bayar-pdam-online",
  ],
  "cara-membangun-panel-reseller-dengan-h2h-pulsakilat": [
    "cara-menjadi-member-h2h-pulsakilat",
    "cara-membuat-website-jualan-pulsa-sendiri",
    "cara-rekonsiliasi-transaksi-produk-digital",
  ],
  "cara-menjual-top-up-game-untuk-komunitas-dan-anak-muda": [
    "cara-top-up-game-mobile-legends-free-fire-pubg",
    "cara-top-up-ewallet-dana-ovo-gopay-linkaja",
    "cara-beli-paket-data-harian-dan-bulanan",
  ],
  "cara-mengelola-layanan-pelanggan-usaha-produk-digital": [
    "cara-rekonsiliasi-transaksi-produk-digital",
    "cara-mengelola-markup-dan-profit-produk-digital",
    "cara-jualan-pulsa-dari-rumah-untuk-pemula",
  ],
  "cara-membangun-warung-pulsa-desa-yang-tetap-ramai": [
    "cara-membangun-warung-pulsa-dan-ppob-yang-lengkap",
    "cara-mengelola-pelanggan-rutin-token-listrik-bpjs-dan-pdam",
    "cara-menyiapkan-modal-awal-dan-deposit-untuk-jualan-produk-digital",
  ],
  "cara-jadi-reseller-produk-digital-lewat-whatsapp": [
    "cara-jualan-pulsa-dari-rumah-untuk-pemula",
    "cara-menjadi-agen-pulsa-di-pulsakilat",
    "cara-menyiapkan-modal-awal-dan-deposit-untuk-jualan-produk-digital",
  ],
  "cara-membangun-usaha-pulsa-keluarga-dari-rumah": [
    "cara-jualan-pulsa-dari-rumah-untuk-pemula",
    "cara-menjadi-agen-pulsa-di-pulsakilat",
    "cara-menyiapkan-modal-awal-dan-deposit-untuk-jualan-produk-digital",
  ],
  "cara-membuka-konter-pulsa-untuk-anak-kampus-dan-kos": [
    "cara-beli-paket-data-harian-dan-bulanan",
    "cara-menjual-top-up-game-untuk-komunitas-dan-anak-muda",
    "cara-top-up-ewallet-dana-ovo-gopay-linkaja",
  ],
  "cara-memanfaatkan-promo-operator-tanpa-bikin-katalog-acak": [
    "cara-menyusun-katalog-produk-digital-yang-mudah-dipahami",
    "cara-beli-paket-data-harian-dan-bulanan",
    "cara-beli-pulsa-data-dan-kombinasi-pulsa-kuota",
  ],
  "cara-menyiapkan-modal-awal-dan-deposit-untuk-jualan-produk-digital": [
    "cara-top-up-saldo-member-dengan-qris-dan-bank",
    "cara-jualan-pulsa-dari-rumah-untuk-pemula",
    "cara-mengelola-markup-dan-profit-produk-digital",
  ],
};

export function getRelatedArticles(slug: string, limit = 3): SeoArticle[] {
  const current = getSeoArticle(slug);
  if (!current) return [];

  const curated = relatedArticleMap[slug]
    ?.map((relatedSlug) => getSeoArticle(relatedSlug))
    .filter(Boolean) as SeoArticle[] | undefined;
  if (curated && curated.length > 0) {
    return curated.slice(0, limit);
  }

  return seoArticles
    .filter((article) => article.slug !== slug)
    .map((article) => {
      const sharedProducts = article.products.filter((product) => current.products.includes(product)).length;
      const sharedCategory = article.category === current.category ? 1 : 0;
      return { article, score: sharedProducts * 2 + sharedCategory };
    })
    .sort((left, right) => right.score - left.score || left.article.title.localeCompare(right.article.title))
    .slice(0, limit)
    .map((item) => item.article);
}
