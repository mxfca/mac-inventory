// Package render turns an inventory into the Markdown of the inventory file.
package render

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/mxfca/mac-inventory/internal/inventory"
)

// Markdown returns the inventory file: frontmatter plus one section per source.
func Markdown(inv inventory.Inventory, collectedAt time.Time) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, `---
atualizado: %s
por: mac-inventory
status: ativo
tags: [inventario]
coletado_em: %s
coleta: completa
falhas: []
---

<!-- Gerado pela coleta semanal do mac-inventory. Não editar à mão. -->
`, collectedAt.Format(time.DateOnly), collectedAt.Format(time.RFC3339))

	section(&b, "brew", inv.Brew)
	section(&b, "App Store", inv.AppStore)
	section(&b, "outro", inv.Other)
	return b.Bytes()
}

func section(b *bytes.Buffer, title string, items []inventory.Item) {
	fmt.Fprintf(b, "\n## %s\n\n", title)
	if len(items) == 0 {
		b.WriteString("_Nenhum aplicativo._\n")
		return
	}
	b.WriteString("| Nome | Versão |\n| --- | --- |\n")
	for _, it := range items {
		version := it.Version
		if version == "" {
			version = "desconhecida"
		}
		fmt.Fprintf(b, "| %s | %s |\n", escapeCell(it.Name), escapeCell(version))
	}
}

func escapeCell(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}
