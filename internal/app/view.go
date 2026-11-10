package app

import (
	"runtime"

	"github.com/egoist/mygo/ui"
	"starora/internal/theme"
)

// View is the root view callback for Starora Studio.
func (a *App) View(c *ui.Context) {
	if a.IsDarkMode {
		c.SetTheme(theme.DarkStudioTheme())
	} else {
		c.SetTheme(theme.LightStudioTheme())
	}
	t := c.Theme()

	ui.Column(c).Fill().Background(t.Background).Children(func() {
		// Sleek studio header & workspace tabs
		a.viewHeader(c)

		// Workspace body
		ui.Column(c).Grow(1).MinHeight(0).Children(func() {
			switch a.ActiveTab {
			case 1:
				a.viewVideoUpscaler(c)
			case 2:
				a.viewEngineSettings(c)
			case 3:
				a.viewLogs(c)
			case 4:
				a.viewAbout(c)
			default:
				a.viewImageUpscaler(c)
			}
		})

		// Hardware telemetry status bar at the bottom
		a.viewFooter(c)
	})
}

func (a *App) viewHeader(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Background(t.Surface).Border(1, t.Border).Shrink(0).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Padding(12, 20, 10).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Gap(10).Children(func() {
				if a.AppLogoBitmap != nil {
					ui.Image(c, a.AppLogoBitmap).Size(34, 34).Radius(8).Border(1, t.Border).Clip()
				} else {
					ui.Row(c).Width(34).Height(34).Radius(8).Center().Background(t.SurfaceHover).Border(1, t.Border).Children(func() {
						ui.Icon(c, layersIcon).Width(18).Height(18).TextColor(t.Accent)
					})
				}
				ui.Column(c).Gap(1).Children(func() {
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						ui.Text(c, "Starora AI Upscaler").FontSize(14).Bold().TextColor(t.Text)
						ui.Row(c).Padding(1, 5).Radius(4).Background(t.Background).Border(1, t.Border).Children(func() {
							ui.Text(c, "STUDIO PRO").FontSize(9).Bold().TextColor(t.Accent)
						})
					})
					ui.Text(c, "Super-resolution neural enhancer").FontSize(11).TextColor(t.TextMuted)
				})
			})

			ui.Spacer(c)

			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				statusBadge(c, "Real-ESRGAN", a.RealesrganPath != "")
				statusBadge(c, "FFmpeg", a.FfmpegPath != "")

				themeLabel := "Dark"
				themeIcon := moonIcon
				if a.IsDarkMode {
					themeLabel = "Light"
					themeIcon = sunIcon
				}
				if ui.Button(c, themeLabel).Children(func() {
					ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
						ui.Icon(c, themeIcon).Width(12).Height(12).TextColor(t.TextMuted)
						ui.Text(c, themeLabel).FontSize(11).Bold().TextColor(t.Text)
					})
				}).Clicked() {
					a.IsDarkMode = !a.IsDarkMode
					if a.IsDarkMode {
						c.Toast("Switched to Dark Studio mode")
					} else {
						c.Toast("Switched to Clean Light mode")
					}
				}
			})
		})

		ui.Row(c).Padding(2, 16, 4).Children(func() {
			ui.Tabs(c, &a.ActiveTab, "Image Upscale", "Video Upscale", "Engine Settings", "Activity Logs", "About")
		})
	})
}

func (a *App) viewFooter(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Height(28).Shrink(0).Padding(0, 18).AlignItems(ui.Center).Background(t.Surface).Border(1, t.Border).Children(func() {
		dotColor := t.Success
		if a.StatusIsError {
			dotColor = t.Danger
		} else if a.IsProcessing {
			dotColor = t.Accent
		}
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Row(c).Width(6).Height(6).Radius(3).Background(dotColor)
			ui.Text(c, a.StatusMessage).FontSize(11).Bold().TextColor(t.Text)
		})

		ui.Spacer(c)

		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "NCNN Vulkan GPU").FontSize(11).TextColor(t.TextMuted)
			ui.Text(c, "•").FontSize(11).TextColor(t.Border)
			ui.Textf(c, "%d CPU Cores", runtime.NumCPU()).FontSize(11).TextColor(t.TextMuted)
			ui.Text(c, "•").FontSize(11).TextColor(t.Border)
			ui.Text(c, "Starora Studio v0.1.0").FontSize(11).Bold().TextColor(t.TextMuted)
		})
	})
}
