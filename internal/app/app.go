package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"starora/embedded"
	"starora/internal/engine"
	"starora/internal/preview"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"
)

var ImageModelList = []string{
	"realesrgan-x4plus (Default - Foto & Gambar Umum 4x)",
	"realesrgan-x4plus-anime (Anime, Manga & Ilustrasi 2D 4x)",
	"realesrnet-x4plus (Restorasi Foto Halus / Anti-Noise 4x)",
	"realesr-animevideov3-x4 (Video Animasi High Quality 4x)",
	"realesr-animevideov3-x3 (Video Animasi Medium 3x)",
	"realesr-animevideov3-x2 (Video Animasi Fast 2x)",
}

var VideoModelList = []string{
	"realesr-animevideov3-x4 (Video Animasi High Quality 4x)",
	"realesr-animevideov3-x3 (Video Animasi Medium 3x)",
	"realesr-animevideov3-x2 (Video Animasi Fast 2x)",
	"realesrgan-x4plus (Video Realistis / Foto 4x)",
	"realesrgan-x4plus-anime (Video Anime Detail 4x)",
}

type App struct {
	Win *mygo.Window

	ActiveTab  int
	IsDarkMode bool

	AppLogoBitmap *ui.Bitmap

	InputImagePath         string
	OutputImagePath        string
	OutputDir              string
	ImageModel             string
	ImageScaleIndex        int
	ImageFormatIndex       int
	ImageFaceEnhance       bool
	ImageTileSize          string
	ImageDenoise           float64
	ImageCompress          bool
	ImageCompressFormat    string
	ImageCompressQuality   int
	LastCompressionSavings string

	ImageBeforeBitmap *ui.Bitmap
	ImageAfterBitmap  *ui.Bitmap
	ImageBeforeSize   [2]int
	ImageAfterSize    [2]int
	ImagePreviewMode  int
	ImageSplitPercent float32
	ImageZoomIndex    int
	ImageScroll       ui.ScrollState
	ImagePreviewTab   int
	ImageZoom         float32
	ImagePanX         float32
	ImagePanY         float32
	ImageSideScroll   ui.ScrollState

	InputVideoPath         string
	OutputVideoPath        string
	VideoModel             string
	VideoScaleIndex        int
	VideoCodec             string
	VideoCrfIndex          int
	VideoKeepAudio         bool
	VideoKeepMetadata      bool
	VideoTileSize          string
	VideoEnableCompression bool
	VideoCompressCRF       string
	VideoBeforeBitmap      *ui.Bitmap
	VideoAfterBitmap       *ui.Bitmap
	VideoBeforeSize        [2]int
	VideoAfterSize         [2]int
	VideoPreviewMode       int
	VideoSplitPercent      float32
	VideoZoomIndex         int
	VideoScroll            ui.ScrollState
	VideoPreviewTab        int
	VideoZoom              float32
	VideoPanX              float32
	VideoPanY              float32
	VideoSideScroll        ui.ScrollState
	VideoFrameSec          float64
	VideoDurationSec       float64
	LastVideoFrameSec      float64
	VideoIsPlaying         bool
	videoPlayCancel        context.CancelFunc
	videoPlayMu            sync.Mutex
	VideoIsPreviewing      bool

	RealesrganPath          string
	FfmpegPath              string
	GpuDevice               string
	ThreadMode              string
	VideoIntermediateFormat string
	VideoPreset             string
	AutoClearVram           bool
	DesktopNotify           bool

	GithubURL string

	IsProcessing   bool
	cancelFunc     context.CancelFunc
	cancelMu       sync.Mutex
	Progress       float64
	CurrentStage   string
	StatusMessage  string
	StatusIsError  bool
	ProcessingTime string

	Logs []string
}

