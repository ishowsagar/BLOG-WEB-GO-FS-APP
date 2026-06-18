This Go application provides an AI-powered agent leveraging the Gemini API for code analysis.

**Files Reviewed:**

*   **`agent_handler.go`**: Manages prompt construction (`BuildPrompt`, `BuildDirPrompt`) for different modes (review, docs, qa) and handles conversation history persistence (`CheckConvoHistoryFILE`, `HandleHistoryWrites`, `HandleHistoryReads`) for "deep learning" mode.
*   **`ai_agent.go`**: Offers core file I/O utilities (`ReadFileContent`, `WriteToFile`) essential for reading source code and storing AI responses.
*   **`main.go`**: The orchestrator, parsing flags (`--mode`, `--deeplearning`, `--type`) to define agent behavior. It supports analyzing single files or entire directories, sends formatted content to Gemini, and manages AI request/response payloads and user feedback.

This system efficiently performs code review, documentation, or Q&A with optional historical context.