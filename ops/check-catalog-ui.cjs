const assert = require('node:assert/strict');
const path = require('node:path');
const os = require('node:os');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const base = process.env.CATALOG_UI_URL || 'http://localhost:3110';
  const api = process.env.CATALOG_API_URL || 'https://isiloka.com';
  const categories = (await (await fetch(`${api}/v1/app/kategori`)).json()).items;
  const wallet = categories.find((item) => item.nama === 'E-Wallet');
  const brands = (await (await fetch(`${api}/v1/app/brand?kategori_id=${wallet.id}`)).json()).items;
  const dana = brands.find((item) => item.nama === 'DANA');
  const danaProducts = (await (await fetch(`${api}/v1/app/produk?kategori_id=${wallet.id}&brand_id=${dana.id}`)).json()).items;
  const danaSKU = danaProducts.some((item) => item.sku === 'PPOBDANA') ? 'PPOBDANA' : 'DANA';
  const browser = await chromium.launch({ channel: 'chrome', headless: true });
  try {
    for (const width of [390, 1280, 320]) {
      const page = await browser.newPage({ viewport: { width, height: 900 } });
      const errors = [];
      page.on('pageerror', (error) => errors.push(error.message));
      await page.goto(`${base}/kategori`, { waitUntil: 'networkidle', timeout: 120000 });
      await page.getByRole('heading', { name: 'Semua Layanan', exact: true }).waitFor();
      for (const item of categories) assert.equal(await page.getByRole('link', { name: item.nama, exact: true }).count(), 1, item.nama);
      assert.equal(await page.getByRole('link', { name: 'eSIM & Roaming', exact: true }).count(), 0);
      await page.getByRole('textbox', { name: 'Cari layanan' }).fill('asuransi');
      assert.equal(await page.getByRole('link', { name: 'Asuransi', exact: true }).count(), 1);
      await page.getByRole('textbox', { name: 'Cari layanan' }).fill('');
      await page.screenshot({ path: path.join(os.tmpdir(), `isiloka-catalog-${width}.png`), fullPage: true });
      await page.getByRole('link', { name: 'E-Wallet', exact: true }).click();
      await page.getByRole('link', { name: 'DANA', exact: true }).waitFor();
      for (const brand of brands) assert.equal(await page.getByRole('link', { name: brand.nama, exact: true }).count(), 1, brand.nama);
      const broken = await page.locator('img').evaluateAll(async (images) => {
        for (const image of images) { image.loading = 'eager'; await image.decode().catch(() => {}); }
        return images.filter((image) => !image.complete || image.naturalWidth === 0).map((image) => image.src);
      });
      assert.deepEqual(broken, []);
      await page.screenshot({ path: path.join(os.tmpdir(), `isiloka-providers-${width}.png`), fullPage: true });
      await page.goto(`${base}/kategori/${wallet.id}/brand/${dana.id}`, { waitUntil: 'networkidle', timeout: 120000 });
      await page.getByRole('textbox', { name: 'Cari nama atau SKU' }).fill('DANA');
      const buy = page.getByRole('button', { name: `Top Up ${danaSKU}`, exact: true });
      await buy.waitFor();
      assert.match(await buy.locator('..').innerText(), /Biaya tambahan\s+Rp 1\.000/);
      const cards = page.locator('article');
      const first = await cards.nth(0).boundingBox();
      assert.equal(await cards.first().locator('..').evaluate((grid) => getComputedStyle(grid).gridTemplateColumns.split(' ').length), 2);
      if (await cards.count() > 1) {
        const second = await cards.nth(1).boundingBox();
        assert.equal(first.y, second.y, 'wallet cards must share a two-column row');
        assert.ok(second.x >= first.x + first.width, 'wallet cards must not overlap');
        assert.ok(Math.abs(first.height - second.height) < 1, 'wallet cards must have equal heights');
      }
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false, `overflow ${width}`);
      await page.screenshot({ path: path.join(os.tmpdir(), `isiloka-products-${width}.png`), fullPage: true });
      await buy.click();
      await page.getByText('Checkout', { exact: true }).waitFor({ timeout: 10000 });
      await page.getByPlaceholder('Contoh: 10000').fill('10000');
      await page.getByText('Rp 11.077', { exact: true }).waitFor();
      await page.screenshot({ path: path.join(os.tmpdir(), `isiloka-checkout-${width}.png`) });
      await page.goto(`${base}/ewallet/gopay`, { waitUntil: 'networkidle', timeout: 120000 });
      const legacyCards = page.locator('article');
      await legacyCards.first().waitFor();
      const legacyFirst = await legacyCards.nth(0).boundingBox();
      assert.equal(await legacyCards.first().locator('..').evaluate((grid) => getComputedStyle(grid).gridTemplateColumns.split(' ').length), 2);
      if (await legacyCards.count() > 1) {
        const legacySecond = await legacyCards.nth(1).boundingBox();
        assert.equal(legacyFirst.y, legacySecond.y, 'legacy wallet cards must also use two columns');
        assert.ok(Math.abs(legacyFirst.height - legacySecond.height) < 1);
      }
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
      await legacyCards.first().scrollIntoViewIfNeeded();
      await page.screenshot({ path: path.join(os.tmpdir(), `isiloka-ewallet-legacy-${width}.png`) });
      await legacyCards.first().getByRole('button', { name: /Top Up/ }).click();
      await page.getByText('Checkout', { exact: true }).waitFor();
      assert.deepEqual(errors, []);
      console.log(JSON.stringify({ width, categories: categories.length, providerImages: 'loaded', search: 'passed', legacyWalletCards: 'passed', checkoutOpenedWithoutSubmitting: true }));
      await page.close();
    }
  } finally {
    await browser.close();
  }
})().catch((error) => { console.error(error); process.exitCode = 1; });
