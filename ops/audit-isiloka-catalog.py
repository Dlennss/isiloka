"""Read-only comparison of the app catalogue against a captured H2HR response."""
import json
import sys
import urllib.request

with open(sys.argv[1], encoding="utf-8-sig") as source:
    upstream = json.load(source)
assert upstream.get("ok") is True
base = sys.argv[2].rstrip("/") if len(sys.argv) > 2 else "http://127.0.0.1:8102"

def fetch(path):
    with urllib.request.urlopen(base + path, timeout=60) as response:
        result = json.load(response)
    assert result.get("ok") is True, path
    return result["items"]

expected = {p["sku"]: p for p in upstream["items"] if p.get("aktif", True)}
products = fetch("/v1/app/produk")
actual = {p["sku"]: p for p in products}
assert len(actual) == len(products), "Duplicate app SKU"
assert set(actual) == set(expected), "Missing or foreign app SKUs"
for sku, source in expected.items():
    item = actual[sku]
    for field in ("nama", "group_name", "kategori_nama", "brand_nama", "tipe_harga", "maksimal_nominal"):
        assert item.get(field) == source.get(field), (sku, field)
    price = source.get("harga_dasar_app")
    if price is None:
        price = source.get("harga") if source["tipe_harga"] == "FIXED" else source.get("fee_tambahan")
    assert price is not None and item["harga_dasar_app"] == price, (sku, "price")
    assert item["aktif"] is True
    for fee in ("fee_master", "fee_agent", "fee_user", "fee_guest"):
        assert item[fee] == 0, (sku, fee)
categories = fetch("/v1/app/kategori")
brands = fetch("/v1/app/brand")
assert {p["nama"] for p in categories} == {p["kategori_nama"] for p in expected.values()}
assert {p["nama"] for p in brands} == {p["brand_nama"] for p in expected.values()}
print(json.dumps({"verified_products": len(products), "categories": len(categories), "brands": len(brands), "result": "all fields and base prices match"}))
