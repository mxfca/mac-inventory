package render_test

import (
	"testing"
	"time"

	"github.com/mxfca/mac-inventory/internal/inventory"
	"github.com/mxfca/mac-inventory/internal/render"
)

func TestMarkdown(t *testing.T) {
	collectedAt := time.Date(2026, 11, 15, 12, 0, 3, 0, time.FixedZone("BRT", -3*60*60))
	inv := inventory.Inventory{
		AppStore: []inventory.Item{{Name: "Telegram", Version: "13.1.0"}},
		Other: []inventory.Item{
			{Name: "Pipe | App", Version: "1|2"},
			{Name: "Zoom"},
		},
	}

	got := string(render.Markdown(inv, collectedAt))

	want := `---
atualizado: 2026-11-15
por: mac-inventory
status: ativo
tags: [inventario]
coletado_em: 2026-11-15T12:00:03-03:00
coleta: completa
falhas: []
---

<!-- Gerado pela coleta semanal do mac-inventory. Não editar à mão. -->

## brew

_Nenhum aplicativo._

## App Store

| Nome | Versão |
| --- | --- |
| Telegram | 13.1.0 |

## outro

| Nome | Versão |
| --- | --- |
| Pipe \| App | 1\|2 |
| Zoom | desconhecida |
`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