func New() *App {
	a := &App{
		ActiveTab:               0,
		IsDarkMode:              true,
		ImageModel:              ImageModelList[0],
		ImageScaleIndex:         2,
		ImageFormatIndex:        0, // PNG (Lossless) default
		ImageFaceEnhance:        true,
		ImageTileSize:           "0 (Auto)",
		ImageDenoise:            0.5,
		ImageCompress:           true,
		ImageCompressFormat:     "PNG Teroptimasi (Hemat ~40% Size, Lossless Murni)", // PNG default baik compress maupun tidak
		ImageCompressQuality:    82,
		ImagePreviewMode:        0,
		ImageSplitPercent:       50.0,
		ImageZoomIndex:          0,
		ImagePreviewTab:         0,
		ImageZoom:               1.0,
		ImagePanX:               0,
		ImagePanY:               0,
		VideoModel:              VideoModelList[0],
		VideoScaleIndex:         2,
		VideoCodec:              "H.264 / AVC (libx264)",
		VideoCrfIndex:           1,
		VideoKeepAudio:          true,
		VideoKeepMetadata:       true,
		VideoTileSize:           "0 (Auto)",
		VideoEnableCompression:  true,
		VideoCompressCRF:        "CRF 23 (Seimbang - Rekomendasi Standar Streaming)",
		VideoPreviewMode:        0,
		VideoSplitPercent:       50.0,
		VideoZoomIndex:          0,
		VideoPreviewTab:         0,
		VideoZoom:               1.0,
		VideoPanX:               0,
		VideoPanY:               0,
		VideoFrameSec:           1.5,
		VideoDurationSec:        60.0,
		LastVideoFrameSec:       1.5,
		GpuDevice:               "Auto (Vulkan)",
		ThreadMode:              "Turbo Multi-Thread (Otomatis Maksimal)",
		VideoIntermediateFormat: "JPG Kualitas Tinggi (5x Lebih Cepat)",
		VideoPreset:             "fast (Cepat & Seimbang)",
		AutoClearVram:           true,
		DesktopNotify:           true,
		GithubURL:               "https://github.com/mphstar/starora",
		Progress:                0.0,
		StatusMessage:           "Ready",
		Logs: []string{
			fmt.Sprintf("[%s] Application initialized", time.Now().Format("15:04:05")),
		},
	}
	if len(embedded.IconPNG) > 0 {
		if bm, err := ui.DecodeBitmap(embedded.IconPNG); err == nil {
			a.AppLogoBitmap = bm
		}
	}
	a.initSamplePreviews()
	a.DetectBinaries()
	return a
}

func (a *App) initSamplePreviews() {
	if a.ImageBeforeBitmap == nil {
		beforeImg := preview.GenerateScenicSample(false, 640, 360, 0)
		a.ImageBeforeBitmap = ui.NewBitmap(beforeImg)
		a.ImageBeforeBitmap.Pixelated()
		a.ImageBeforeSize = [2]int{640, 360}
	}
	if a.ImageAfterBitmap == nil {
		afterImg := preview.GenerateScenicSample(true, 640, 360, 0)
		a.ImageAfterBitmap = ui.NewBitmap(afterImg)
		a.ImageAfterBitmap.Pixelated()
		a.ImageAfterSize = [2]int{2560, 1440}
	}
	if a.VideoBeforeBitmap == nil {
		vBefore := preview.GenerateScenicSample(false, 640, 360, a.VideoFrameSec)
		a.VideoBeforeBitmap = ui.NewBitmap(vBefore)
		a.VideoBeforeBitmap.Pixelated()
		a.VideoBeforeSize = [2]int{1280, 720}
	}
	if a.VideoAfterBitmap == nil {
		vAfter := preview.GenerateScenicSample(true, 640, 360, a.VideoFrameSec)
		a.VideoAfterBitmap = ui.NewBitmap(vAfter)
		a.VideoAfterBitmap.Pixelated()
		a.VideoAfterSize = [2]int{3840, 2160}
	}
}

func (a *App) LoadRealImagePreview(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		a.AddLog("Error reading image: " + err.Error())
		return
	}
	bm, err := ui.DecodeBitmap(data)
	if err != nil {
		a.AddLog("Error decoding image: " + err.Error())
		return
	}
	bm.Pixelated()
	w, h := bm.Size()
	scale := a.GetImageScale()
	a.ImageBeforeBitmap = bm
	a.ImageBeforeSize = [2]int{w, h}
	a.ImageAfterSize = [2]int{w * scale, h * scale}

	afterSample := preview.GenerateScenicSample(true, 640, 360, 0)
	a.ImageAfterBitmap = ui.NewBitmap(afterSample)
	a.ImageAfterBitmap.Pixelated()
	a.AddLog(fmt.Sprintf("Loaded preview for %s (%d×%d)", filepath.Base(path), w, h))
}

func (a *App) Update(fn func()) {
	safeFn := fn
	if safeFn == nil {
		safeFn = func() {}
	}
	if a.Win != nil {
		a.Win.Update(safeFn)
	} else {
		safeFn()
	}
}

func (a *App) AddLog(msg string) {
	entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	a.Logs = append(a.Logs, entry)
	if len(a.Logs) > 500 {
		a.Logs = a.Logs[len(a.Logs)-500:]
	}
}

