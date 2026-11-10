package main

import (
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"starora/internal/app"
	"starora/internal/engine"
)

func TestViewInitialization(t *testing.T) {
	a := app.New()
	tt := ui.NewTester(a.View, 1120, 760)

	if !tt.HasText("Starora AI Upscaler") {
		t.Errorf("expected title 'Starora AI Upscaler', got texts: %v", tt.Texts())
	}
	if !tt.HasText("Pilihan Model AI Real-ESRGAN") {
		t.Errorf("expected model selection card")
	}
	if !tt.HasText("Start Image Upscale") {
		t.Errorf("expected button 'Start Image Upscale'")
	}
	if !tt.HasText("Preview: Sebelum vs Sesudah (Visual Comparison)") {
		t.Errorf("expected visual comparison card")
	}
}

func TestTabNavigation(t *testing.T) {
	a := app.New()
	tt := ui.NewTester(a.View, 1120, 760)

	if err := tt.Click("Video Upscale"); err != nil {
		t.Fatalf("failed to click Video Upscale tab: %v", err)
	}
	tt.Frame()
	if !tt.HasText("Start Video Upscale") {
		t.Errorf("expected Video Upscale view content")
	}

	if err := tt.Click("Engine Settings"); err != nil {
		t.Fatalf("failed to click Engine Settings tab: %v", err)
	}
	tt.Frame()
	if !tt.HasText("Real-ESRGAN & FFmpeg Engine Status") {
		t.Errorf("expected Engine Settings view content")
	}

	if err := tt.Click("Activity Logs"); err != nil {
		t.Fatalf("failed to click Activity Logs tab: %v", err)
	}
	tt.Frame()
	if !tt.HasText("Application Activity & Console Logs") {
		t.Errorf("expected Activity Logs view content")
	}

	if err := tt.Click("About"); err != nil {
		t.Fatalf("failed to click About tab: %v", err)
	}
	tt.Frame()
	if !tt.HasText("Starora AI Studio") {
		t.Errorf("expected About view content")
	}

	if err := tt.Click("Image Upscale"); err != nil {
		t.Fatalf("failed to click Image Upscale tab: %v", err)
	}
	tt.Frame()
	if !tt.HasText("Start Image Upscale") {
		t.Errorf("expected Image Upscale view content")
	}
}

func TestAboutView(t *testing.T) {
	a := app.New()
	tt := ui.NewTester(a.View, 1120, 1100)

	if err := tt.Click("About"); err != nil {
		t.Fatalf("failed to click About tab: %v", err)
	}
	tt.Frame()

	if !tt.HasText("Starora AI Studio") {
		t.Errorf("expected app title in About view")
	}
	if !tt.HasText("GitHub: mphstar/starora") {
		t.Errorf("expected GitHub card")
	}
	if !tt.HasText("Ucapan Terima Kasih (Special Thanks & Sumber Terbuka)") {
		t.Errorf("expected Special Thanks section")
	}

	// Click external action buttons
	if err := tt.Click("Buka GitHub"); err != nil {
		t.Fatalf("failed to click Buka GitHub: %v", err)
	}
	if err := tt.Click("Sumber: GitHub Real-ESRGAN"); err != nil {
		t.Fatalf("failed to click Sumber Real-ESRGAN: %v", err)
	}
	if err := tt.Click("Sumber: FFmpeg.org"); err != nil {
		t.Fatalf("failed to click Sumber FFmpeg: %v", err)
	}
}

func TestImageUpscaleValidationAndCancel(t *testing.T) {
	a := app.New()
	tt := ui.NewTester(a.View, 1120, 760)

	if err := tt.Click("Start Image Upscale"); err != nil {
		t.Fatalf("failed to click Start Image Upscale: %v", err)
	}
	if !a.StatusIsError {
		t.Errorf("expected statusIsError to be true when no input image is provided")
	}

	if err := tt.Click("Input Image Path"); err != nil {
		t.Fatalf("failed to click Input Image Path: %v", err)
	}
	tt.Type("sample.png")
	tt.Frame()

	if err := tt.Click("Start Image Upscale"); err != nil {
		t.Fatalf("failed to click Start Image Upscale with input: %v", err)
	}

	if !a.IsProcessing {
		t.Errorf("expected isProcessing to be true")
	}

	if err := tt.Click("Cancel"); err != nil {
		t.Fatalf("failed to click Cancel: %v", err)
	}

	if a.IsProcessing {
		t.Errorf("expected isProcessing to be false after Cancel")
	}
}

