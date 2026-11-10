package engine

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// VideoUpscaleParams contains all configuration required to upscale a video.
type VideoUpscaleParams struct {
	RealESRGANPath     string
	FFmpegPath         string
	InputPath          string
	OutputPath         string
	Model              string
	Scale              int
	Codec              string
	CrfIndex           int
	CompressCRF        string
	KeepAudio          bool
	TileSize           string
	ThreadMode         string
	GPUDevice          string
	IntermediateFormat string
	Preset             string
	AutoClearVram      bool
	OnProgress         func(pct float64, stage string)
	OnLog              func(line string)
}

// ProbeVideoFPS detects the video frames per second.
func ProbeVideoFPS(ffmpegPath, inputPath string) string {
	cmd := exec.Command(ffmpegPath, "-i", inputPath)
	SetHideWindow(cmd)
	out, _ := cmd.CombinedOutput()
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*fps`)
	matches := re.FindSubmatch(out)
	if len(matches) > 1 {
		return string(matches[1])
	}
	return "30"
}

// ProbeVideoDuration detects the total duration of the video in seconds.
func ProbeVideoDuration(ffmpegPath, inputPath string) float64 {
	cmd := exec.Command(ffmpegPath, "-i", inputPath)
	SetHideWindow(cmd)
	out, _ := cmd.CombinedOutput()
	re := regexp.MustCompile(`Duration:\s*(\d+):(\d+):(\d+(?:\.\d+)?)`)
	matches := re.FindSubmatch(out)
	if len(matches) > 3 {
		var h, m, s float64
		fmt.Sscanf(string(matches[1]), "%f", &h)
		fmt.Sscanf(string(matches[2]), "%f", &m)
		fmt.Sscanf(string(matches[3]), "%f", &s)
		return h*3600 + m*60 + s
	}
	return 60.0
}

// ExtractSingleFrameReal extracts a single video frame as raw image data at a specific timestamp.
func ExtractSingleFrameReal(ffmpegPath, videoPath string, sec float64) ([]byte, error) {
	cmd := exec.Command(ffmpegPath, "-y", "-threads", "0", "-ss", fmt.Sprintf("%.2f", sec), "-i", videoPath, "-vframes", "1", "-f", "image2pipe", "-c:v", "mjpeg", "-")
	SetHideWindow(cmd)
	return cmd.Output()
}

// ExecuteVideoUpscale processes an entire video pipeline through FFmpeg and Real-ESRGAN.
func ExecuteVideoUpscale(ctx context.Context, p VideoUpscaleParams) (time.Duration, error) {
	start := time.Now()
	tempDir, err := os.MkdirTemp("", "starora_video_*")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if p.AutoClearVram {
			os.RemoveAll(tempDir)
		}
	}()

	framesInDir := filepath.Join(tempDir, "frames")
	framesOutDir := filepath.Join(tempDir, "upscaled_frames")
	audioPath := filepath.Join(tempDir, "audio.m4a")

	if err := os.MkdirAll(framesInDir, 0755); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(framesOutDir, 0755); err != nil {
		return 0, err
	}

	if p.OnProgress != nil {
		p.OnProgress(0.05, "Merender video keseluruhan (5%)...")
	}
	if p.OnLog != nil {
		p.OnLog("Extracting audio with FFmpeg...")
	}

	audioCmd := exec.CommandContext(ctx, p.FFmpegPath, "-y", "-i", p.InputPath, "-vn", "-c:a", "copy", audioPath)
	SetHideWindow(audioCmd)
	_ = audioCmd.Run()
	hasAudio := false
	if fi, err := os.Stat(audioPath); err == nil && fi.Size() > 0 {
		hasAudio = true
		if p.OnLog != nil {
			p.OnLog("Audio stream extracted successfully.")
		}
	} else if p.OnLog != nil {
		p.OnLog("No audio stream detected (video will be silent).")
	}

	if p.OnProgress != nil {
		p.OnProgress(0.15, "Merender video keseluruhan (15%)...")
	}
	if p.OnLog != nil {
		p.OnLog("Extracting frames with FFmpeg...")
	}

	isJpg := strings.Contains(strings.ToUpper(p.IntermediateFormat), "JPG")
	frameExt := "png"
	if isJpg {
		frameExt = "jpg"
	}
	framePattern := filepath.Join(framesInDir, fmt.Sprintf("frame_%%08d.%s", frameExt))

	var extractArgs []string
	if isJpg {
		extractArgs = []string{
			"-y",
			"-threads", "0",
			"-i", p.InputPath,
			"-q:v", "1",
			"-qmin", "1",
			"-qmax", "1",
			framePattern,
		}
	} else {
		extractArgs = []string{
			"-y",
			"-threads", "0",
			"-i", p.InputPath,
			"-qscale:v", "1",
			"-qmin", "1",
			"-qmax", "1",
			framePattern,
		}
	}
	extractCmd := exec.CommandContext(ctx, p.FFmpegPath, extractArgs...)
	SetHideWindow(extractCmd)
	if err := extractCmd.Run(); err != nil {
		return 0, fmt.Errorf("FFmpeg frame extraction failed: %w", err)
	}

	frameEntries, err := os.ReadDir(framesInDir)
	if err != nil || len(frameEntries) == 0 {
		return 0, fmt.Errorf("no frames were extracted from video")
	}
	totalFrames := len(frameEntries)
	if p.OnLog != nil {
		p.OnLog(fmt.Sprintf("Extracted %d total frames for upscaling.", totalFrames))
	}

	modelName := ResolveModelName(p.Model, p.Scale)
	modelsDir := filepath.Join(GetEngineRootDir(), "models")
	tileSize := ParseTileSize(p.TileSize)
	threads := ResolveThreads(p.ThreadMode)

	if p.OnProgress != nil {
		p.OnProgress(0.15, "Merender video keseluruhan (15%)...")
	}
	if p.OnLog != nil {
		p.OnLog(fmt.Sprintf("Starting Real-ESRGAN batch inference (%dx, Model: %s, Threads: %s)...", p.Scale, modelName, threads))
	}

	upscaleArgs := []string{
		"-i", framesInDir,
		"-o", framesOutDir,
		"-n", modelName,
		"-s", fmt.Sprintf("%d", p.Scale),
		"-m", modelsDir,
		"-t", tileSize,
		"-j", threads,
	}
	if isJpg {
		upscaleArgs = append(upscaleArgs, "-f", "jpg")
	} else {
		upscaleArgs = append(upscaleArgs, "-f", "png")
	}

	switch p.GPUDevice {
	case "CPU Only (-g -1)":
		upscaleArgs = append(upscaleArgs, "-g", "-1")
	case "GPU 0 (Primary Discrete GPU)":
		upscaleArgs = append(upscaleArgs, "-g", "0")
	case "GPU 1 (Secondary GPU)":
		upscaleArgs = append(upscaleArgs, "-g", "1")
	}

	upscaleCmd := exec.CommandContext(ctx, p.RealESRGANPath, upscaleArgs...)
	SetHideWindow(upscaleCmd)
	upscaleCmd.Dir = GetEngineRootDir()

	stdout, err := upscaleCmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	upscaleCmd.Stderr = upscaleCmd.Stdout

	if err := upscaleCmd.Start(); err != nil {
		return 0, err
	}

	doneTracking := make(chan struct{})
	ticker := time.NewTicker(200 * time.Millisecond)
	go func() {
		for {
			select {
			case <-doneTracking:
				return
			case <-ticker.C:
				entries, err := os.ReadDir(framesOutDir)
				if err == nil && totalFrames > 0 {
					completed := len(entries)
					overall := 0.15 + (float64(completed)/float64(totalFrames))*0.75
					if overall > 0.90 {
						overall = 0.90
					}
					if p.OnProgress != nil {
						p.OnProgress(overall, fmt.Sprintf("Merender video keseluruhan (%.0f%%)...", overall*100))
					}
				}
			}
		}
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && p.OnLog != nil {
			p.OnLog(line)
		}
	}

	waitErr := upscaleCmd.Wait()
	close(doneTracking)
	ticker.Stop()

	if waitErr != nil {
		return 0, fmt.Errorf("Real-ESRGAN video frame upscale failed: %w", waitErr)
	}

	fpsStr := ProbeVideoFPS(p.FFmpegPath, p.InputPath)
	codec := ResolveCodec(p.Codec)
	crfVal := ResolveVideoCRF(p.CompressCRF, p.CrfIndex)
	preset := ResolvePreset(p.Preset)

	if p.OnProgress != nil {
		p.OnProgress(0.90, "Merender video keseluruhan (90%)...")
	}
	if p.OnLog != nil {
		p.OnLog(fmt.Sprintf("Re-encoding video with FFmpeg (%s, preset: %s)...", codec, preset))
	}

	reencodeArgs := []string{
		"-y",
		"-r", fpsStr,
		"-threads", "0",
		"-i", filepath.Join(framesOutDir, fmt.Sprintf("frame_%%08d.%s", frameExt)),
	}

	if hasAudio && p.KeepAudio {
		reencodeArgs = append(reencodeArgs, "-i", audioPath, "-c:a", "copy")
	}

	reencodeArgs = append(reencodeArgs,
		"-c:v", codec,
		"-preset", preset,
		"-crf", crfVal,
		"-pix_fmt", "yuv420p",
		"-shortest",
		p.OutputPath,
	)

	muxCmd := exec.CommandContext(ctx, p.FFmpegPath, reencodeArgs...)
	SetHideWindow(muxCmd)
	if err := muxCmd.Run(); err != nil {
		return 0, fmt.Errorf("FFmpeg re-encoding failed: %w", err)
	}

	elapsed := time.Since(start).Round(100 * time.Millisecond)
	return elapsed, nil
}
