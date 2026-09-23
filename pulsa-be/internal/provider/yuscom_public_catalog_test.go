package provider

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFindYuscomCatalogFrame(t *testing.T) {
	got, err := findYuscomCatalogFrame(strings.NewReader(`<html><iframe src="https://report.yuscom.co.id/harga.js.php?id=test"></iframe></html>`), DefaultYuscomPublicCatalogURL)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://report.yuscom.co.id/harga.js.php?id=test" {
		t.Fatalf("unexpected frame URL: %s", got)
	}
}

func TestParseOpenYuscomProductCodes(t *testing.T) {
	input := `<table>
		<tr class="td2"><tr class="td2"><td>UDGD10</td><td>Gopay Driver 10.000</td><td>10.450</td><td><span>Open</span></td></tr>
		<tr><td>OFF1</td><td>Produk Gangguan</td><td>5.000</td><td><span>Gangguan</span></td></tr>
	</table>`
	codes, err := parseOpenYuscomProductCodes(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := codes["UDGD10"]; !ok {
		t.Fatal("produk Open tidak terbaca")
	}
	if _, ok := codes["OFF1"]; ok {
		t.Fatal("produk Gangguan tidak boleh aktif")
	}
}

func TestYuscomPublicCatalogLive(t *testing.T) {
	if os.Getenv("YUSCOM_LIVE_TEST") != "1" {
		t.Skip("set YUSCOM_LIVE_TEST=1 untuk menguji halaman resmi Yuscom")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	codes, err := NewYuscomPublicCatalog("", 45*time.Second).OpenProductCodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) < 1000 {
		t.Fatalf("jumlah produk Open Yuscom tidak wajar: %d", len(codes))
	}
	if _, ok := codes["UDGD10"]; !ok {
		t.Fatal("produk Yuscom UDGD10 tidak terbaca")
	}
}