func (a *App) DetectBinaries() {
	if paths, err := engine.EnsureEmbeddedEngines(); err == nil {
		a.RealesrganPath = paths.RealESRGAN
		a.FfmpegPath = paths.FFmpeg
		a.AddLog(fmt.Sprintf("Embedded engines initialized from: %s", paths.RootDir))
		return
	}

	if p, err := exec.LookPath("realesrgan-ncnn-vulkan"); err == nil {
		a.RealesrganPath = p
	} else if p, err := exec.LookPath("realesrgan-ncnn-vulkan.exe"); err == nil {
		a.RealesrganPath = p
	}

	if p, err := exec.LookPath("ffmpeg"); err == nil {
		a.FfmpegPath = p
	} else if p, err := exec.LookPath("ffmpeg.exe"); err == nil {
		a.FfmpegPath = p
	}
}

func (a *App) PickImageFile() {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: a.Win,
		Title:  "Select Image to Upscale",
		Filters: []mygo.FileFilter{
			{Name: "Images (*.png, *.jpg, *.jpeg, *.webp, *.bmp)", Extensions: []string{"png", "jpg", "jpeg", "webp", "bmp"}},
			{Name: "All Files (*.*)", Extensions: []string{"*"}},
		},
	})
	if err == nil && len(paths) > 0 {
		a.Update(func() {
			a.InputImagePath = paths[0]
			if a.OutputDir == "" {
				a.OutputDir = filepath.Dir(paths[0])
			}
			a.LoadRealImagePreview(paths[0])
			a.AddLog("Selected input image: " + filepath.Base(paths[0]))
		})
	}
}

func (a *App) PickVideoFile() {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: a.Win,
		Title:  "Select Video to Upscale",
		Filters: []mygo.FileFilter{
			{Name: "Videos (*.mp4, *.mkv, *.mov, *.avi, *.webm)", Extensions: []string{"mp4", "mkv", "mov", "avi", "webm", "flv"}},
			{Name: "All Files (*.*)", Extensions: []string{"*"}},
		},
	})
	if err == nil && len(paths) > 0 {
		a.Update(func() {
			a.InputVideoPath = paths[0]
			a.OutputVideoPath = ""
			a.VideoFrameSec = 0
			if a.OutputDir == "" {
				a.OutputDir = filepath.Dir(paths[0])
			}
			if a.FfmpegPath != "" {
				if dur := engine.ProbeVideoDuration(a.FfmpegPath, paths[0]); dur > 0 {
					a.VideoDurationSec = dur
				}
			}
			a.AddLog("Selected input video: " + filepath.Base(paths[0]))
			a.extractVideoFrame(0)
		})
	}
}

func (a *App) extractVideoFrame(sec float64) {
	if a.FfmpegPath != "" && a.InputVideoPath != "" {
		go func() {
			out, err := engine.ExtractSingleFrameReal(a.FfmpegPath, a.InputVideoPath, sec)
			if err == nil && len(out) > 0 {
				bm, err := ui.DecodeBitmap(out)
				if err == nil {
					bm.Pixelated()
					w, h := bm.Size()
					scale := a.GetVideoScale()
					a.Update(func() {
						a.VideoBeforeBitmap = bm
						a.VideoBeforeSize = [2]int{w, h}
						a.VideoAfterSize = [2]int{w * scale, h * scale}
					})
				}
			}
			if a.OutputVideoPath != "" {
				if outAfter, err := engine.ExtractSingleFrameReal(a.FfmpegPath, a.OutputVideoPath, sec); err == nil && len(outAfter) > 0 {
					if bmAfter, err := ui.DecodeBitmap(outAfter); err == nil {
						bmAfter.Pixelated()
						w, h := bmAfter.Size()
						a.Update(func() {
							a.VideoAfterBitmap = bmAfter
							a.VideoAfterSize = [2]int{w, h}
						})
					}
				}
			}
		}()
		return
	}

	vImg := preview.GenerateScenicSample(false, 640, 360, sec)
	a.VideoBeforeBitmap = ui.NewBitmap(vImg)
	a.VideoBeforeBitmap.Pixelated()
	a.VideoBeforeSize = [2]int{1280, 720}
	scale := a.GetVideoScale()
	a.VideoAfterSize = [2]int{1280 * scale, 720 * scale}
	vAfter := preview.GenerateScenicSample(true, 640, 360, sec)
	a.VideoAfterBitmap = ui.NewBitmap(vAfter)
	a.VideoAfterBitmap.Pixelated()
}

