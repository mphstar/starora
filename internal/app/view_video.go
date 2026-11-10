package app

import (
	"os/exec"
	"path/filepath"

	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"
	"starora/internal/engine"
)

func (a *App) viewVideoUpscaler(c *ui.Context) {
	ui.Row(c).Fill().Padding(10, 14, 10).Gap(12).Children(func() {
		// Left / Center: Dominant Viewport Canvas & Timeline
		ui.Column(c).Grow(1).MinHeight(0).Children(func() {
			a.viewVideoPreview(c)
		})

		// Right: Tactile Video Inspector Sidebar (Width: 360px, Scrollable)
		ui.Scroll(c).Width(360).Shrink(0).FillHeight().KeepScrollbars(true).TrackScroll(&a.VideoSideScroll).Gap(10).Children(func() {
			a.viewVideoSourceCard(c)
			a.viewVideoModelCard(c)
			a.viewVideoRenderCard(c)
		})
	})
}

func (a *App) viewVideoSourceCard(c *ui.Context) {
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
		a.HandleDropFile(drop, true)
		c.Toast("File video berhasil diimpor!")
	}

	dropZone.Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(6).Children(func() {
			ui.Icon(c, videoIcon).Width(14).Height(14).TextColor(t.Accent)
			ui.Text(c, "File Video Input & Lokasi Output").FontSize(12).Bold().TextColor(t.Text)
		})

		ui.Column(c).Gap(3).Children(func() {
			ui.Text(c, "File Video Input:").FontSize(11).Bold().TextColor(t.TextMuted)
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				placeholder := "Pilih atau drag & drop file video..."
				ui.TextInput(c, &a.InputVideoPath).Label("Input Video Path").Placeholder(placeholder).Grow(1)
				if ui.Button(c, "Browse Video...").Clicked() {
					go a.PickVideoFile()
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

		if a.InputVideoPath != "" {
			ui.Row(c).Gap(8).Padding(6, 8).Radius(6).Background(t.Background).Border(1, t.Border).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, checkIcon).Width(13).Height(13).TextColor(t.Success)
				ui.Column(c).Gap(1).Grow(1).Children(func() {
					ui.Textf(c, "Video: %s", filepath.Base(a.InputVideoPath)).FontSize(11).Bold().TextColor(t.Text)
					ui.Textf(c, "Durasi: %s", engine.FormatDuration(a.VideoDurationSec)).FontSize(10).TextColor(t.TextMuted)
				})
			})
		}
	})
}

func (a *App) viewVideoModelCard(c *ui.Context) {
	t := c.Theme()

	card(c, "Pilihan Model AI Video Real-ESRGAN & Audio", func() {
		ui.Column(c).Gap(8).Children(func() {
			ui.Text(c, "Pilih Model Video Real-ESRGAN:").FontSize(11).Bold().TextColor(t.TextMuted)
			ui.Select(c, &a.VideoModel, VideoModelList)

			scale := a.GetVideoScale()
			ui.Row(c).Gap(6).Padding(5, 8).Radius(6).Background(t.Background).Border(1, t.Border).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, cpuIcon).Width(13).Height(13).TextColor(t.Accent)
				ui.Textf(c, "%s  •  %dx 4K Super-Res", GetCleanModelName(a.VideoModel), scale).FontSize(11).Bold().TextColor(t.Text)
			})

			ui.Column(c).Gap(3).Children(func() {
				ui.Text(c, "Format Video Output:").FontSize(11).Bold().TextColor(t.TextMuted)
				codecs := []string{
					"H.264 / AVC (libx264 - Kompatibel Semua Pemutar)",
					"H.265 / HEVC (libx265 - Ukuran Lebih Kecil)",
					"Apple ProRes (prores_ks - Kualitas Master)",
				}
				ui.Select(c, &a.VideoCodec, codecs)
			})

			ui.Checkbox(c, &a.VideoKeepAudio, "Pertahankan trek audio asli (FFmpeg copy)")

			ui.Column(c).Gap(3).Children(func() {
				ui.Text(c, "Ekstraksi Frame & Kecepatan:").FontSize(11).Bold().TextColor(t.TextMuted)
				interFormats := []string{
					"JPG Kualitas Tinggi (5x Lebih Cepat)",
					"PNG (Lossless Murni)",
				}
				ui.Select(c, &a.VideoIntermediateFormat, interFormats)
			})

			ui.Column(c).Gap(3).Children(func() {
				ui.Text(c, "Kompresi Video (CRF):").FontSize(11).Bold().TextColor(t.TextMuted)
				crfs := []string{
					"CRF 28 (Kompresi Tinggi - Hemat ~70% Ukuran File)",
					"CRF 23 (Seimbang - Rekomendasi Standar Streaming)",
					"CRF 18 (Detail Maksimal - Kualitas Master / File Besar)",
				}
				ui.Select(c, &a.VideoCompressCRF, crfs)
			})
		})
	})
}

