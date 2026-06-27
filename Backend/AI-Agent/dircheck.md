Ye dekho, tumhare code ka human-friendly Hinglish documentation, 470 words ki limit mein, saari files ko cover karte hue:

---

### **Code Documentation (Hinglish)**

Ye pura codebase ek smart AI agent banata hai jo code review, documentation generation, QA, aur Git automations jaise kaamo mein madad karta hai. Ye `Go` language mein likha gaya hai aur `Google Gemini` API ka use karta hai.

**Files Reviewed:**

#### **1. `agent_executable.go`**
Ye file `RunGitOperator` naam ka ek function deta hai. Iska main kaam hai `Git` commands ko execute karna aur unka output `io.Reader` format mein provide karna. Matlab, agar app ko `git status` ya `git diff` jaise commands chalane hain, toh ye function unhe run karta hai aur unka result app ke dusre parts tak pahunchata hai.

#### **2. `agent_handler.go`**
Ye file agent ke core logic aur utility functions ka hub hai. Yahan `BuildPrompt`, `BuildDirPrompt`, aur `BuildGitPrompt` jaise functions hain jo AI ke liye specific context-based prompts banate hain (jaise code review ke liye, ya docs generate karne ke liye). `CheckConvoHistoryFILE`, `HandleHistoryWrites`, aur `HandleHistoryReads` functions user aur AI ke beech ki conversations ko `history.json` file mein manage karte hain, taki agent previous talks ko yaad rakh sake (deep learning mode). `WatchDirChanges` ek background mein chalta hai jo directory mein file changes (jaise `.go` files mein edit) ko monitor karta hai aur automatic AI trigger karta hai. Ismein general-purpose `Writer` functions bhi hain jo data ko alag-alag destinations par likhte hain.

#### **3. `ai_agent.go`**
Is file mein basic file input/output operations hain. `ReadFileContent` function kisi bhi file se content read karta hai. Aur `WriteToFile` function AI se mile hue responses ko ek specified file mein save karta hai. Ye functions agent ko data input aur output efficiently handle karne mein help karte hain.

#### **4. `ans.go`**
Ye file application ke functional code ka part nahi hai. Ye developer ke personal notes hain, jismein previous questions ke answers aur architectural thoughts hain.

#### **5. `agent.txt` (Nested dir -demoDir)**
Ye `demoDir` ke andar ek sample text file hai, jise `main.go` mein context ke roop mein read kiya jata hai. Ye typically testing aur demo purposes ke liye use hota hai, taaki AI agent ko test kiya ja sake.

#### **6. `main.go`**
Ye puri application ka entry point hai. Yahan `godotenv` ka use karke API keys load ki jaati hain. `AgentConfig` struct user ke saare selections store karti hai. Google Gemini API ke saath communicate karne ke liye `OutboundPayloadGem` aur `InboundPayloadGem` jaise structs define kiye gaye hain. `IntializeTUI` function (Terminal User Interface) ke through user se interactive tarike se inputs liye jaate hain – jaise kaun sa agent mode choose karna hai (file/directory scan, deep learning, Git automations), kis file mein output chahiye, aur koi special instructions. User ke selections ke basis par, ye deep learning, directory scanning, ya Git automation modes ko execute karta hai. Ye Gemini API ko HTTP requests bhejta hai aur streamed responses ko console aur output file dono par display karta hai. Agar user file watching mode select karta hai, toh `WatchDirChanges` goroutine trigger ho jaati hai. `select {}` statement main program ko chalte rehne deta hai taaki background services active rahein.

#### **7. `streams.go`**
Ye file terminal input/output (I/O) ko handle karta hai. `CaptureInputFromShell` function `os.Stdin` (standard input) ka use karke user se input leta hai aur `os.Stdout` (standard output) par prompts dikhata hai. Ye `io.Reader` aur `io.Writer` interfaces ke fundamental concepts ko explain karta hai, ki Go mein sab kuch bytes ke movement par based hai.

#### **8. `tui.go`**
Ye file `charmbracelet/bubbletea` library ka use karke Terminal User Interface (TUI) implement karta hai. `TUIModel` struct TUI application ki current state (choices, cursor position, selected item) ko maintain karti hai. `IntializeTUI` function `bubbletea` program ko start karta hai aur user ki final selection return karta hai. `Init()`, `View()`, aur `Update()` methods `tea.Model` interface ko satisfy karte hain, jo TUI ko render karte hain aur user ke keypress events (up, down, enter, quit) ko handle karke UI ko update karte hain.