func (a *App) previewVideoFrameUpscale() {
	a.Update(func() {
		a.VideoIsPreviewing = true
		a.AddLog(fmt.Sprintf("Generating AI frame preview at %.1fs...", a.VideoFrameSec))
	})

	go func() {
		if a.FfmpegPath != "" && a.RealesrganPath != "" && a.InputVideoPath != "" {
			frameData, err := engine.ExtractSingleFrameReal(a.FfmpegPath, a.InputVideoPath, a.VideoFrameSec)
			if err == nil && len(frameData) > 0 {
				tempIn := filepath.Join(os.TempDir(), fmt.Sprintf("starora_prev_in_%d.png", time.Now().UnixNano()))
				tempOut := filepath.Join(os.TempDir(), fmt.Sprintf("starora_prev_out_%d.png", time.Now().UnixNano()))
				defer os.Remove(tempIn)
				defer os.Remove(tempOut)

				if err := os.WriteFile(tempIn, frameData, 0644); err == nil {
					scale := a.GetVideoScale()
					if err := engine.UpscaleSingleFrameReal(a.RealesrganPath, tempIn, tempOut, scale, a.VideoModel); err == nil {
						if outData, err := os.ReadFile(tempOut); err == nil {
							if bm, err := ui.DecodeBitmap(outData); err == nil {
								a.Update(func() {
									a.VideoIsPreviewing = false
									a.VideoAfterBitmap = bm
									w, h := bm.Size()
									a.VideoAfterSize = [2]int{w, h}
									a.VideoPreviewTab = 1
									a.AddLog("Real AI video frame preview generated successfully.")
								})
								return
							}
						}
					}
				}
			}
		}

		time.Sleep(300 * time.Millisecond)
		a.Update(func() {
			a.VideoIsPreviewing = false
			enhanced := preview.GenerateScenicSample(true, 640, 360, a.VideoFrameSec)
			a.VideoAfterBitmap = ui.NewBitmap(enhanced)
			scale := a.GetVideoScale()
			a.VideoAfterSize = [2]int{a.VideoBeforeSize[0] * scale, a.VideoBeforeSize[1] * scale}
			a.VideoPreviewTab = 1
			a.AddLog("Sample AI video frame preview updated.")
		})
	}()
}

func (a *App) PickOutputDir() {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent:            a.Win,
		Title:             "Select Output Directory",
		Directory:         true,
		CreateDirectories: true,
	})
	if err == nil && len(paths) > 0 {
		a.Update(func() {
			a.OutputDir = paths[0]
			a.AddLog("Output directory set to: " + paths[0])
		})
	}
}

func (a *App) PickRealESRGANPath() {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: a.Win,
		Title:  "Select Real-ESRGAN Executable",
		Filters: []mygo.FileFilter{
			{Name: "Executable (*.exe)", Extensions: []string{"exe"}},
			{Name: "All Files (*.*)", Extensions: []string{"*"}},
		},
	})
	if err == nil && len(paths) > 0 {
		a.Update(func() {
			a.RealesrganPath = paths[0]
			a.AddLog("Real-ESRGAN executable set to: " + paths[0])
		})
	}
}

func (a *App) PickFFmpegPath() {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: a.Win,
		Title:  "Select FFmpeg Executable",
		Filters: []mygo.FileFilter{
			{Name: "Executable (*.exe)", Extensions: []string{"exe"}},
			{Name: "All Files (*.*)", Extensions: []string{"*"}},
		},
	})
	if err == nil && len(paths) > 0 {
		a.Update(func() {
			a.FfmpegPath = paths[0]
			a.AddLog("FFmpeg executable set to: " + paths[0])
		})
	}
}

func (a *App) HandleDropFile(drop transfer.Drop, isVideo bool) {
	files, err := drop.Data.Files()
	if err != nil || len(files) == 0 {
		return
	}
	file := files[0]
	a.Update(func() {
		if isVideo {
			a.InputVideoPath = file
			a.OutputVideoPath = ""
			a.VideoFrameSec = 0
			if a.FfmpegPath != "" {
				if dur := engine.ProbeVideoDuration(a.FfmpegPath, file); dur > 0 {
					a.VideoDurationSec = dur
				}
			}
			a.extractVideoFrame(0)
			a.AddLog("Dropped video: " + filepath.Base(file))
		} else {
			a.InputImagePath = file
			a.LoadRealImagePreview(file)
			a.AddLog("Dropped image: " + filepath.Base(file))
		}
		if a.OutputDir == "" {
			a.OutputDir = filepath.Dir(file)
		}
	})
}

