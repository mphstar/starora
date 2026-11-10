package engine

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ImageUpscaleParams contains all configuration required to upscale an image.
type ImageUpscaleParams struct {
	RealESRGANPath  string
	FFmpegPath      string
	InputPath       string
	OutputPath      string
	Model           string
	Scale           int
	Format          string
	TileSize        string
	ThreadMode      string
	GPUDevice       string
	Compress        bool
	CompressFormat  string
	CompressQuality int
	OnProgress      func(pct float64, stage string)
	OnLog           func(line string)
}

// ExecuteImageUpscale runs Real-ESRGAN and optional compression on an image.
func ExecuteImageUpscale(ctx context.Context, p ImageUpscaleParams) (time.Duration, string, error) {
	start := time.Now()
	modelName := ResolveModelName(p.Model, p.Scale)
	modelsDir := filepath.Join(GetEngineRootDir(), "models")
	tileSize := ParseTileSize(p.TileSize)
	format := strings.ToLower(p.Format)
	if format == "" {
		format = "png"
	}
	threads := ResolveThreads(p.ThreadMode)

	args := []string{
		"-i", p.InputPath,
		"-o", p.OutputPath,
		"-n", modelName,
		"-s", fmt.Sprintf("%d", p.Scale),
		"-m", modelsDir,
		"-t", tileSize,
		"-j", threads,
		"-f", format,
	}

	switch p.GPUDevice {
	case "CPU Only (-g -1)":
		args = append(args, "-g", "-1")
	case "GPU 0 (Primary Discrete GPU)":
		args = append(args, "-g", "0")
	case "GPU 1 (Secondary GPU)":
		args = append(args, "-g", "1")
	}

	cmd := exec.CommandContext(ctx, p.RealESRGANPath, args...)
	SetHideWindow(cmd)
	cmd.Dir = GetEngineRootDir()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, "", err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return 0, "", err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		cleanLine := strings.ReplaceAll(line, ",", ".")
		if strings.HasSuffix(cleanLine, "%") {
			var pct float64
			if _, err := fmt.Sscanf(cleanLine, "%f%%", &pct); err == nil {
				if p.OnProgress != nil {
					p.OnProgress(pct/100.0, fmt.Sprintf("Merender gambar keseluruhan (%.0f%%)...", pct))
				}
			}
		}
		if p.OnLog != nil {
			p.OnLog(line)
		}
	}

	if err := cmd.Wait(); err != nil {
		return 0, "", err
	}

	elapsed := time.Since(start).Round(100 * time.Millisecond)
	savings := ""

	// Post-upscale compression if enabled (defaulting to PNG or user choice)
	if p.Compress && p.FFmpegPath != "" {
		if fi, err := os.Stat(p.OutputPath); err == nil && fi.Size() > 0 {
			rawSize := fi.Size()
			if p.OnProgress != nil {
				p.OnProgress(0.98, "Mengompresi ukuran file gambar...")
			}
			if p.OnLog != nil {
				p.OnLog(fmt.Sprintf("Compressing output image with FFmpeg (original: %s)...", FormatBytes(rawSize)))
			}
			tempCompressed := filepath.Join(os.TempDir(), fmt.Sprintf("starora_cmp_%d%s", time.Now().UnixNano(), filepath.Ext(p.OutputPath)))
			if err := CompressImageWithFFmpeg(ctx, p.FFmpegPath, p.OutputPath, tempCompressed, p.CompressFormat, p.CompressQuality); err == nil {
				if fiCmp, err := os.Stat(tempCompressed); err == nil && fiCmp.Size() > 0 {
					cmpSize := fiCmp.Size()
					if data, err := os.ReadFile(tempCompressed); err == nil {
						_ = os.WriteFile(p.OutputPath, data, 0644)
						_ = os.Remove(tempCompressed)
					}
					savingsPct := float64(rawSize-cmpSize) / float64(rawSize) * 100
					if savingsPct > 0 {
						savings = fmt.Sprintf("Hemat %.1f%% (%s → %s)", savingsPct, FormatBytes(rawSize), FormatBytes(cmpSize))
						if p.OnLog != nil {
							p.OnLog(fmt.Sprintf("Kompresi selesai: %s", savings))
						}
					} else {
						savings = FormatBytes(cmpSize)
					}
				}
			}
		}
	}

	return elapsed, savings, nil
}

// UpscaleSingleFrameReal runs a single frame inference for quick video preview.
func UpscaleSingleFrameReal(realesrganPath, inFramePath, outFramePath string, scale int, modelName string) error {
	modelsDir := filepath.Join(GetEngineRootDir(), "models")
	resolved := ResolveModelName(modelName, scale)
	cmd := exec.Command(realesrganPath,
		"-i", inFramePath,
		"-o", outFramePath,
		"-n", resolved,
		"-s", fmt.Sprintf("%d", scale),
		"-m", modelsDir,
		"-t", "0",
		"-j", "2:2:2",
	)
	SetHideWindow(cmd)
	cmd.Dir = GetEngineRootDir()
	return cmd.Run()
}
