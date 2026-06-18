# 🤖 Go AI Agent — Automated Code Review & Documentation

A standalone CLI AI agent built in pure Go that reads your code files and uses the **Gemini API** to automatically review, document, or quiz you on your code. Save a file and the agent reviews it instantly.

---

## ✨ Features

- **Code Review** — get expert feedback on your code
- **Documentation** — auto-generate docs for any file
- **Q&A Mode** — get questions to test your understanding
- **File Watcher** — auto-triggers on every file save
- **Directory Mode** — review your entire codebase at once
- **Conversation Memory** — Gemini remembers previous context

---

## 📋 Prerequisites

### 1. Get a Gemini API Key (Free)
1. Go to [aistudio.google.com/api-keys](https://aistudio.google.com/api-keys)
2. Sign in with a **personal Google account** (not a Workspace/org account)
3. Click **Create API key** → **Create API key in new project**
4. Copy the key — it should start with `AIzaSy...`

### 2. Create a `.env` file
Create a `.env` file in the **parent folder** of `AI-Agent/`:

```
GEM_KEY=AIzaSy_your_key_here
```

> ⚠️ Never share or commit this file. Add `.env` to your `.gitignore`.

### 3. Install Go (if running from source)
Download from [go.dev/dl](https://go.dev/dl) — version 1.24+

---

## 🚀 Quick Start

### Option A — Run from source
```bash
git clone https://github.com/ishowsagar/BLOG-WEB-GO-APP
cd BLOG-WEB-GO-APP/Backend/AI-Agent
go run . --mode review "main.go" "out.md" "." ".go"
```

### Option B — Run the binary (Windows)
Download `agent.exe`, open PowerShell in the same folder and run:
```powershell
.\agent.exe --mode review "main.go" "out.md" "." ".go"
```

---

## 🎮 Usage

### Basic Command Structure
```bash
go run . [flags] [input_file] [output_file] [watch_dir] [file_suffix]
```

| Argument | Description | Example |
|----------|-------------|---------|
| `input_file` | File to analyze | `main.go` |
| `output_file` | Where to save AI response | `out.md` |
| `watch_dir` | Directory to watch (use `.` for current) | `.` |
| `file_suffix` | File type to watch | `.go` |

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--mode` | `review` | Agent mode: `review`, `docs`, `qa` |
| `--type` | `file` | Input type: `file` or `dir` |
| `--deeplearning` | `false` | Enable conversation memory |

---

## 📖 Examples

### Review a single file
```bash
go run . --mode review "main.go" "out.md" "." ".go"
```

### Generate documentation
```bash
go run . --mode docs "main.go" "out.md" "." ".go"
```

### Q&A on your code
```bash
go run . --mode qa "main.go" "out.md" "." ".go"
```

### Review entire directory
```bash
go run . --type dir --mode review "none" "out.md" "." ".go"
```

### Enable conversation memory (Gemini remembers context)
```bash
go run . --mode review --deeplearning=true "main.go" "out.md" "." ".go"
```

### Auto-watch mode (reviews on every file save)
```bash
go run . --mode review --type file "main.go" "out.md" "." ".go"
```
> After startup, save any `.go` file and the agent auto-triggers! 🔥

---

## 📁 Project Structure

```
AI-Agent/
├── main.go           # Entry point — orchestrates all modes
├── ai_agent.go       # File I/O utilities (ReadFileContent, WriteToFile)
├── agent_handler.go  # Prompt engineering (BuildPrompt, BuildDirPrompt)
├── history.json      # Conversation memory (auto-created)
└── out.md            # AI response output (auto-created)
```

---

## 🧠 How It Works

```
1. Read file/directory content
        ↓
2. Build mode-specific prompt
        ↓
3. Send to Gemini API
        ↓
4. Receive AI response
        ↓
5. Save to output .md file
```

For **watch mode**:
```
Start watcher → detect file save → auto-trigger above flow
```

For **conversation memory**:
```
Load history.json → append new message → send full history → save response
```

---

## ⚡ Modes Explained

### `--mode review`
Gemini acts as an expert code reviewer — gives feedback on code quality, patterns, and improvements.

### `--mode docs`
Gemini acts as a technical writer — generates clean documentation for your code.

### `--mode qa`
Gemini acts as a coding mentor — asks you up to 5 questions about your code to test and deepen your understanding.

---

## 🔧 Build Binary

```bash
# Windows
go build -o agent.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o agent .

# Mac
GOOS=darwin GOARCH=amd64 go build -o agent .
```

---

## ⚠️ Known Limitations

- Gemini free tier allows **15 requests/minute** — saving files rapidly may hit rate limits (429 error)
- `--deeplearning` mode grows `history.json` over time — delete it to reset memory
- Directory mode sends all `.go` files combined — very large codebases may exceed Gemini's context window
- Watch mode currently works best on the current directory (`.`)

---

## 🛠️ Tech Stack

- **Language** — Go 1.24
- **AI** — Gemini 2.5 Flash (REST API, no SDK)
- **Auth** — `x-goog-api-key` header
- **Dependencies** — `godotenv`, `fsnotify`, `briandowns/spinner`
- **Storage** — local `.json` and `.md` files
- **Architecture** — pure Go stdlib, no frameworks

---

## 👨‍💻 Author

Built by [Jr.Sagar](https://github.com/ishowsagar) — part of the [BLOG-WEB-GO-APP](https://github.com/ishowsagar/BLOG-WEB-GO-APP) project.
