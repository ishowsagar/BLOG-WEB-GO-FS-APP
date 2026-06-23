As an expert code documentation generator, I've thoroughly analyzed the provided Go codebase, which functions as an AI-powered agent for developer assistance. The system is designed to perform tasks like code review, documentation generation, QA, and Git automation, all while providing an interactive and responsive user experience.

Here's a comprehensive human-readable documentation of the entire project, analyzing the files and their interconnections:

---

## AI Agent Codebase Documentation

This project implements an AI agent in Go, leveraging the Gemini API to provide intelligent assistance for various development workflows. It features interactive command-line input, real-time file system monitoring, and seamless integration with Git operations.

### Reviewed Files:

#### `agent_executable.go`

This file is responsible for executing arbitrary Git commands and capturing their output.

*   **`RunGitOperator(args ...string) (io.Reader, error)`**:
    This function serves as the core Git command executor. It takes a variadic list of strings, which represent the Git command and its arguments (e.g., `"git", "status"`).
    1.  It constructs an `exec.Command` for the "git" executable.
    2.  `cmd.Stderr` is redirected to `os.Stderr`, ensuring that any errors from the Git command are immediately visible in the console.
    3.  `cmd.StdoutPipe()` is used to obtain an `io.Reader` from which the standard output of the Git command can be read. This allows the application to capture the command's results programmatically.
    4.  `cmd.Start()` initiates the Git process.
    5.  The entire output from the `outputPipe` is then read into a byte slice using `io.ReadAll()`.
    6.  `cmd.Wait()` ensures that the Git command completes its execution before the function proceeds, capturing any final errors.
    7.  Finally, the captured output bytes are wrapped in a `bytes.Buffer`, which conveniently implements the `io.Reader` interface. This makes the Git command's output highly flexible, allowing any `io.Reader` consumer (like other functions in the agent) to process it in chunks or all at once.

    **Relationship:** This function is a critical utility for `main.go` and `agent_handler.go` when performing Git automation tasks, providing a standardized way to interact with the local Git repository.

#### `agent_handler.go`

This file houses much of the agent's core logic, including prompt generation, conversation history management, file system watching, and specific AI request handlers for Git. It also provides a suite of generic `io.Writer` utility functions.

*   **`BuildPrompt(mode, content string) string`**: Generates AI prompts for single-file analysis (e.g., "review", "docs", "qa"). It validates the `mode` and appends the provided `content` to the prompt.
*   **`BuildDirPrompt(mode, content string) string`**: Similar to `BuildPrompt`, but specifically tailored for directory analysis. It adds extra context to the prompt, instructing the AI to expect multi-file data, analyze relationships, and label reviewed files.
*   **`BuildGitPrompt(gitMode, content string) string`**: Creates specialized prompts for Git operations (`status`, `add`, `commit`, `diff`). These prompts are designed to elicit specific AI responses, such as a concise commit message or actionable review comments.
*   **`CheckConvoHistoryFILE(filename string) (historyExists bool, State error)`**: Checks for the existence of a conversation history file (`history.json`), which is crucial for the "deep learning" (memory) feature.
*   **`HandleHistoryWrites(writeToFile string, history []byte) error`**: Writes the marshaled conversation history (a slice of `ContentsSliceKeyWrapperGem` structs) to the specified file.
*   **`HandleHistoryReads(history []byte) ([]*ContentsSliceKeyWrapperGem, error)`**: Unmarshals the byte data from the history file into a slice of `ContentsSliceKeyWrapperGem` structs, reconstructing the conversation.
*   **`WatchDirChanges(dirToWatchOut, fileSuffix, Mode, APIKEY, OutputFileName string) (watcherErr error)`**:
    This is a powerful background goroutine that uses `fsnotify` to monitor a specified directory for file write events.
    1.  It initializes an `fsnotify.Watcher` and adds the target directory for observation.
    2.  An infinite `go func` loop continuously listens for events on `dirWatcher.Events`.
    3.  When a `fsnotify.Write` event is detected on a file matching the `fileSuffix` (e.g., `.go`), it triggers the AI analysis.
    4.  It reads the modified file's content using `ReadFileContent` (from `ai_agent.go`).
    5.  An AI request payload is constructed using `BuildPrompt` with the content.
    6.  A `net/http` client sends this request to the Gemini API. A `spinner` from `github.com/briandowns/spinner` provides user feedback during the AI processing.
    7.  The AI's response is retrieved, unmarshaled, and then written to the `OutputFileName` using `WriteToFile` (from `ai_agent.go`), along with a success message to the console.
    8.  Error events from the watcher are logged, and the loop can safely `break` on unrecoverable errors.

    **Relationship:** `WatchDirChanges` is initiated by `main.go` and provides a continuous, automated review/documentation loop for actively developed files. It leverages `ai_agent.go