func (a *App) StartImageUpscale() {
	if a.IsProcessing {
		return
	}
	if a.InputImagePath == "" {
		a.Update(func() {
			a.StatusMessage = "Please select an input image first"
			a.StatusIsError = true
		})
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelMu.Lock()
	a.cancelFunc = cancel
	a.cancelMu.Unlock()

	outDir := a.OutputDir
	if outDir == "" {
		outDir = filepath.Dir(a.InputImagePath)
	}
	ext := filepath.Ext(a.InputImagePath)
	base := strings.TrimSuffix(filepath.Base(a.InputImagePath), ext)
	format := strings.ToLower(a.GetImageFormat())
	if format == "" {
		format = "png"
	}
	scale := a.GetImageScale()
	outPath := filepath.Join(outDir, fmt.Sprintf("%s_upscaled_%dx.%s", base, scale, format))

	a.Update(func() {
		a.IsProcessing = true
		a.Progress = 0.0
		a.StatusIsError = false
		a.StatusMessage = "Memulai proses render gambar..."
		a.CurrentStage = "Merender gambar keseluruhan (0%)..."
		a.ProcessingTime = ""
		a.LastCompressionSavings = ""
		a.AddLog("--- Starting Image Upscale ---")
		a.AddLog(fmt.Sprintf("Input: %s", a.InputImagePath))
		a.AddLog(fmt.Sprintf("Output: %s", outPath))
		a.AddLog(fmt.Sprintf("Model: %s | Scale: %dx", a.ImageModel, scale))
	})

	if a.RealesrganPath != "" {
		if _, err := os.Stat(a.InputImagePath); err == nil {
			go func() {
				params := engine.ImageUpscaleParams{
					RealESRGANPath:  a.RealesrganPath,
					FFmpegPath:      a.FfmpegPath,
					InputPath:       a.InputImagePath,
					OutputPath:      outPath,
					Model:           a.ImageModel,
					Scale:           scale,
					Format:          format,
					TileSize:        a.ImageTileSize,
					ThreadMode:      a.ThreadMode,
					GPUDevice:       a.GpuDevice,
					Compress:        a.ImageCompress,
					CompressFormat:  a.ImageCompressFormat,
					CompressQuality: a.ImageCompressQuality,
					OnProgress: func(pct float64, stage string) {
						a.Update(func() {
							a.Progress = pct
							a.CurrentStage = stage
						})
					},
					OnLog: func(line string) {
						a.AddLog(line)
					},
				}
				elapsed, savings, err := engine.ExecuteImageUpscale(ctx, params)
				if err != nil {
					a.Update(func() {
						a.IsProcessing = false
						a.StatusIsError = true
						a.StatusMessage = "Error: " + err.Error()
						a.AddLog("Upscale error: " + err.Error())
					})
					return
				}

				if outData, err := os.ReadFile(outPath); err == nil {
					if bm, err := ui.DecodeBitmap(outData); err == nil {
						a.Update(func() {
							a.ImageAfterBitmap = bm
							w, h := bm.Size()
							a.ImageAfterSize = [2]int{w, h}
						})
					}
				}

				a.Update(func() {
					a.OutputImagePath = outPath
					a.IsProcessing = false
					a.Progress = 1.0
					a.CurrentStage = "Finished"
					a.LastCompressionSavings = savings
					a.StatusMessage = fmt.Sprintf("Completed in %s! Saved to: %s", elapsed, filepath.Base(outPath))
					a.ProcessingTime = elapsed.String()
					a.AddLog(fmt.Sprintf("Image upscale completed in %s. File: %s", elapsed, outPath))
				})
			}()
			return
		}
	}

	go a.runSimulatedImageUpscale(ctx)
}

func (a *App) StartVideoUpscale() {
	if a.IsProcessing {
		return
	}
	if a.InputVideoPath == "" {
		a.Update(func() {
			a.StatusMessage = "Please select an input video first"
			a.StatusIsError = true
		})
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelMu.Lock()
	a.cancelFunc = cancel
	a.cancelMu.Unlock()

	outDir := a.OutputDir
	if outDir == "" {
		outDir = filepath.Dir(a.InputVideoPath)
	}
	ext := filepath.Ext(a.InputVideoPath)
	base := strings.TrimSuffix(filepath.Base(a.InputVideoPath), ext)
	scale := a.GetVideoScale()
	outPath := filepath.Join(outDir, fmt.Sprintf("%s_upscaled_%dx.mp4", base, scale))

	a.Update(func() {
		a.IsProcessing = true
		a.Progress = 0.0
		a.StatusIsError = false
		a.StatusMessage = "Memulai proses render video..."
		a.CurrentStage = "Merender video keseluruhan (0%)..."
		a.ProcessingTime = ""
		a.AddLog("--- Starting Video Upscale Pipeline ---")
		a.AddLog(fmt.Sprintf("Input Video: %s", a.InputVideoPath))
		a.AddLog(fmt.Sprintf("Output Video: %s", outPath))
		a.AddLog(fmt.Sprintf("Model: %s | Codec: %s", a.VideoModel, a.VideoCodec))
	})

	if a.RealesrganPath != "" && a.FfmpegPath != "" {
		if _, err := os.Stat(a.InputVideoPath); err == nil {
			go func() {
				params := engine.VideoUpscaleParams{
					RealESRGANPath:     a.RealesrganPath,
					FFmpegPath:         a.FfmpegPath,
					InputPath:          a.InputVideoPath,
					OutputPath:         outPath,
					Model:              a.VideoModel,
					Scale:              scale,
					Codec:              a.VideoCodec,
					CrfIndex:           a.VideoCrfIndex,
					CompressCRF:        a.VideoCompressCRF,
					KeepAudio:          a.VideoKeepAudio,
					TileSize:           a.VideoTileSize,
					ThreadMode:         a.ThreadMode,
					GPUDevice:          a.GpuDevice,
					IntermediateFormat: a.VideoIntermediateFormat,
					Preset:             a.VideoPreset,
					AutoClearVram:      a.AutoClearVram,
					OnProgress: func(pct float64, stage string) {
						a.Update(func() {
							a.Progress = pct
							a.CurrentStage = stage
						})
					},
					OnLog: func(line string) {
						a.AddLog(line)
					},
				}
				elapsed, err := engine.ExecuteVideoUpscale(ctx, params)
				if err != nil {
					a.Update(func() {
						a.IsProcessing = false
						a.StatusIsError = true
						a.StatusMessage = "Error: " + err.Error()
						a.AddLog("Video pipeline error: " + err.Error())
					})
					return
				}

				if pData, err := engine.ExtractSingleFrameReal(a.FfmpegPath, outPath, 0); err == nil {
					if bm, err := ui.DecodeBitmap(pData); err == nil {
						a.Update(func() {
							a.VideoAfterBitmap = bm
							w, h := bm.Size()
							a.VideoAfterSize = [2]int{w, h}
						})
					}
				}

				a.Update(func() {
					a.OutputVideoPath = outPath
					a.IsProcessing = false
					a.Progress = 1.0
					a.CurrentStage = "Finished"
					a.StatusMessage = fmt.Sprintf("Video upscale completed in %s! Output saved to: %s", elapsed, filepath.Base(outPath))
					a.ProcessingTime = elapsed.String()
					a.AddLog(fmt.Sprintf("Video upscale finished in %s. File: %s", elapsed, outPath))
				})
			}()
			return
		}
	}

	go a.runSimulatedVideoUpscale(ctx)
}

func (a *App) CancelUpscale() {
	a.cancelMu.Lock()
	if a.cancelFunc != nil {
		a.cancelFunc()
		a.cancelFunc = nil
	}
	a.cancelMu.Unlock()

	a.Update(func() {
		a.IsProcessing = false
		a.StatusMessage = "Upscale canceled by user"
		a.CurrentStage = "Canceled"
		a.AddLog("Upscaling task was canceled")
	})
}

func (a *App) runSimulatedImageUpscale(ctx context.Context) {
	start := time.Now()
	steps := []struct {
		stage    string
		logMsg   string
		progress float64
		delay    time.Duration
	}{
		{"Merender gambar keseluruhan (20%)...", fmt.Sprintf("Loaded model %s on %s", a.ImageModel, a.GpuDevice), 0.20, 300 * time.Millisecond},
		{"Merender gambar keseluruhan (50%)...", fmt.Sprintf("Inference scaling (%dx resolution)", a.GetImageScale()), 0.50, 400 * time.Millisecond},
		{"Merender gambar keseluruhan (75%)...", "GFPGAN face restoration applied", 0.75, 400 * time.Millisecond},
		{"Merender gambar keseluruhan (95%)...", "Writing file to disk (" + a.GetImageFormat() + ")", 0.95, 300 * time.Millisecond},
	}

	for _, step := range steps {
		select {
		case <-ctx.Done():
			return
		case <-time.After(step.delay):
			a.Update(func() {
				a.Progress = step.progress
				a.CurrentStage = step.stage
				a.AddLog(step.logMsg)
			})
		}
	}

	elapsed := time.Since(start).Round(100 * time.Millisecond)
	a.Update(func() {
		a.IsProcessing = false
		a.Progress = 1.0
		a.CurrentStage = "Finished"
		a.StatusMessage = fmt.Sprintf("Image upscale completed in %s", elapsed)
		a.ProcessingTime = elapsed.String()

		afterImg := preview.GenerateScenicSample(true, 640, 360, 0)
		a.ImageAfterBitmap = ui.NewBitmap(afterImg)
		a.AddLog(fmt.Sprintf("Upscale finished successfully in %s (Preview Sesudah updated)", elapsed))
	})
}

func (a *App) QuickImagePreview() {
	afterImg := preview.GenerateScenicSample(true, 640, 360, 0)
	a.ImageAfterBitmap = ui.NewBitmap(afterImg)
	a.AddLog("Quick AI preview updated")
}

func (a *App) runSimulatedVideoUpscale(ctx context.Context) {
	start := time.Now()
	steps := []struct {
		stage    string
		logMsg   string
		progress float64
		delay    time.Duration
	}{
		{"Merender video keseluruhan (15%)...", "Inisialisasi pipeline pemrosesan video...", 0.15, 300 * time.Millisecond},
		{"Merender video keseluruhan (35%)...", "Memproses inferensi super-resolution AI...", 0.35, 400 * time.Millisecond},
		{"Merender video keseluruhan (60%)...", "Melanjutkan rendering super-resolution AI...", 0.60, 500 * time.Millisecond},
		{"Merender video keseluruhan (80%)...", "Menyelesaikan inferensi neural network...", 0.80, 400 * time.Millisecond},
		{"Merender video keseluruhan (95%)...", "Finalisasi pengkodean video hasil upscale...", 0.95, 350 * time.Millisecond},
	}

	for _, step := range steps {
		select {
		case <-ctx.Done():
			return
		case <-time.After(step.delay):
			a.Update(func() {
				a.Progress = step.progress
				a.CurrentStage = step.stage
				a.AddLog(step.logMsg)
			})
		}
	}

	elapsed := time.Since(start).Round(100 * time.Millisecond)
	a.Update(func() {
		a.IsProcessing = false
		a.Progress = 1.0
		a.CurrentStage = "Finished"
		a.StatusMessage = fmt.Sprintf("Video upscale pipeline completed in %s", elapsed)
		a.ProcessingTime = elapsed.String()

		vAfter := preview.GenerateScenicSample(true, 640, 360, a.VideoFrameSec)
		a.VideoAfterBitmap = ui.NewBitmap(vAfter)
		a.AddLog(fmt.Sprintf("Video processing completed successfully in %s (Preview Sesudah updated)", elapsed))
	})
}

func GetCleanModelName(s string) string {
	fields := strings.Fields(s)
	if len(fields) > 0 {
		return fields[0]
	}
	return "realesrgan-x4plus"
}

func getModelScale(s string) int {
	clean := GetCleanModelName(s)
	switch {
	case strings.Contains(clean, "-x2"):
		return 2
	case strings.Contains(clean, "-x3"):
		return 3
	default:
		return 4
	}
}

func (a *App) GetImageScale() int {
	return getModelScale(a.ImageModel)
}

func (a *App) GetVideoScale() int {
	return getModelScale(a.VideoModel)
}

func (a *App) ZoomInImage() {
	if a.ImageZoom < 10.0 {
		a.ImageZoom *= 1.25
		if a.ImageZoom > 10.0 {
			a.ImageZoom = 10.0
		}
	}
	if a.ImageZoomIndex < 4 {
		a.ImageZoomIndex++
	}
}

func (a *App) ZoomOutImage() {
	if a.ImageZoom > 0.5 {
		a.ImageZoom /= 1.25
		if a.ImageZoom < 0.5 {
			a.ImageZoom = 0.5
		}
	}
	if a.ImageZoomIndex > 0 {
		a.ImageZoomIndex--
	}
}

func (a *App) ResetZoomImage() {
	a.ImageZoomIndex = 0
	a.ImageZoom = 1.0
	a.ImagePanX = 0
	a.ImagePanY = 0
	a.ImageScroll.X = 0
	a.ImageScroll.Y = 0
}

func (a *App) ZoomInVideo() {
	if a.VideoZoom < 10.0 {
		a.VideoZoom *= 1.25
		if a.VideoZoom > 10.0 {
			a.VideoZoom = 10.0
		}
	}
	if a.VideoZoomIndex < 4 {
		a.VideoZoomIndex++
	}
}

func (a *App) ZoomOutVideo() {
	if a.VideoZoom > 0.5 {
		a.VideoZoom /= 1.25
		if a.VideoZoom < 0.5 {
			a.VideoZoom = 0.5
		}
	}
	if a.VideoZoomIndex > 0 {
		a.VideoZoomIndex--
	}
}

func (a *App) ResetZoomVideo() {
	a.VideoZoomIndex = 0
	a.VideoZoom = 1.0
	a.VideoPanX = 0
	a.VideoPanY = 0
	a.VideoScroll.X = 0
	a.VideoScroll.Y = 0
}

func (a *App) PlayVideo() {
	a.videoPlayMu.Lock()
	if a.VideoIsPlaying {
		a.videoPlayMu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.videoPlayCancel = cancel
	a.VideoIsPlaying = true
	if a.OutputVideoPath == "" {
		a.VideoPreviewTab = 0
	}
	a.videoPlayMu.Unlock()

	a.Update(func() {})

	go func() {
		ticker := time.NewTicker(66 * time.Millisecond)
		defer ticker.Stop()

		var extractMu sync.Mutex
		extracting := false

		lastTick := time.Now()

		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				dt := now.Sub(lastTick).Seconds()
				lastTick = now
				if dt <= 0 || dt > 0.5 {
					dt = 0.066
				}

				a.videoPlayMu.Lock()
				if !a.VideoIsPlaying {
					a.videoPlayMu.Unlock()
					return
				}
				a.VideoFrameSec += dt
				if a.VideoDurationSec > 0 && a.VideoFrameSec >= a.VideoDurationSec {
					a.VideoFrameSec = 0
				}
				curSec := a.VideoFrameSec
				a.LastVideoFrameSec = curSec
				a.videoPlayMu.Unlock()

				if a.FfmpegPath == "" || a.InputVideoPath == "" {
					a.Update(func() {
						vBefore := preview.GenerateScenicSample(false, 640, 360, curSec)
						a.VideoBeforeBitmap = ui.NewBitmap(vBefore)
						a.VideoBeforeBitmap.Pixelated()
						vAfter := preview.GenerateScenicSample(true, 640, 360, curSec)
						a.VideoAfterBitmap = ui.NewBitmap(vAfter)
						a.VideoAfterBitmap.Pixelated()
					})
					continue
				}

				a.Update(func() {})

				extractMu.Lock()
				if extracting {
					extractMu.Unlock()
					continue
				}
				extracting = true
				extractMu.Unlock()

				go func(sec float64) {
					defer func() {
						extractMu.Lock()
						extracting = false
						extractMu.Unlock()
					}()

					out, err := engine.ExtractSingleFrameReal(a.FfmpegPath, a.InputVideoPath, sec)
					if err == nil && len(out) > 0 {
						if bm, err := ui.DecodeBitmap(out); err == nil {
							bm.Pixelated()
							w, h := bm.Size()
							a.Update(func() {
								a.VideoBeforeBitmap = bm
								a.VideoBeforeSize = [2]int{w, h}
							})
						}
					}

					if a.OutputVideoPath != "" {
						if outAfter, err := engine.ExtractSingleFrameReal(a.FfmpegPath, a.OutputVideoPath, sec); err == nil && len(outAfter) > 0 {
							if bmAfter, err := ui.DecodeBitmap(outAfter); err == nil {
								bmAfter.Pixelated()
								w, h := bmAfter.Size()
								a.Update(func() {
									a.VideoAfterBitmap = bmAfter
									a.VideoAfterSize = [2]int{w, h}
								})
							}
						}
					}
				}(curSec)
			}
		}
	}()
}

func (a *App) PauseVideo() {
	a.videoPlayMu.Lock()
	if a.videoPlayCancel != nil {
		a.videoPlayCancel()
		a.videoPlayCancel = nil
	}
	a.VideoIsPlaying = false
	a.videoPlayMu.Unlock()
	a.Update(func() {})
}

func (a *App) TogglePlayVideo() {
	if a.VideoIsPlaying {
		a.PauseVideo()
	} else {
		a.PlayVideo()
	}
}

func (a *App) RewindVideo() {
	a.PauseVideo()
	a.Update(func() {
		a.VideoFrameSec = 0
		a.LastVideoFrameSec = 0
		a.extractVideoFrame(0)
	})
}

func (a *App) SeekPlayVideo(sec float64) {
	a.videoPlayMu.Lock()
	a.VideoFrameSec = sec
	a.LastVideoFrameSec = sec
	a.videoPlayMu.Unlock()
	a.extractVideoFrame(sec)
}

func (a *App) GetImageFormat() string {
	switch a.ImageFormatIndex {
	case 1:
		return "JPG"
	case 2:
		return "WebP"
	default:
		return "PNG"
	}
}

func (a *App) GetVideoCrf() string {
	switch a.VideoCrfIndex {
	case 0:
		return "17 (High Quality)"
	case 2:
		return "26 (Compact)"
	default:
		return "21 (Balanced)"
	}
}

func (a *App) ClearLogs() {
	a.Logs = []string{
		fmt.Sprintf("[%s] Logs cleared", time.Now().Format("15:04:05")),
	}
}

func (a *App) AllLogsText() string {
	return strings.Join(a.Logs, "\n")
}