func (a *App) viewVideoRenderCard(c *ui.Context) {
	t := c.Theme()

	card(c, "Progres Render Video Keseluruhan", func() {
		ui.Column(c).Gap(10).Children(func() {
			ui.Row(c).Gap(6).Padding(6, 10).Radius(6).Background(t.Background).Border(1, t.Border).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Progres render dihitung secara langsung dan menyeluruh dari 0% hingga 100% selesai.").FontSize(11).TextColor(t.TextMuted)
			})

			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				if !a.IsProcessing {
					if ui.PrimaryButton(c, "Start Video Upscale").Grow(1).Clicked() {
						a.StartVideoUpscale()
					}
				} else {
					if ui.Button(c, "Cancel").Grow(1).Clicked() {
						a.CancelUpscale()
						c.Toast("Video upscale dibatalkan")
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

func (a *App) viewVideoPreview(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Grow(1).MinHeight(0).Padding(14, 16).Gap(10).Radius(t.Radius).Background(t.Surface).Border(1, t.Border).Children(func() {
		ui.Text(c, "Preview Frame Video: Sebelum vs Sesudah").FontSize(12).Bold().TextColor(t.Text)

		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Tabs(c, &a.VideoPreviewTab, "Sebelum", "Sesudah")

			scale := a.GetVideoScale()
			if a.VideoPreviewTab == 0 {
				ui.Row(c).Margin(0, 0, 0, 10).Padding(4, 8).Radius(6).Background(t.Background).Border(1, t.Border).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "FRAME SEBELUM (ORIGINAL)").FontSize(10).Bold().TextColor(t.Text)
					ui.Textf(c, "• %d×%d", a.VideoBeforeSize[0], a.VideoBeforeSize[1]).FontSize(10).TextColor(t.TextMuted)
				})
			} else {
				ui.Row(c).Margin(0, 0, 0, 10).Padding(4, 8).Radius(6).Background(t.SurfaceHover).Border(1, t.Border).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "FRAME SESUDAH (AI HD)").FontSize(10).Bold().TextColor(t.Accent)
					ui.Textf(c, "• %d×%d (%dx 4K)", a.VideoAfterSize[0], a.VideoAfterSize[1], scale).FontSize(10).Bold().TextColor(t.Text)
				})
			}

			ui.Spacer(c)

			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Textf(c, "Zoom: %.0f%%", a.VideoZoom*100).FontSize(11).Bold().TextColor(t.Accent)
				if ui.Button(c, "−").Clicked() {
					a.ZoomOutVideo()
				}
				if ui.Button(c, "+").Clicked() {
					a.ZoomInVideo()
				}
				if a.VideoZoom != 1.0 || a.VideoPanX != 0 || a.VideoPanY != 0 {
					if ui.Button(c, "Reset").Clicked() {
						a.ResetZoomVideo()
					}
				}
			})
		})

		var currentBm *ui.Bitmap
		if a.VideoPreviewTab == 0 {
			currentBm = a.VideoBeforeBitmap
		} else {
			currentBm = a.VideoAfterBitmap
			if currentBm == nil {
				currentBm = a.VideoBeforeBitmap
			}
		}

		canvasBg := ui.Hex("#070B12")
		canvas := ui.Column(c.Key("video-canvas")).Grow(1).MinHeight(320).Radius(t.Radius).Background(canvasBg).Border(1, t.Border).Clip()

		dx, dy, held := canvas.Dragged()
		if held {
			canvas.Cursor(ui.CursorGrabbing)
			if dx != 0 || dy != 0 {
				a.VideoPanX += dx
				a.VideoPanY += dy
			}
		} else {
			canvas.Cursor(ui.CursorGrab)
		}

		canvas.HandleInput(func(ev ui.InputEvent) bool {
			switch ev.Kind {
			case ui.InputZoom:
				a.VideoZoom *= ev.Scale
				if a.VideoZoom < 0.5 {
					a.VideoZoom = 0.5
				} else if a.VideoZoom > 10.0 {
					a.VideoZoom = 10.0
				}
				return true
			case ui.InputScroll:
				if ev.DY < 0 {
					a.VideoZoom *= 1.15
				} else if ev.DY > 0 {
					a.VideoZoom /= 1.15
				}
				if a.VideoZoom < 0.5 {
					a.VideoZoom = 0.5
				} else if a.VideoZoom > 10.0 {
					a.VideoZoom = 10.0
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
				dispW := baseW * a.VideoZoom
				dispH := baseH * a.VideoZoom

				cx := r.X + r.W/2 + a.VideoPanX
				cy := r.Y + r.H/2 + a.VideoPanY

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

		// Sleek video player transport bar
		ui.Row(c).Gap(8).AlignItems(ui.Center).Padding(8, 12).Radius(8).Background(t.Background).Border(1, t.Border).Children(func() {
			if a.VideoIsPlaying {
				if ui.Button(c, "⏸ Pause").Width(85).Clicked() {
					a.PauseVideo()
				}
			} else {
				if ui.PrimaryButton(c, "▶ Play").Width(85).Clicked() {
					a.PlayVideo()
				}
			}

			if ui.Button(c, "⏮").Width(36).Clicked() {
				a.RewindVideo()
			}

			ui.Text(c, engine.FormatDuration(a.VideoFrameSec)).FontSize(11).Bold().TextColor(t.Text).Width(60)

			slider := ui.Slider(c, &a.VideoFrameSec, 0, a.VideoDurationSec).Grow(1)
			if slider.Changed() {
				a.LastVideoFrameSec = a.VideoFrameSec
				if a.VideoIsPlaying {
					a.SeekPlayVideo(a.VideoFrameSec)
				} else {
					a.extractVideoFrame(a.VideoFrameSec)
				}
			}

			ui.Text(c, engine.FormatDuration(a.VideoDurationSec)).FontSize(11).TextColor(t.TextMuted).Width(60)

			if ui.Button(c, "Ambil Frame").Clicked() {
				a.extractVideoFrame(a.VideoFrameSec)
			}

			if !a.VideoIsPreviewing {
				if ui.Button(c, "Preview AI Frame").Clicked() {
					a.previewVideoFrameUpscale()
				}
			} else {
				ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
					ui.Spinner(c).Size(12, 12)
					ui.Text(c, "Upscaling...").FontSize(11).Bold().TextColor(t.Accent)
				})
			}
		})

		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Drag mouse untuk geser • Scroll mouse untuk zoom").FontSize(11).TextColor(t.TextMuted)
		})
	})
}
