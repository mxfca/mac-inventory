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

func TestMarkdownPartialCollection(t *testing.T) {
	collectedAt := time.Date(2026, 11, 15, 12, 0, 3, 0, time.FixedZone("BRT", -3*60*60))
	head := `---
atualizado: 2026-11-15
por: mac-inventory
status: ativo
tags: [inventario]
coletado_em: 2026-11-15T12:00:03-03:00
coleta: parcial
`
	tail := "---\n\n<!-- Gerado pela coleta semanal do mac-inventory. Não editar à mão. -->\n"
	table := "| Nome | Versão |\n| --- | --- |\n"

	tests := []struct {
		name string
		inv  inventory.Inventory
		want string
	}{
		{
			"brew fails",
			inventory.Inventory{
				Other:       []inventory.Item{{Name: "Zoom", Version: "6.6.0"}},
				BrewFailure: "exit status 1",
			},
			head + "falhas: [\"brew: exit status 1\"]\n" + tail +
				"\n## brew\n\n_Coleta falhou: exit status 1._\n" +
				"\n## App Store\n\n_Nenhum aplicativo._\n" +
				"\n## outro\n\n" + table + "| Zoom | 6.6.0 |\n",
		},
		{
			"system_profiler fails",
			inventory.Inventory{
				Brew:        []inventory.Item{{Name: "gh", Version: "2.103.3"}},
				AppsFailure: "invalid JSON",
			},
			head + "falhas: [\"system_profiler: invalid JSON\"]\n" + tail +
				"\n## brew\n\n" + table + "| gh | 2.103.3 |\n" +
				"\n## App Store\n\n_Coleta falhou: invalid JSON._\n" +
				"\n## outro\n\n_Coleta falhou: invalid JSON._\n",
		},
		{
			"both fail",
			inventory.Inventory{BrewFailure: "exit status 1", AppsFailure: "invalid JSON"},
			head + "falhas: [\"brew: exit status 1\", \"system_profiler: invalid JSON\"]\n" + tail +
				"\n## brew\n\n_Coleta falhou: exit status 1._\n" +
				"\n## App Store\n\n_Coleta falhou: invalid JSON._\n" +
				"\n## outro\n\n_Coleta falhou: invalid JSON._\n",
		},
		{
			"cause with quote and backslash",
			inventory.Inventory{BrewFailure: `bad "json" \ path`},
			head + `falhas: ["brew: bad \"json\" \\ path"]` + "\n" + tail +
				"\n## brew\n\n_Coleta falhou: bad \"json\" \\ path._\n" +
				"\n## App Store\n\n_Nenhum aplicativo._\n" +
				"\n## outro\n\n_Nenhum aplicativo._\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(render.Markdown(tt.inv, collectedAt)); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
