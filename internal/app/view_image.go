package app

import (
	"os/exec"
	"path/filepath"

	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"
)

func (a *App) viewImageUpscaler(c *ui.Context) {
	ui.Row(c).Fill().Padding(10, 14, 10).Gap(12).Children(func() {
		// Left / Center: Dominant Viewport Canvas
		ui.Column(c).Grow(1).MinHeight(0).Children(func() {
			a.viewImagePreview(c)
		})

		// Right: Tactile Inspector Sidebar (Width: 360px, Scrollable)
		ui.Scroll(c).Width(360).Shrink(0).FillHeight().KeepScrollbars(true).TrackScroll(&a.ImageSideScroll).Gap(10).Children(func() {
			a.viewImageSourceCard(c)
			a.viewImageModelCard(c)
			a.viewImageRenderCard(c)
		})
	})
}

func (a *App) viewImageSourceCard(c *ui.Context) {
	t := c.Theme()

	dropZone := ui.Column(c).Padding(10, 12).Gap(8).Radius(t.Radius).Background(t.Surface).Border(1, t.Border)
	opts := transfer.DropOptions{
		Formats:    []transfer.Format{transfer.FileList, transfer.URIList},
		Operations: transfer.Copy,
	}
	if _, over := ui.DataDragOver(dropZone, opts); over {
		dropZone.Border(1, t.Accent).Background(t.SurfaceHover)
	}
	if drop, ok := ui.DropData(dropZone, opts); ok {
		a.HandleDropFile(drop, false)
		c.Toast("File gambar berhasil diimpor!")
	}

	dropZone.Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(6).Children(func() {
			ui.Icon(c, imageIcon).Width(14).Height(14).TextColor(t.Accent)
			ui.Text(c, "File Gambar Input & Lokasi Output").FontSize(12).Bold().TextColor(t.Text)
		})

		ui.Column(c).Gap(3).Children(func() {
			ui.Text(c, "File Gambar Input:").FontSize(11).Bold().TextColor(t.TextMuted)
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				placeholder := "Pilih atau drag & drop file..."
				ui.TextInput(c, &a.InputImagePath).Label("Input Image Path").Placeholder(placeholder).Grow(1)
				if ui.Button(c, "Browse Image...").Clicked() {
					go a.PickImageFile()
				}
			})
		})

		ui.Column(c).Gap(3).Children(func() {
			ui.Text(c, "Folder Output:").FontSize(11).Bold().TextColor(t.TextMuted)
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.TextInput(c, &a.OutputDir).Label("Output Folder").Placeholder("Folder output...").Grow(1)
				if ui.Button(c, "Browse Folder...").Clicked() {
					go a.PickOutputDir()
				}
				if a.OutputDir != "" {
					if ui.Button(c, "Buka Folder").Clicked() {
						go exec.Command("explorer", filepath.Clean(a.OutputDir)).Start()
						c.Toast("Membuka folder output di Explorer")
					}
				}
			})
		})

		if a.InputImagePath != "" {
			ui.Row(c).Gap(8).Padding(6, 8).Radius(6).Background(t.Background).Border(1, t.Border).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, checkIcon).Width(13).Height(13).TextColor(t.Success)
				ui.Column(c).Gap(1).Grow(1).Children(func() {
					ui.Textf(c, "File: %s", filepath.Base(a.InputImagePath)).FontSize(11).Bold().TextColor(t.Text)
					ui.Textf(c, "%d×%d px", a.ImageBeforeSize[0], a.ImageBeforeSize[1]).FontSize(10).TextColor(t.TextMuted)
				})
			})
		}
	})
}

