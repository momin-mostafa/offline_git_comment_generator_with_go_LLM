# CommitSmith

Local LLM-powered Git commit message generator. Forges meaningful commit messages from your staged changes using Ollama.

## Prerequisites

- Go 1.21+
- [Ollama](https://ollama.ai) installed and running
- A supported model: `qwen2.5-coder`, `deepseek-coder`, `llama3`, or `mistral`

## Installation

```bash
git clone https://github.com/momin-mostafa/CommitSmith.git
cd CommitSmith
./build.sh
```

This creates a `cmtr` executable in the current directory.

## Usage

1. Stage your changes:
```bash
git add .
```

2. Run CommitSmith:
```bash
./cmtr
```

3. Pick a suggested message (1-3), enter `0` to write your own, or `q` to cancel.

### Options

| Flag | Description |
|------|-------------|
| `-d` | Show staged diff before generating messages |

```bash
./cmtr -d
```

### Example Output

```
Analyzing staged changes...

Suggested commit messages:
1. feat(auth): add JWT token refresh endpoint
2. fix(auth): handle expired token edge case
3. chore(auth): update token validation middleware
0. Write your own commit message

Choose (1-3, 0 to write your own) or q: 1

Running

git commit -m "feat(auth): add JWT token refresh endpoint"

✓ Commit successful

Push now? (y/N):
```

When you choose `0` (write your own):

```
Choose (1-3, 0 to write your own) or q: 0

Enter your commit message: update the token stuff
Formatting commit message...

Formatted: fix(auth): update token refresh logic
Original:  update the token stuff

Use formatted? (Y/n): Y

Running

git commit -m "fix(auth): update token refresh logic"

✓ Commit successful
```

Pressing Enter without input selects the first suggestion.

## Configuration

On first run, `~/.commitsmith.yaml` is created automatically with defaults. Customize as needed:

```yaml
model: qwen2.5-coder
host: http://localhost:11434
temperature: 0.2
max_options: 3
use_conventional_commits: true
```

| Setting | Default | Description |
|---------|---------|-------------|
| `model` | `qwen2.5-coder` | Ollama model to use. Falls back to first available model if not found. |
| `host` | `http://localhost:11434` | Ollama API endpoint |
| `temperature` | `0.2` | Generation temperature (0-2). Lower = more deterministic. |
| `max_options` | `3` | Number of commit message suggestions to generate |
| `use_conventional_commits` | `true` | Follow [Conventional Commits](https://www.conventionalcommits.org/) format |

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
go install ./cmd/cmtr
```

## Project Structure

```
git_comment_maker/
├── build.sh
├── cmd/
│   └── cmtr/
│       └── main.go
├── internal/
│   ├── config/        # YAML config loading and defaults
│   ├── git/           # Git operations (diff, commit, push)
│   └── ollama/        # Ollama API client and prompt handling
├── pkg/
│   └── models/        # Shared data types
└── README.md
```

## License

MIT
