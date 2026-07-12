# Git Comment

Local LLM-powered Git commit message generator. Analyzes your staged changes and suggests meaningful commit messages using Ollama.

## Prerequisites

- Go 1.21+
- [Ollama](https://ollama.ai) installed and running
- A supported model: `qwen2.5-coder`, `deepseek-coder`, `llama3`, or `mistral`

## Installation

```bash
git clone <repo-url>
cd git_comment_maker
./build.sh
```

This creates a `git_comment` executable in the current directory.

## Usage

1. Stage your changes:
```bash
git add .
```

2. Run git_comment:
```bash
./git_comment
```

3. Select a commit message (1-3) or press `q` to cancel.

**Options:**
- `-d` — Show staged diff before generating messages

```bash
./git_comment -d
```

## Configuration

Create `~/.git_comment.yaml` to customize settings:

```yaml
model: qwen2.5-coder
host: http://localhost:11434
temperature: 0.2
max_options: 3
use_conventional_commits: true
```

**Defaults:**
- `model`: qwen2.5-coder
- `host`: http://localhost:11434
- `temperature`: 0.2
- `max_options`: 3
- `use_conventional_commits`: true

## Error Messages

| Error | Meaning |
|-------|---------|
| Current directory is not a Git repository | Run from inside a git repo |
| No staged changes found | Run `git add .` first |
| Unable to connect to Ollama | Start Ollama: `ollama serve` |
| Configured model not found | Pull the model: `ollama pull qwen2.5-coder` |

## Development

```bash
# Run tests
go test ./...

# Build
./build.sh

# Install (optional)
go install ./cmd/git_comment
```

## Project Structure

```
git_comment_maker/
├── build.sh
├── cmd/
│   └── git_comment/
│       └── main.go
├── internal/
│   ├── config/
│   ├── git/
│   └── ollama/
├── pkg/
│   └── models/
└── tasks.md
```

## License

MIT
# git_comment_generator_with_go_LLM