func TestLogsClearing(t *testing.T) {
	a := app.New()
	a.AddLog("test message 1")
	a.AddLog("test message 2")

	if len(a.Logs) < 2 {
		t.Errorf("expected logs to have entries")
	}

	a.ClearLogs()
	if len(a.Logs) != 1 {
		t.Errorf("expected logs to be reset, got count %d", len(a.Logs))
	}
}

func TestBeforeAfterPreviews(t *testing.T) {
	a := app.New()
	tt := ui.NewTester(a.View, 1120, 760)

	if !tt.HasText("Preview: Sebelum vs Sesudah (Visual Comparison)") {
		t.Errorf("expected image visual comparison card")
	}
	if !tt.HasText("SEBELUM (ORIGINAL)") {
		t.Errorf("expected 'SEBELUM (ORIGINAL)' label")
	}

	if err := tt.Click("Sesudah"); err != nil {
		t.Fatalf("failed to switch to Sesudah tab: %v", err)
	}
	tt.Frame()
	if a.ImagePreviewTab != 1 {
		t.Errorf("expected imagePreviewTab to be 1, got %d", a.ImagePreviewTab)
	}
	if !tt.HasText("SESUDAH (REAL-ESRGAN HD)") {
		t.Errorf("expected 'SESUDAH (REAL-ESRGAN HD)' label after switching to Sesudah")
	}

	if err := tt.Click("Sebelum"); err != nil {
		t.Fatalf("failed to switch to Sebelum tab: %v", err)
	}
	tt.Frame()
	if a.ImagePreviewTab != 0 {
		t.Errorf("expected imagePreviewTab to be 0, got %d", a.ImagePreviewTab)
	}

	if err := tt.Click("+"); err != nil {
		t.Fatalf("failed to click zoom in: %v", err)
	}
	tt.Frame()
	if a.ImageZoom <= 1.0 {
		t.Errorf("expected imageZoom to be > 1.0 after zoom in, got %f", a.ImageZoom)
	}

	if err := tt.Click("Reset"); err != nil {
		t.Fatalf("failed to click Reset zoom: %v", err)
	}
	tt.Frame()
	if a.ImageZoom != 1.0 {
		t.Errorf("expected imageZoom to be 1.0 after reset, got %f", a.ImageZoom)
	}

	if err := tt.Click("Quick AI Preview"); err != nil {
		t.Fatalf("failed to click Quick AI Preview: %v", err)
	}
	tt.Frame()
	if a.ImageAfterBitmap == nil {
		t.Errorf("expected imageAfterBitmap to be non-nil")
	}

	if err := tt.Click("Video Upscale"); err != nil {
		t.Fatalf("failed to click Video Upscale: %v", err)
	}
	tt.Frame()

	if !tt.HasText("Preview Frame Video: Sebelum vs Sesudah") {
		t.Errorf("expected video visual comparison card")
	}
	if !tt.HasText("FRAME SEBELUM (ORIGINAL)") {
		t.Errorf("expected 'FRAME SEBELUM (ORIGINAL)' label")
	}
	if !tt.HasText("Ambil Frame") {
		t.Errorf("expected 'Ambil Frame' button")
	}

	if !tt.HasText("▶ Play") {
		t.Errorf("expected '▶ Play' button in video timeline")
	}
	if err := tt.Click("▶ Play"); err != nil {
		t.Fatalf("failed to click Play: %v", err)
	}
	tt.Frame()
	if !a.VideoIsPlaying {
		t.Errorf("expected videoIsPlaying to be true after clicking Play")
	}
	if err := tt.Click("⏸ Pause"); err != nil {
		t.Fatalf("failed to click Pause: %v", err)
	}
	tt.Frame()
	if a.VideoIsPlaying {
		t.Errorf("expected videoIsPlaying to be false after clicking Pause")
	}

	if !tt.HasText("Progres Render Video Keseluruhan") {
		t.Errorf("expected 'Progres Render Video Keseluruhan' card")
	}
}