func (a *App) viewImageModelCard(c *ui.Context) {
	t := c.Theme()

	card(c, "Pilihan Model AI Real-ESRGAN", func() {
		ui.Column(c).Gap(8).Children(func() {
			ui.Text(c, "Pilih Model Real-ESRGAN:").FontSize(11).Bold().TextColor(t.TextMuted)
			ui.Select(c, &a.ImageModel, ImageModelList)

			scale := a.GetImageScale()
			ui.Row(c).Gap(6).Padding(5, 8).Radius(6).Background(t.Background).Border(1, t.Border).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, cpuIcon).Width(13).Height(13).TextColor(t.Accent)
				ui.Textf(c, "%s  •  %dx HD", GetCleanModelName(a.ImageModel), scale).FontSize(11).Bold().TextColor(t.Text)
			})

			ui.Column(c).Gap(3).Children(func() {
				ui.Text(c, "Format Output (Default: PNG):").FontSize(11).Bold().TextColor(t.TextMuted)
				ui.Segmented(c, &a.ImageFormatIndex, "PNG", "JPG", "WebP")
			})

			ui.Checkbox(c, &a.ImageFaceEnhance, "Restorasi & Peningkatan Wajah (GFPGAN)")

			ui.Column(c).Gap(3).Children(func() {
				ui.Text(c, "Kompresi Output (Default: PNG):").FontSize(11).Bold().TextColor(t.TextMuted)
				ui.Checkbox(c, &a.ImageCompress, "Aktifkan Kompresi Ukuran Gambar")
				if a.ImageCompress {
					modes := []string{
						"PNG Teroptimasi (Hemat ~40% Size, Lossless Murni)",
						"WebP Optimal (Hemat ~85% Size, Kualitas Jernih)",
						"JPG Terkompresi (Hemat ~75% Size, Kompatibel)",
					}
					ui.Select(c, &a.ImageCompressFormat, modes)
					if a.LastCompressionSavings != "" {
						ui.Row(c).Gap(6).Padding(4, 8).Radius(4).Background(t.Background).Border(1, t.Border).Children(func() {
							ui.Textf(c, "Hasil: %s", a.LastCompressionSavings).FontSize(10).Bold().TextColor(t.Success)
						})
					}
				}
			})
		})
	})
}

func (a *App) viewImageRenderCard(c *ui.Context) {
	t := c.Theme()

	card(c, "Progres Render Gambar Keseluruhan", func() {
		ui.Column(c).Gap(10).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				if !a.IsProcessing {
					if ui.PrimaryButton(c, "Start Image Upscale").Grow(1).Clicked() {
						a.StartImageUpscale()
					}
				} else {
					if ui.Button(c, "Cancel").Grow(1).Clicked() {
						a.CancelUpscale()
						c.Toast("Upscale dibatalkan")
					}
				}

				if a.IsProcessing {
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						ui.Spinner(c).Size(14, 14)
						ui.Textf(c, "%.0f%%", a.Progress*100).FontSize(13).Bold().TextColor(t.Accent)
					})
				}
			})

			ui.Column(c).Gap(2).Children(func() {
				statusColor := t.Text
				if a.StatusIsError {
					statusColor = t.Danger
				}
				ui.Text(c, a.StatusMessage).Bold().FontSize(12).TextColor(statusColor)
				if a.CurrentStage != "" {
					ui.Text(c, a.CurrentStage).FontSize(11).TextColor(t.TextMuted)
				}
			})

			if a.IsProcessing || a.Progress > 0 {
				ui.Progress(c, a.Progress)
			}
		})
	})
}

