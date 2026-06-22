This Go-based AI Agent (`main.go`) streamlines code analysis, documentation, and Git automation. It dynamically generates AI prompts (`agent_handler.go`) for specific modes (review, docs, QA), supporting both single-file analysis (`ai_agent.go`) and comprehensive multi-file directory scans, including nested structures. The agent integrates with Git (`agent_executable.go`, `agent_handler.go`) to generate commit messages based on staged changes. Advanced features include conversation history for deep learning and a real-time file watcher (`agent_handler.go`) that triggers AI analysis upon code modifications. User input for configuration is handled by `streams.go`. The nested `agent.txt` serves as a demo file for content processing.

**Reviewed Files:**
*   `agent_executable.go`
*   `agent_handler.go`
*   `ai_agent.go`
*   `agent.txt`
*   `main.go`
*   `streams.go`