func TestSplitStudioLayout(t *testing.T) {
	a := app.New()
	tt := ui.NewTester(a.View, 1120, 760)

	// Canvas and inspector are simultaneously visible
	if !tt.HasText("Preview: Sebelum vs Sesudah (Visual Comparison)") {
		t.Errorf("expected preview canvas on left")
	}
	if !tt.HasText("Pilihan Model AI Real-ESRGAN") {
		t.Errorf("expected model selection in inspector on right")
	}
	if !tt.HasText("File Gambar Input & Lokasi Output") {
		t.Errorf("expected file source card in inspector")
	}
	if !tt.HasText("Progres Render Gambar Keseluruhan") {
		t.Errorf("expected render card in inspector")
	}

	// Verify default format is PNG
	if a.GetImageFormat() != "PNG" {
		t.Errorf("expected default image format to be PNG, got %s", a.GetImageFormat())
	}

	// Verify switching tabs in image format
	if err := tt.Click("JPG"); err != nil {
		t.Fatalf("failed to click JPG format: %v", err)
	}
	tt.Frame()
	if a.ImageFormatIndex != 1 {
		t.Errorf("expected imageFormatIndex to be 1, got %d", a.ImageFormatIndex)
	}

	// Switch to video and check video split layout
	if err := tt.Click("Video Upscale"); err != nil {
		t.Fatalf("failed to switch to Video Upscale: %v", err)
	}
	tt.Frame()
	if !tt.HasText("Preview Frame Video: Sebelum vs Sesudah") {
		t.Errorf("expected video preview canvas on left")
	}
	if !tt.HasText("Pilihan Model AI Video Real-ESRGAN & Audio") {
		t.Errorf("expected video model card in inspector on right")
	}
}

func TestPerformanceResolvers(t *testing.T) {
	turbo := engine.ResolveThreads("Turbo Multi-Thread (Otomatis Maksimal)")
	if turbo == "" {
		t.Errorf("expected non-empty turbo thread config")
	}

	balanced := engine.ResolveThreads("Seimbang (Balanced 2:2:2)")
	if balanced != "2:2:2" {
		t.Errorf("expected 2:2:2 for balanced, got %s", balanced)
	}

	standard := engine.ResolveThreads("Standard (1:2:1)")
	if standard != "1:2:1" {
		t.Errorf("expected 1:2:1 for standard, got %s", standard)
	}

	if p := engine.ResolvePreset("veryfast (Super Cepat)"); p != "veryfast" {
		t.Errorf("expected veryfast preset, got %s", p)
	}
	if p := engine.ResolvePreset("fast (Cepat & Seimbang)"); p != "fast" {
		t.Errorf("expected fast preset, got %s", p)
	}
	if p := engine.ResolvePreset("medium (Kualitas Tinggi)"); p != "medium" {
		t.Errorf("expected medium preset, got %s", p)
	}

	// Compression tests
	if b := engine.FormatBytes(500); b != "500 B" {
		t.Errorf("expected '500 B', got %s", b)
	}
	if b := engine.FormatBytes(2048); b != "2.0 KB" {
		t.Errorf("expected '2.0 KB', got %s", b)
	}
	if b := engine.FormatBytes(10485760); b != "10.0 MB" {
		t.Errorf("expected '10.0 MB', got %s", b)
	}

	if crf := engine.ResolveVideoCRF("CRF 28 (Kompresi Tinggi)", 1); crf != "28" {
		t.Errorf("expected CRF 28, got %s", crf)
	}
	if crf := engine.ResolveVideoCRF("CRF 23 (Seimbang)", 1); crf != "23" {
		t.Errorf("expected CRF 23, got %s", crf)
	}
	if crf := engine.ResolveVideoCRF("CRF 18 (Detail Maksimal)", 1); crf != "18" {
		t.Errorf("expected CRF 18, got %s", crf)
	}
}

func TestZoomScrollAndDrag(t *testing.T) {
	a := app.New()

	if a.ImageZoom != 1.0 || a.ImagePanX != 0 || a.ImagePanY != 0 {
		t.Errorf("initial image zoom/pan should be 1.0, 0, 0")
	}

	a.ZoomInImage()
	if a.ImageZoom <= 1.0 {
		t.Errorf("expected zoom in to increase imageZoom, got %f", a.ImageZoom)
	}

	a.ImagePanX += 50
	a.ImagePanY -= 20
	if a.ImagePanX != 50 || a.ImagePanY != -20 {
		t.Errorf("expected pan offset to update")
	}

	a.ResetZoomImage()
	if a.ImageZoom != 1.0 || a.ImagePanX != 0 || a.ImagePanY != 0 {
		t.Errorf("expected reset to clear zoom and pan")
	}

	a.ZoomInVideo()
	if a.VideoZoom <= 1.0 {
		t.Errorf("expected zoom in to increase videoZoom, got %f", a.VideoZoom)
	}

	a.VideoPanX += 30
	a.VideoPanY += 15
	a.ResetZoomVideo()
	if a.VideoZoom != 1.0 || a.VideoPanX != 0 || a.VideoPanY != 0 {
		t.Errorf("expected reset to clear video zoom and pan")
	}
}

