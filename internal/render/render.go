// Package render turns an inventory into the Markdown of the inventory file.
package render

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mxfca/mac-inventory/internal/inventory"
)

// Markdown returns the inventory file: frontmatter plus one section per source.
func Markdown(inv inventory.Inventory, collectedAt time.Time) []byte {
	var b bytes.Buffer
	failures := inv.Failures()
	status := "completa"
	if len(failures) > 0 {
		status = "parcial"
	}
	fmt.Fprintf(&b, `---
atualizado: %s
por: mac-inventory
status: ativo
tags: [inventario]
coletado_em: %s
coleta: %s
falhas: %s
---

<!-- Gerado pela coleta semanal do mac-inventory. Não editar à mão. -->
`, collectedAt.Format(time.DateOnly), collectedAt.Format(time.RFC3339), status, flowList(failures))

	section(&b, "brew", inv.Brew, inv.BrewFailure)
	section(&b, "App Store", inv.AppStore, inv.AppsFailure)
	section(&b, "outro", inv.Other, inv.AppsFailure)
	return b.Bytes()
}

func section(b *bytes.Buffer, title string, items []inventory.Item, failure string) {
	fmt.Fprintf(b, "\n## %s\n\n", title)
	if failure != "" {
		fmt.Fprintf(b, "_Coleta falhou: %s._\n", failure)
		return
	}
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

// flowList formats a YAML flow sequence of quoted strings, e.g. ["a", "b"].
func flowList(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = strconv.Quote(s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
