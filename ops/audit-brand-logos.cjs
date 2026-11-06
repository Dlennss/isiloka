const fs = require('node:fs');
const path = require('node:path');
const Module = require('node:module');
const ts = require('../pulsa-fe/node_modules/typescript');

const filename = path.resolve(__dirname, '../pulsa-fe/lib/brand-logos.ts');
const compiled = ts.transpileModule(fs.readFileSync(filename, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } });
const mapping = new Module(filename, module);
mapping._compile(compiled.outputText, filename);
const source = JSON.parse(fs.readFileSync(process.argv[2], 'utf8').replace(/^\uFEFF/, ''));
const brands = [...new Set(source.items.map((item) => item.brand_nama))].sort();
const entries = brands.map((brand) => {
  const logo = mapping.exports.getBrandLogo(brand);
  return { brand, ...logo, exists: logo ? fs.existsSync(path.join(__dirname, '../pulsa-fe/public', logo.src)) : false };
});
const missing = entries.filter((entry) => !entry.exists);
const report = { totalBrands: brands.length, mappedAssets: entries.length - missing.length, missing: missing.map((entry) => entry.brand), entries };
if (process.argv[3]) fs.writeFileSync(process.argv[3], JSON.stringify(report, null, 2) + '\n');
console.log(JSON.stringify({ totalBrands: report.totalBrands, mappedAssets: report.mappedAssets, missing: report.missing }));
