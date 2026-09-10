# 📸 StockTaggerIA

> **AI-Powered Stock Photography SEO Metadata & Tagging Tool**  
> Optimized for **Unsplash**, **Pexels**, **Pixabay**, and **Adobe Lightroom Classic**.

[![CI Pipeline](https://github.com/lagosproject/StockTaggerIA/actions/workflows/ci.yml/badge.svg)](https://github.com/lagosproject/StockTaggerIA/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Platforms](https://img.shields.io/badge/Platforms-Linux%20%7C%20macOS%20%7C%20Windows%20x64%20%7C%20Windows%20ARM64-blue?style=for-the-badge)]()
[![i18n](https://img.shields.io/badge/UI%20Languages-EN%20%7C%20ES%20%7C%20FR-orange?style=for-the-badge)]()

---

## 🌟 Overview & Architecture

**StockTaggerIA** is a production-ready, cross-platform CLI agent and Lightroom Classic integration designed to automate stock photography keyword tagging and SEO metadata generation using Google Gemini Vision AI.

```
                    ┌────────────────────────┐
                    │   Input Photo Folder   │
                    │ (.jpg, .png, .webp...) │
                    └───────────┬────────────┘
                                │
                    ┌───────────▼────────────┐
                    │   StockTaggerIA CLI    │
                    │ (Interactive / Headless│
                    └─────┬────────────┬─────┘
                          │            │
         [Direct Vision AI]            [Offline Metadata]
                          │            │
       ┌──────────────────▼──────┐  ┌──▼─────────────────┐
       │   Google Gemini API     │  │ JSON Metadata File │
       │ (Auto-negotiated Model) │  │  (new_tags.json)   │
       │ (Rate-throttled 15 RPM) │  └──┬─────────────────┘
       └──────────────┬──────────┘     │
                      │                │
                      └───────┬────────┘
                              │
                    ┌─────────▼───────────┐
                    │ Pre-flight Backup   │
                    │ stock_tags_*.json   │
                    └─────────┬───────────┘
                              │
                    ┌─────────▼───────────┐
                    │ ExifTool Daemon     │
                    │ (-stay_open mode)   │
                    └─────────┬───────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
     Standard Keyword Fields:        Description Fields:
     - XMP-dc:Subject                - XMP-dc:Description
     - IPTC:Keywords                 - IPTC:Caption-Abstract
     - EXIF:XPKeywords (Windows)     - EXIF:ImageDescription
```

### Key Highlights
* **🤖 Dual AI & Offline Tagging Engines:**
  * **Direct Gemini Vision AI:** Multimodal visual analysis with intelligent model discovery (`gemini-3.6-flash`, `gemini-2.5-flash`, `gemini-2.0-flash`, `gemini-1.5-flash`) and automatic fallback negotiation upon API quota or deprecation.
  * **Offline JSON Mode:** Batch import pre-calculated keywords without cloud latency.
* **🌍 Localized Interactive Wizard (i18n):**
  * Step-by-step terminal wizard in **English**, **Spanish (Español)**, and **French (Français)** with OS locale auto-detection.
* **🎯 Unified Stock Platform Presets:**
  * **Unsplash & Pexels:** Generates up to 30 high-ranking SEO tags with literal & conceptual balance.
  * **Pixabay:** Generates up to 25 keywords matching Pixabay search bounds.
  * **Multilingual Metadata:** Generates metadata in English, Spanish, French, or trilingual combinations.
* **⚡ High-Throughput ExifTool Daemon:**
  * Leverages ExifTool's persistent `-stay_open` daemon mode to batch-tag thousands of images in seconds.
* **🧩 Adobe Lightroom Classic Lua Plugin:**
  * Bundled [`lrplugin/StockTaggerIA.lrplugin`](lrplugin/StockTaggerIA.lrplugin) allows seamless post-export metadata injection.
* **🔒 Privacy & Safety First:**
  * Masked API key entry in terminal wizard.
  * In-memory proportional downscaling (max 1500px) before upload to protect bandwidth and avoid API payload rejection.
  * Automatic local backup (`stock_tags_<timestamp>.json`) saved prior to file modification.

---

## 📋 Prerequisites

1. **Go 1.22+** (only required if building from source).
2. **ExifTool:**
   * **Linux:** `sudo apt-get install libimage-exiftool-perl`
   * **macOS:** `brew install exiftool`
   * **Windows:** Download portable `exiftool.exe` and place it in the application folder or add to `PATH`.
3. **Google Gemini API Key** (optional, only for direct AI tagging):
   * Obtain a key at [Google AI Studio](https://aistudio.google.com/app/apikey).

---

## ⚙️ Environment Variables Setup

Copy the sample environment file:

```bash
cp .env.example .env
```

Available environment variables:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `GEMINI_API_KEY` | Google Gemini API key for vision tagging | *None* |
| `GEMINI_MODEL` | Override default AI vision model | `gemini-3.6-flash` |
| `EXIFTOOL_PATH` | Explicit path to ExifTool executable | *Auto-detected* |
| `DEFAULT_PHOTOS_DIR` | Fallback directory containing photos | `.` |
| `DEFAULT_PLATFORM` | Default preset: `stock` or `pixabay` | `stock` |
| `DEFAULT_LANG` | Default metadata language: `en`, `es`, `fr` | `en` |

---

## 🚀 Quick Start

### 1. Interactive Terminal Wizard (Guided)
Simply launch without arguments:

```bash
stocktaggeria
# or explicitly:
stocktaggeria -i
```

The wizard will guide you through:
1. Interface language selection (English / Español / Français).
2. Mode selection (Gemini Vision AI or Local JSON).
3. Photo folder selection (supports drag-and-drop file paths).
4. Stock preset (`stock` for Unsplash/Pexels or `pixabay`).
5. Target metadata language.
6. Execution mode: **Dry Run** (preview only) or **Live Update**.

---

### 2. Headless CLI Mode (Automated / Scriptable)

#### Tagging with Direct Gemini Vision AI:
```bash
# Export API key
export GEMINI_API_KEY="your_api_key_here"

# Tag photos for Unsplash & Pexels in English:
stocktaggeria --gemini -p stock -l en -d "/path/to/photos"

# Tag photos for Pixabay in Spanish:
stocktaggeria --gemini -p pixabay -l es -d "/path/to/photos"

# Perform a safe dry-run (simulation):
stocktaggeria --gemini -p stock -l fr -d "/path/to/photos" --dry-run
```

#### Offline Tagging via JSON:
```bash
stocktaggeria -p stock -l en -j data/new_tags.json -d "/path/to/photos"
```

---

## 📖 CLI Flags Reference

| Flag | Long Flag | Description | Default |
| :--- | :--- | :--- | :--- |
| `-p` | `--platform` | Stock preset: `stock` (Unsplash & Pexels, max 30) or `pixabay` (max 25) | `stock` |
| `-l` | `--lang` | Target metadata language: `en` (English), `es` (Spanish), `fr` (French) | `en` |
| `-d` | `--dir` | Directory containing photos to tag (recursive) | `.` |
| `-j` | `--json` | Path to JSON metadata file for offline mode | `new_tags.json` |
| | `--gemini` | Enable direct Gemini Vision AI tagging | `false` |
| | `--api-key` | Gemini API key (fallback: `GEMINI_API_KEY` env var) | `""` |
| | `--dry-run` | Simulate operations without modifying image files | `false` |
| `-i` | `--interactive`| Launch interactive terminal wizard | `false` |
| `-v` | `--version` | Display version information | |

---

## 🛠️ Development, Tests & Cross-Compilation

### Run Tests & Verification
```bash
# Run tests with race condition detector and coverage
go test -v -race -cover ./...

# Run code analysis
go vet ./...
```

### Local Build for Current Platform
```bash
go build -trimpath -ldflags="-s -w" -o bin/stocktaggeria ./src/main.go
```

### Cross-Compilation (Multiplatform)

Thanks to Go's native cross-compiler, you can build binaries for any architecture without CGO:

```bash
# Linux x86_64
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/stocktaggeria-linux-amd64 ./src/main.go

# macOS Apple Silicon (ARM64)
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o bin/stocktaggeria-darwin-arm64 ./src/main.go

# Windows x86_64
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/stocktaggeria-windows-amd64.exe ./src/main.go

# Windows on ARM64 (Qualcomm Snapdragon X Elite / Surface Pro ARM)
GOOS=windows GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o bin/stocktaggeria-windows-arm64.exe ./src/main.go
```

---

## 🧩 Adobe Lightroom Classic Plugin

StockTaggerIA includes an optional export plugin for Adobe Lightroom Classic located at [`lrplugin/StockTaggerIA.lrplugin`](lrplugin/StockTaggerIA.lrplugin).

### Installation:
1. Open Adobe Lightroom Classic.
2. Navigate to **File > Plug-in Manager**.
3. Click **Add** and select the `lrplugin/StockTaggerIA.lrplugin` folder.
4. When exporting photos (**File > Export**), expand the **StockTaggerIA** post-processing panel:
   * Select platform preset (**Unsplash & Pexels** or **Pixabay**).
   * Select metadata language (**English**, **Spanish**, or **French**).
   * Choose Dry Run or Live mode.
5. Click **Export** — Lightroom will export the photos and StockTaggerIA will automatically tag them.

---

## 🤝 Companion Projects

Looking for a **100% offline desktop GUI** running local ONNX models without uploading any data to the cloud?  
Check out [**ImageLabelIA**](https://github.com/lagosproject/ImageLabelIA) — powered by Tauri v2 (Rust) and Angular for offline photo tagging using ConvNeXt and DETR models.

---

## 📜 License

Distributed under the [MIT License](LICENSE).
