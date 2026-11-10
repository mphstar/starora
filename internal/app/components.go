package app

import "github.com/egoist/mygo/ui"

// card creates a tactile engineering surface with crisp 1px border.
func card(c *ui.Context, title string, body func()) ui.Element {
	t := c.Theme()
	return ui.Column(c).Padding(14, 16).Gap(10).Radius(t.Radius).Background(t.Surface).Border(1, t.Border).Children(func() {
		if title != "" {
			ui.Row(c).AlignItems(ui.Center).Gap(6).Children(func() {
				ui.Text(c, title).FontSize(12).Bold().TextColor(t.Text)
			})
		}
		body()
	})
}

// statusBadge renders a status indicator with a crisp state dot.
func statusBadge(c *ui.Context, name string, ok bool) ui.Element {
	t := c.Theme()
	statusText := "Ready"
	dotColor := t.Success
	if !ok {
		statusText = "Missing"
		dotColor = t.Danger
	}
	return ui.Row(c).Gap(6).AlignItems(ui.Center).Padding(4, 8).Radius(6).Background(t.Background).Border(1, t.Border).Children(func() {
		ui.Row(c).Width(6).Height(6).Radius(3).Background(dotColor)
		ui.Textf(c, "%s: %s", name, statusText).FontSize(11).Bold().TextColor(t.Text)
	})
}
