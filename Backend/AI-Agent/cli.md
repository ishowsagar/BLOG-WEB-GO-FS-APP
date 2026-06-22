This Go codebase powers an AI agent for comprehensive code analysis.
**`main.go`** serves as the interactive CLI, allowing users to select modes (review, docs, QA), enable deep learning (conversation history via **`agent_handler.go`**), and specify file or directory targets leveraging **`ai_agent.go`** for I/O.
It integrates **`agent_executable.go`** to run `git` commands, automating commit message generation or diff reviews.
**`agent_handler.go`** also dynamically builds AI prompts and features a `fsnotify`-based `WatchDirChanges` for real-time file monitoring.
**`streams.go`** facilitates user input. The system communicates with the Gemini API, showcasing robust integration of file system, Git, and conversational AI functionalities.