This Go application serves as an AI-powered agent to enhance developer workflows. It leverages the Gemini API for intelligent code review, documentation generation, and Q&A analysis. The agent can process individual files (`ai_agent.go`) or scan entire directories, including nested structures (`agent_handler.go`, `main.go`).

Key features include:
*   Automated Git operations for smart commit message generation from staged changes and diff analysis (`agent_executable.go`).
*   Conversational memory for context-aware interactions (`agent_handler.go`).
*   A real-time file watcher that automatically triggers AI analysis on code modifications (`agent_handler.go`).
*   Interactive user input via terminal streams (`streams.go`).

The `main.go` file orchestrates these functionalities, handling user selections and API interactions.