func (a *App) viewImagePreview(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Grow(1).MinHeight(0).Padding(14, 16).Gap(10).Radius(t.Radius).Background(t.Surface).Border(1, t.Border).Children(func() {
		ui.Text(c, "Preview: Sebelum vs Sesudah (Visual Comparison)").FontSize(12).Bold().TextColor(t.Text)

		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Tabs(c, &a.ImagePreviewTab, "Sebelum", "Sesudah")

			scale := a.GetImageScale()
			if a.ImagePreviewTab == 0 {
				ui.Row(c).Margin(0, 0, 0, 10).Padding(4, 8).Radius(6).Background(t.Background).Border(1, t.Border).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "SEBELUM (ORIGINAL)").FontSize(10).Bold().TextColor(t.Text)
					ui.Textf(c, "• %d×%d", a.ImageBeforeSize[0], a.ImageBeforeSize[1]).FontSize(10).TextColor(t.TextMuted)
				})
			} else {
				ui.Row(c).Margin(0, 0, 0, 10).Padding(4, 8).Radius(6).Background(t.SurfaceHover).Border(1, t.Border).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "SESUDAH (REAL-ESRGAN HD)").FontSize(10).Bold().TextColor(t.Accent)
					ui.Textf(c, "• %d×%d (%dx HD)", a.ImageAfterSize[0], a.ImageAfterSize[1], scale).FontSize(10).Bold().TextColor(t.Text)
				})
			}

			ui.Spacer(c)

			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Textf(c, "Zoom: %.0f%%", a.ImageZoom*100).FontSize(11).Bold().TextColor(t.Accent)
				if ui.Button(c, "−").Clicked() {
					a.ZoomOutImage()
				}
				if ui.Button(c, "+").Clicked() {
					a.ZoomInImage()
				}
				if a.ImageZoom != 1.0 || a.ImagePanX != 0 || a.ImagePanY != 0 {
					if ui.Button(c, "Reset").Clicked() {
						a.ResetZoomImage()
					}
				}
				if ui.Button(c, "Quick AI Preview").Clicked() {
					a.QuickImagePreview()
					c.Toast("Quick AI preview updated")
				}
			})
		})

		var currentBm *ui.Bitmap
		if a.ImagePreviewTab == 0 {
			currentBm = a.ImageBeforeBitmap
		} else {
			currentBm = a.ImageAfterBitmap
		}

		canvasBg := ui.Hex("#070B12")
		canvas := ui.Column(c.Key("image-canvas")).Grow(1).MinHeight(320).Radius(t.Radius).Background(canvasBg).Border(1, t.Border).Clip()

		dx, dy, held := canvas.Dragged()
		if held {
			canvas.Cursor(ui.CursorGrabbing)
			if dx != 0 || dy != 0 {
				a.ImagePanX += dx
				a.ImagePanY += dy
			}
		} else {
			canvas.Cursor(ui.CursorGrab)
		}

		canvas.HandleInput(func(ev ui.InputEvent) bool {
			switch ev.Kind {
			case ui.InputZoom:
				a.ImageZoom *= ev.Scale
				if a.ImageZoom < 0.5 {
					a.ImageZoom = 0.5
				} else if a.ImageZoom > 10.0 {
					a.ImageZoom = 10.0
				}
				return true
			case ui.InputScroll:
				if ev.DY < 0 {
					a.ImageZoom *= 1.15
				} else if ev.DY > 0 {
					a.ImageZoom /= 1.15
				}
				if a.ImageZoom < 0.5 {
					a.ImageZoom = 0.5
				} else if a.ImageZoom > 10.0 {
					a.ImageZoom = 10.0
				}
				return true
			}
			return false
		})

		canvas.Draw(func(p *ui.Painter, r ui.Rect) {
			if currentBm != nil {
				origW, origH := currentBm.Size()
				if origW <= 0 || origH <= 0 {
					origW, origH = 640, 360
				}
				aspect := float32(origW) / float32(origH)
				baseH := r.H * 0.85
				baseW := baseH * aspect
				if baseW > r.W*0.90 {
					baseW = r.W * 0.90
					baseH = baseW / aspect
				}
				dispW := baseW * a.ImageZoom
				dispH := baseH * a.ImageZoom

				cx := r.X + r.W/2 + a.ImagePanX
				cy := r.Y + r.H/2 + a.ImagePanY

				imgBox := ui.Rect{
					X: cx - dispW/2,
					Y: cy - dispH/2,
					W: dispW,
					H: dispH,
				}

				p.Clip(r, t.Radius, func() {
					p.Image(currentBm, imgBox, ui.FillBox)
				})
			}
		})

		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Perbandingan langsung piksel resolusi sebelum dan sesudah upscale • Drag mouse untuk geser • Scroll mouse untuk zoom").FontSize(11).TextColor(t.TextMuted)
		})
	})
}