func TestVideoTimelineRewindAndSeek(t *testing.T) {
	a := app.New()

	a.VideoFrameSec = 15.0
	a.LastVideoFrameSec = 15.0

	a.PlayVideo()
	if !a.VideoIsPlaying {
		t.Errorf("expected videoIsPlaying to be true")
	}

	a.PauseVideo()
	if a.VideoIsPlaying {
		t.Errorf("expected videoIsPlaying to be false")
	}

	a.RewindVideo()
	if a.VideoFrameSec != 0 || a.LastVideoFrameSec != 0 {
		t.Errorf("expected rewind to reset videoFrameSec to 0, got %f", a.VideoFrameSec)
	}

	durStr := engine.FormatDuration(65.5)
	if durStr != "01:05.5" {
		t.Errorf("expected formatDuration(65.5) to be '01:05.5', got %s", durStr)
	}
}

func TestOverallRenderProgress(t *testing.T) {
	a := app.New()
	tt := ui.NewTester(a.View, 1120, 760)

	if !tt.HasText("Progres Render Gambar Keseluruhan") {
		t.Errorf("expected 'Progres Render Gambar Keseluruhan' card")
	}

	if err := tt.Click("Video Upscale"); err != nil {
		t.Fatalf("failed to switch to Video Upscale: %v", err)
	}
	tt.Frame()

	if !tt.HasText("Progres Render Video Keseluruhan") {
		t.Errorf("expected 'Progres Render Video Keseluruhan' card")
	}
	if !tt.HasText("Progres render dihitung secara langsung dan menyeluruh dari 0% hingga 100% selesai.") {
		t.Errorf("expected overall render explanation text")
	}
}

func TestDecodeJpegBitmap(t *testing.T) {
	time.Sleep(1 * time.Millisecond)
	testVid := `C:\Users\POLIJE\AppData\Local\Temp\test_video.mp4`
	if _, err := os.Stat(testVid); os.IsNotExist(err) {
		t.Skip("test video not found: " + testVid)
	}
	a := app.New()
	if a.FfmpegPath == "" {
		t.Skip("ffmpeg not found")
	}
	data, err := engine.ExtractSingleFrameReal(a.FfmpegPath, testVid, 1.0)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	bm, err := ui.DecodeBitmap(data)
	if err != nil {
		t.Fatalf("DecodeBitmap failed on extracted jpeg: %v", err)
	}
	w, h := bm.Size()
	t.Logf("Extracted frame size: %d x %d", w, h)
	if w <= 0 || h <= 0 {
		t.Errorf("invalid bitmap size: %d x %d", w, h)
	}
}

func TestVideoPlaybackWithSeeking(t *testing.T) {
	a := app.New()
	testVid := `C:\Users\POLIJE\AppData\Local\Temp\test_video.mp4`
	a.InputVideoPath = testVid
	a.VideoDurationSec = 3.0
	a.VideoFrameSec = 0.0

	tt := ui.NewTester(a.View, 1120, 760)
	if err := tt.Click("Video Upscale"); err != nil {
		t.Fatalf("failed to click Video Upscale: %v", err)
	}
	tt.Frame()

	if err := tt.Click("▶ Play"); err != nil {
		t.Fatalf("failed to click Play: %v", err)
	}
	tt.Frame()

	time.Sleep(200 * time.Millisecond)
	tt.Frame()

	if a.VideoFrameSec <= 0.05 {
		t.Errorf("expected videoFrameSec to advance during playback, got %f", a.VideoFrameSec)
	}

	a.SeekPlayVideo(1.5)
	if a.VideoFrameSec != 1.5 {
		t.Errorf("expected videoFrameSec to seek to 1.5, got %f", a.VideoFrameSec)
	}

	time.Sleep(150 * time.Millisecond)
	tt.Frame()

	if a.VideoFrameSec <= 1.5 {
		t.Errorf("expected playback to continue advancing after seeking, got %f", a.VideoFrameSec)
	}

	if err := tt.Click("⏸ Pause"); err != nil {
		t.Fatalf("failed to click Pause: %v", err)
	}
	tt.Frame()

	if a.VideoIsPlaying {
		t.Errorf("expected playback to stop after clicking Pause")
	}
}
