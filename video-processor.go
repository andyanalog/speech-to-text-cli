package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AudioProcessor handles the speech-to-text pipeline
type AudioProcessor struct {
	InputPath  string
	TempDir    string
	FFmpegPath string
	PythonPath string
}

// processAudioSTT orchestrates the speech-to-text process
func processAudioSTT(inputPath string) (string, error) {
	processor := &AudioProcessor{
		InputPath: inputPath,
		TempDir:   filepath.Join(os.TempDir(), "audio_stt"),
	}

	// Create temp directory
	if err := os.MkdirAll(processor.TempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(processor.TempDir)

	// Check dependencies
	if err := processor.checkDependencies(); err != nil {
		return "", fmt.Errorf("dependency check failed: %w", err)
	}

	// Extract audio from video/audio file
	audioPath := filepath.Join(processor.TempDir, "audio.wav")
	if err := processor.extractAudio(audioPath); err != nil {
		return "", fmt.Errorf("audio extraction failed: %w", err)
	}

	// Transcribe audio
	transcription, err := processor.transcribeAudio(audioPath)
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	return transcription, nil
}

// checkDependencies verifies required tools are available
func (p *AudioProcessor) checkDependencies() error {
	dependencies := map[string]*string{
		"ffmpeg": &p.FFmpegPath,
		"python": &p.PythonPath,
	}

	for tool, pathVar := range dependencies {
		path, err := exec.LookPath(tool)
		if err != nil {
			// Try common Windows locations for ffmpeg
			if tool == "ffmpeg" {
				commonPaths := []string{
					"C:\\ffmpeg\\bin\\ffmpeg.exe",
					"C:\\Program Files\\ffmpeg\\bin\\ffmpeg.exe",
					".\\ffmpeg.exe",
				}
				for _, commonPath := range commonPaths {
					if _, err := os.Stat(commonPath); err == nil {
						*pathVar = commonPath
						break
					}
				}
				if *pathVar == "" {
					return fmt.Errorf("ffmpeg not found. Please install FFmpeg or place ffmpeg.exe in the current directory")
				}
			} else {
				return fmt.Errorf("%s not found in PATH", tool)
			}
		} else {
			*pathVar = path
		}
	}

	// Install required Python packages
	cmd := exec.Command(p.PythonPath, "-m", "pip", "install", "openai-whisper")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install openai-whisper: %w", err)
	}

	return nil
}

// extractAudio extracts audio track from video/audio file using FFmpeg
func (p *AudioProcessor) extractAudio(outputPath string) error {
	cmd := exec.Command(p.FFmpegPath,
		"-i", p.InputPath,
		"-vn", // no video
		"-acodec", "pcm_s16le",
		"-ar", "16000", // 16kHz sample rate for Whisper
		"-ac", "1", // mono
		"-f", "wav",
		outputPath,
		"-y", // overwrite output file
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %s", string(output))
	}
	return nil
}

// Helper function to escape paths for Python (using raw strings)
func pythonPath(path string) string {
	// Use raw string representation for Python
	return fmt.Sprintf(`r"%s"`, path)
}

// transcribeAudio uses Whisper to transcribe audio and return the text
func (p *AudioProcessor) transcribeAudio(audioPath string) (string, error) {
	script := fmt.Sprintf(`
import whisper
import os

print("Loading Whisper model...")
model = whisper.load_model("base")
print("Transcribing audio...")
result = model.transcribe(%s)

# Extract the full transcription text
transcription = result["text"].strip()
print("Transcription completed")
print("=" * 50)
print(transcription)
print("=" * 50)

# Save to a temporary file for Go to read
temp_output = %s
with open(temp_output, "w", encoding="utf-8") as f:
    f.write(transcription)
`, pythonPath(audioPath), pythonPath(filepath.Join(p.TempDir, "transcription.txt")))

	scriptPath := filepath.Join(p.TempDir, "transcribe.py")
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return "", err
	}

	cmd := exec.Command(p.PythonPath, scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("python transcription error: %s", string(output))
	}

	// Read the transcription from the temporary file
	transcriptionPath := filepath.Join(p.TempDir, "transcription.txt")
	transcriptionBytes, err := os.ReadFile(transcriptionPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcription file: %w", err)
	}

	transcription := strings.TrimSpace(string(transcriptionBytes))
	if transcription == "" {
		return "No speech detected in the audio file.", nil
	}

	return transcription, nil
}
