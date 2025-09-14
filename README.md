# Speech-to-Text CLI

A terminal-based application for transcribing audio from video and audio files using OpenAI's Whisper model locally.

## Features

- **Multi-format Support**: Works with video files (MP4, AVI, MOV, MKV, WebM) and audio files (MP3, WAV, M4A, FLAC)
- **Accurate Transcription**: Uses OpenAI's Whisper model for high-quality speech recognition
- **Beautiful TUI**: Interactive terminal interface built with Bubble Tea
- **Easy Navigation**: Browse directories with support for going back to parent folders
- **Scrollable Results**: View long transcriptions with smooth scrolling

## Prerequisites

Before running this application, make sure you have the following installed:

### Required Software
- **Go** (1.19 or later)
- **Python** (3.7 or later)
- **FFmpeg** - For audio extraction from video files

### FFmpeg Installation

**Windows:**
- Download from [https://ffmpeg.org/download.html](https://ffmpeg.org/download.html)
- Extract and add to your PATH, or place `ffmpeg.exe` in the project directory

**macOS:**
```bash
brew install ffmpeg
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install ffmpeg
```

## Installation

1. **Clone the repository:**
```bash
git clone https://github.com/andyanalog/speech-to-text-cli.git
cd speech-to-text-cli
```

2. **Install Go dependencies:**
```bash
go mod tidy
```

3. **Build the application (Windows):**
```bash
go build -o stt-cli.exe .
```

## Usage

1. **Run the application:**
```bash
./stt-cli
```

2. **Navigate and select files:**
   - Use arrow keys to navigate through files and folders
   - Press **Enter** to select a file or enter a directory
   - Press **Backspace**, **Left Arrow**, or **H** to go back to parent directory

3. **View transcription results:**
   - After processing, use **Up/Down** arrows or **J/K** keys to scroll through long transcriptions
   - Press **Home** to go to the beginning, **End** to go to the end
   - Press **Enter** to process another file
   - Press **Q** or **Ctrl+C** to exit

## How It Works

The application follows this pipeline:

1. **Audio Extraction**: Uses FFmpeg to extract audio from video files or process audio files directly
2. **Transcription**: Employs OpenAI's Whisper model to convert speech to text with timestamps
3. **Display**: Shows the transcription in a scrollable terminal interface

## Python Dependencies

The application automatically installs the required Python packages on first run:
- `openai-whisper` - For speech recognition
- `pydub` - For audio processing
- Additional dependencies as needed

## Supported File Formats

**Video Files:**
- MP4, AVI, MOV, MKV, WebM

**Audio Files:**
- MP3, WAV, M4A, FLAC

## Keyboard Controls

### File Selection Mode
- **↑/↓** - Navigate files and folders
- **Enter** - Select file or enter directory
- **Backspace/←/H** - Go back to parent directory
- **Q/Ctrl+C** - Quit application

### Transcription View Mode
- **↑/↓ or J/K** - Scroll through transcription
- **Home** - Go to beginning
- **End** - Go to end
- **Enter** - Process another file
- **Q/Ctrl+C** - Quit application

## Technical Details

- Built with Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Uses [Lipgloss](https://github.com/charmbracelet/lipgloss) for terminal styling
- Integrates OpenAI's Whisper for state-of-the-art speech recognition

## Troubleshooting

**FFmpeg not found:**
- Ensure FFmpeg is installed and in your PATH
- On Windows, you can place `ffmpeg.exe` in the same directory as the executable

**Python packages installation fails:**
- Ensure Python and pip are installed and accessible
- Try running `pip install openai-whisper` manually

**Audio extraction fails:**
- Check that your video/audio file is not corrupted
- Ensure the file format is supported

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.