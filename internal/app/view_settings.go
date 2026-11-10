package app

import (
	"os/exec"
	"path/filepath"
	"runtime"

	"starora/internal/engine"

	"github.com/egoist/mygo/ui"
)

func (a *App) viewEngineSettings(c *ui.Context) {
	t := c.Theme()

	ui.Scroll(c).Grow(1).Padding(20, 32, 28).Gap(16).Children(func() {
		card(c, "Real-ESRGAN & FFmpeg Engine Status", func() {
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Engine Real-ESRGAN (NCNN Vulkan) dan FFmpeg terintegrasi langsung pada sistem lokal.").FontSize(12).TextColor(t.TextMuted)
				ui.Spacer(c)
				if ui.Button(c, "Re-detect Engines").Clicked() {
					a.DetectBinaries()
					a.AddLog("Re-scanned engine binaries")
					c.Toast("Engine binaries re-scanned")
				}
				if ui.Button(c, "Open Engines Folder").Clicked() {
					go exec.Command("explorer", filepath.Clean(engine.GetEngineRootDir())).Start()
					c.Toast("Opened engine directory in Explorer")
				}
			})

			ui.Column(c).Gap(6).Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "Real-ESRGAN Executable:").FontSize(12).Bold().TextColor(t.Text)
					statusBadge(c, "Status", a.RealesrganPath != "")
				})
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.TextInput(c, &a.RealesrganPath).ReadOnly(true).Grow(1)
					if ui.Button(c, "Browse...").Clicked() {
						go a.PickRealESRGANPath()
					}
				})
			})

			ui.Column(c).Gap(6).Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "FFmpeg Executable:").FontSize(12).Bold().TextColor(t.Text)
					statusBadge(c, "Status", a.FfmpegPath != "")
				})
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.TextInput(c, &a.FfmpegPath).ReadOnly(true).Grow(1)
					if ui.Button(c, "Browse...").Clicked() {
						go a.PickFFmpegPath()
					}
				})
			})
		})

		card(c, "Akselerasi Perangkat Keras & GPU", func() {
			ui.Row(c).Gap(20).Wrap().Children(func() {
				ui.Column(c).Gap(6).Basis(280).Grow(1).Children(func() {
					ui.Text(c, "Perangkat Pemrosesan (Vulkan GPU):").FontSize(12).Bold().TextColor(t.TextMuted)
					devices := []string{
						"Auto (Vulkan)",
						"GPU 0 (Primary Discrete GPU)",
						"GPU 1 (Secondary GPU)",
						"CPU Only (-g -1)",
					}
					ui.Select(c, &a.GpuDevice, devices)
					ui.Text(c, "Real-ESRGAN memanfaatkan akselerasi GPU compute Vulkan secara native.").FontSize(11).TextColor(t.TextMuted)
				})

				ui.Column(c).Gap(8).Basis(260).Grow(1).Children(func() {
					ui.Text(c, "Opsi Memori & Sistem:").FontSize(12).Bold().TextColor(t.TextMuted)
					ui.Checkbox(c, &a.AutoClearVram, "Bersihkan file sementara / VRAM setelah selesai")
					ui.Checkbox(c, &a.DesktopNotify, "Kirim notifikasi desktop setelah selesai")
				})
			})
		})

		card(c, "Optimasi Performa & Multi-Threading Engine", func() {
			ui.Column(c).Gap(12).Children(func() {
				ui.Row(c).Gap(10).Padding(10, 14).Radius(6).Background(t.Background).Border(1, t.Border).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, cpuIcon).Width(16).Height(16).TextColor(t.Accent)
					ui.Textf(c, "%d Logical CPU Cores Terdeteksi  •  Multi-Threaded I/O & Compute Aktif", runtime.NumCPU()).FontSize(12).Bold().TextColor(t.Text)
				})

				ui.Row(c).Gap(20).Wrap().Children(func() {
					ui.Column(c).Gap(6).Basis(280).Grow(1).Children(func() {
						ui.Text(c, "Distribusi Thread Real-ESRGAN (-j load:proc:save):").FontSize(12).Bold().TextColor(t.TextMuted)
						modes := []string{
							"Turbo Multi-Thread (Otomatis Maksimal)",
							"Seimbang (Balanced 2:2:2)",
							"Standard (1:2:1)",
						}
						ui.Select(c, &a.ThreadMode, modes)
						ui.Text(c, "Memparalelkan decoding, inferensi GPU, dan penulisan disk untuk mencegah bottleneck I/O.").FontSize(11).TextColor(t.TextMuted)
					})

					ui.Column(c).Gap(6).Basis(280).Grow(1).Children(func() {
						ui.Text(c, "Ukuran Tile GPU (-t tile-size):").FontSize(12).Bold().TextColor(t.TextMuted)
						tiles := []string{
							"0 (Auto - Rekomendasi Tercepat / Full GPU)",
							"200 (Hemat VRAM / GPU 2-4GB)",
							"400 (Seimbang GPU 6-8GB)",
							"800 (Performa Tinggi GPU 12GB+)",
						}
						ui.Select(c, &a.ImageTileSize, tiles)
						ui.Text(c, "Tile size 0 memproses gambar secara utuh tanpa overhead pemotongan bila VRAM cukup.").FontSize(11).TextColor(t.TextMuted)
					})
				})
			})
		})
	})
}

func (a *App) viewLogs(c *ui.Context) {
	t := c.Theme()

	ui.Scroll(c).Grow(1).Padding(20, 32, 28).Gap(16).Children(func() {
		card(c, "Application Activity & Console Logs", func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Row(c).Padding(4, 10).Radius(6).Background(t.Background).Border(1, t.Border).AlignItems(ui.Center).Children(func() {
					ui.Textf(c, "Total Log Entries: %d", len(a.Logs)).FontSize(12).Bold().TextColor(t.Text)
				})
				ui.Spacer(c)
				if ui.Button(c, "Copy Logs").Clicked() {
					c.Services().WriteClipboard(a.AllLogsText())
					a.AddLog("Copied logs to clipboard")
					c.Toast("Logs copied to clipboard")
				}
				if ui.Button(c, "Clear Logs").Clicked() {
					a.ClearLogs()
					c.Toast("Logs cleared")
				}
			})

			logBox := ui.Column(c).Height(460).Padding(14).Radius(8).Background(t.Background).Border(1, t.Border)
			logBox.Children(func() {
				ui.Scroll(c).Grow(1).Gap(4).Children(func() {
					for _, logMsg := range a.Logs {
						ui.Text(c, logMsg).FontSize(11).Font("monospace").TextColor(t.Text)
					}
				})
			})
		})
	})
}

func (a *App) viewAbout(c *ui.Context) {
	t := c.Theme()

	ui.Scroll(c).Grow(1).Padding(20, 32, 28).Gap(16).Children(func() {
		card(c, "", func() {
			ui.Row(c).Gap(18).AlignItems(ui.Center).Children(func() {
				if a.AppLogoBitmap != nil {
					ui.Image(c, a.AppLogoBitmap).Size(58, 58).Radius(14).Border(1, t.Border).Clip()
				} else {
					ui.Row(c).Width(58).Height(58).Radius(14).Center().Background(t.SurfaceHover).Border(1, t.Border).Children(func() {
						ui.Icon(c, layersIcon).Width(30).Height(30).TextColor(t.Accent)
					})
				}

				ui.Column(c).Gap(4).Children(func() {
					ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
						ui.Text(c, "Starora AI Studio").FontSize(18).Bold().TextColor(t.Text)
						ui.Row(c).Padding(2, 8).Radius(6).Background(t.Background).Border(1, t.Border).Children(func() {
							ui.Text(c, "v0.1.0 Pro").FontSize(11).Bold().TextColor(t.Accent)
						})
					})
					ui.Text(c, "Aplikasi Desktop AI Super-Resolution untuk Gambar & Video").FontSize(13).Bold().TextColor(t.TextMuted)
					ui.Text(c, "Meningkatkan resolusi dan merestorasi detail gambar maupun frame video secara offline dengan akselerasi GPU Vulkan.").FontSize(12).TextColor(t.TextMuted)
				})
			})
		})

		card(c, "Repositori & Kode Sumber", func() {
			ui.Column(c).Gap(10).Children(func() {
				ui.Row(c).Gap(14).Padding(12, 16).Radius(8).Background(t.Background).Border(1, t.Border).AlignItems(ui.Center).Children(func() {
					ui.Row(c).Width(36).Height(36).Radius(8).Center().Background(t.SurfaceHover).Border(1, t.Border).Children(func() {
						ui.Icon(c, githubIcon).Width(18).Height(18).TextColor(t.Text)
					})
					ui.Column(c).Gap(2).Grow(1).Children(func() {
						ui.Text(c, "GitHub: mphstar/starora").FontSize(13).Bold().TextColor(t.Text)
						ui.Text(c, "Source code Starora, rilis versi terbaru, dan pelaporan issue/bug.").FontSize(11).TextColor(t.TextMuted)
					})
					if ui.PrimaryButton(c, "Buka GitHub").Clicked() {
						c.OpenURL(a.GithubURL)
						c.Toast("Membuka GitHub di browser...")
					}
				})
			})
		})

		card(c, "Ucapan Terima Kasih (Special Thanks & Sumber Terbuka)", func() {
			ui.Column(c).Gap(14).Children(func() {
				ui.Text(c, "Starora dapat berjalan kencang dan handal berkat karya luar biasa dari komunitas open source dunia:").FontSize(12).TextColor(t.TextMuted)

				ui.Column(c).Gap(8).Padding(12, 16).Radius(8).Background(t.Background).Border(1, t.Border).Children(func() {
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Icon(c, cpuIcon).Width(16).Height(16).TextColor(t.Accent)
						ui.Text(c, "Real-ESRGAN (Xintao Wang, Liangbin Xie, Chao Dong, Ying Shan)").FontSize(13).Bold().TextColor(t.Text)
					})
					ui.Text(c, "Algoritma Super-Resolution canggih untuk restorasi gambar dunia nyata dan video. Melatih jaringan neural sintetis untuk meningkatkan ketajaman visual tanpa artefak.").FontSize(11).TextColor(t.TextMuted)
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						if ui.Button(c, "Sumber: GitHub Real-ESRGAN").Clicked() {
							c.OpenURL("https://github.com/xinntao/Real-ESRGAN")
						}
						if ui.Button(c, "Sumber: NCNN Vulkan Engine").Clicked() {
							c.OpenURL("https://github.com/xinntao/Real-ESRGAN-ncnn-vulkan")
						}
					})
				})

				ui.Column(c).Gap(8).Padding(12, 16).Radius(8).Background(t.Background).Border(1, t.Border).Children(func() {
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Icon(c, videoIcon).Width(16).Height(16).TextColor(t.Success)
						ui.Text(c, "FFmpeg Project (Fabrice Bellard & FFmpeg Developers)").FontSize(13).Bold().TextColor(t.Text)
					})
					ui.Text(c, "Solusi multimedia universal terdepan di dunia untuk pemrosesan decoding, demuxing audio/video, ekstraksi frame presisi tinggi, dan re-encoding multi-core.").FontSize(11).TextColor(t.TextMuted)
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						if ui.Button(c, "Sumber: FFmpeg.org").Clicked() {
							c.OpenURL("https://ffmpeg.org")
						}
					})
				})
			})
		})
	})
}
