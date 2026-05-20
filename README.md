# GoFromScratchDB

![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Status](https://img.shields.io/badge/Status-Active_Development-orange.svg)

A database built from scratch in Go, designed as a learning project for tech freshers and anyone curious about database internals.

---

## 🚀 Project Goal
- **Understand how databases work under the hood**
- **Learn Go by building a real system**
- **Create a portfolio project that stands out**
- **Demystify storage engines, indexing, and query parsing**

## 👤 Who is this for?
- Tech freshers who want to go beyond CRUD
- Developers curious about database internals
- Systems design interview preppers
- Self-learners who learn best by building

## 💡 Why is it valuable?
- Hands-on with file I/O, data structures, memory management, concurrency
- Demonstrates low-level systems skills
- Teaches concepts from real databases (SQLite, PostgreSQL, Redis)
- Impressive portfolio piece

---

## 🧑‍💻 User Stories

### Core
- Store, retrieve, update, and delete key-value pairs
- Data persists after restart
- Run simple queries and see indexing in action
- Handle multiple requests (concurrency basics)
- Use a simple CLI to interact with the DB
- Learn from clear, well-documented code

### Stretch
- SQL-like commands (SELECT, INSERT, etc.)
- Transactions (BEGIN, COMMIT, ROLLBACK)
- TCP server for network access
- Benchmarking and performance analysis

---

## 🗃️ Data Model (Sketch)
- **Record**: key, value, timestamp, deleted
- **Page**: page_id, page_type, records, next_page
- **Index**: name, field, B-Tree
- **Table**: name, columns, primary_key, indexes
- **Column**: name, data_type, nullable

---

## 🏗️ MVP Features
| Feature                  | Priority   |
|--------------------------|------------|
| In-memory key-value store| MUST HAVE  |
| File persistence         | MUST HAVE  |
| Basic CRUD operations    | MUST HAVE  |
| Simple CLI interface     | MUST HAVE  |
| B-Tree indexing          | MUST HAVE  |
| Page-based storage       | MUST HAVE  |
| WAL (Write-Ahead Log)    | NICE TO HAVE|
| SQL parser               | NICE TO HAVE|
| TCP server               | NICE TO HAVE|
| Transactions             | FUTURE     |
| Query optimizer          | FUTURE     |
| Replication              | FUTURE     |

---

## 🖥️ CLI Example
```
godb> SET user:1 "John Doe"
OK
godb> GET user:1
"John Doe"
godb> DELETE user:1
OK
godb> GET user:1
(nil)
godb> KEYS user:*
1) user:2
2) user:3
godb> INFO
Records: 2
Storage: 4.2 KB
Uptime: 00:05:32
```

---

## 🏛️ Architecture Overview

*(We plan to add an architecture diagram image here: `![Architecture Diagram](docs/architecture.png)`)*

```text
[ CLI ]
   |
[ Command Parser ]
   |
[ Execution Engine ]
   |---[ Index Manager (B-Tree) ]
   |---[ Buffer Pool ]
   |---[ Query Processor ]
   |
[ Storage Engine ]
   |
[ Data File (database.godb) ]
```

---

## 🛠️ Tech Stack
- **Language:** Go
- **Storage:** Custom binary file
- **Encoding:** encoding/binary, encoding/gob
- **CLI:** bufio + fmt (stdlib)
- **Testing:** testing (stdlib)
- **Build:** go build
- **Version Control:** Git + GitHub
- **Dependencies:** ZERO external packages for MVP

---

## 🗂️ Suggested Folder Structure
```
gofromscratchdb/
├── cmd/
│   └── godb/
│       └── main.go
├── internal/
│   ├── storage/
│   ├── index/
│   ├── engine/
│   ├── parser/
│   └── repl/
├── pkg/
│   └── types/
├── data/ (gitignored)
├── Makefile
├── go.mod
└── README.md
```

---

## 🏃‍♂️ Development Phases

| Phase | Title | Status | Key Features | Tests |
|-------|-------|--------|--------------|-------|
| 1 | Project Setup | ✅ Complete | Go module, folder structure, basic REPL | 5+ |
| 2 | In-Memory Store | ✅ Complete | Record struct, map storage, CRUD operations | 15+ |
| 3 | File Persistence | ✅ Complete | File format, save/load, binary encoding | 12+ |
| 4 | Page-Based Storage | ✅ Complete | Pages, buffer pool, paged file system | 20+ |
| 5 | B-Tree Index | ✅ Complete | B-Tree, insert/delete/search, integration | 80+ |
| 6 | Enhanced CLI | ✅ Complete | KEYS, INFO, history, errors, formatting | 36+ |
| 7 | Testing & Docs | ✅ Complete | 296+ tests, architecture guides, HTML docs | 296+ |

---

## 📊 Phase 6 Details (Latest)

### Features Implemented ✅
- **Command History**: Ring buffer with O(1) add/navigate (100 command capacity)
- **Error Handling**: 3-layer approach (parsing → validation → execution) with context-aware suggestions
- **Pretty Output**: Emoji indicators, Unicode boxes, aligned columns
- **KEYS Command**: Glob pattern matching (`*` = any chars, `?` = single char)
- **INFO Command**: 13+ metrics (storage, operations, cache, system)

### Documentation 📚
- `CLI_FEATURES_GUIDE.md` - 700+ lines with 30s/2m/5m interview pitches
- `REPL_ARCHITECTURE.html` - Interactive diagrams (36 KB, 3 Mermaid charts)
- `LOCK_CONTENTION_STRATEGY.md` - Concurrency analysis
- `BTREE_INTERVIEW_GUIDE.md` - 3,500+ line deep dive

### Code Quality 🎯
- **296+ tests**: All passing, 80%+ coverage
- **Performance**: 2.3M ops/sec (dual-lock strategy)
- **Zero dependencies**: Pure Go implementation
- **Production-ready**: Error handling, metrics, logging

---

## 📚 What You'll Learn
- Data structures (B-Tree), file I/O, binary encoding, memory management, algorithms, concurrency
- Go: structs, interfaces, error handling, testing, CLI, file ops
- Software engineering: design, modularity, testing, git, docs
- Interview prep: "How does a DB work?", "Explain B-Trees", "How is data persisted?"

---

## ✅ Progress Checklist

### Core Implementation
- [x] Project setup & structure
- [x] In-memory key-value store
- [x] File persistence (save/load)
- [x] Page-based storage with buffer pool
- [x] B-Tree indexing (insert/delete/search)
- [x] CRUD operations (Get, Set, Delete)
- [x] CLI with command parser
- [x] Command history (ring buffer)
- [x] Error handling (3-layer validation)
- [x] Pretty output formatting (emojis, boxes)
- [x] KEYS pattern matching (glob wildcards)
- [x] INFO metrics tracking (13+ stats)

### Quality Assurance
- [x] 296+ tests across all packages
- [x] 80%+ code coverage
- [x] Edge case testing
- [x] Concurrency testing with race detector
- [x] Performance benchmarks
- [x] Lock contention analysis

### Documentation
- [x] README with examples
- [x] CLI usage guide
- [x] Architecture documentation (HTML + Markdown)
- [x] Interview preparation guides (30s/2m/5m pitches)
- [x] B-Tree deep-dive (3,500+ lines)
- [x] Lock contention strategy guide
- [x] Code comments and type documentation
- [x] GitHub repository with commit history

### Portfolio & Presentation
- [x] Beautiful GitHub repository
- [x] Interactive architecture diagrams
- [x] Professional HTML documentation
- [x] Interview talking points prepared
- [x] Code quality and patterns demonstrated

---

## 🚀 Next Steps / Future Phases

### Phase 7: Production Extensions (Optional)
**Goal**: Extend CLI with production-grade features

- [ ] **Persistent History**: Save command history to `.godb_history` file
- [ ] **Command Aliases**: `alias shortcut fullcmd`
- [ ] **Macro Recording**: Record and replay command sequences
- [ ] **Shell Integration**: Bash/Zsh completion scripts
- [ ] **Configuration File**: `.godb.rc` for custom settings
- [ ] **Colored Output**: ANSI color support with terminal detection
- [ ] **Command Logging**: Structured logging for audit trails

### Phase 8: REST API & Server (Optional)
**Goal**: Add network access to database

- [ ] **HTTP REST API**: 
  - `POST /api/set` - Set key-value
  - `GET /api/get/:key` - Get value
  - `DELETE /api/delete/:key` - Delete key
  - `GET /api/keys/:pattern` - List keys
  - `GET /api/info` - Get statistics

- [ ] **TCP Server**: Raw protocol server for multiple clients
- [ ] **WebSocket Support**: Real-time updates
- [ ] **gRPC Service**: Type-safe RPC interface
- [ ] **Client Library**: Go client for easy integration

### Phase 9: Advanced Features (Optional)
**Goal**: Add advanced database features

- [ ] **Transactions**: BEGIN, COMMIT, ROLLBACK
- [ ] **Write-Ahead Log (WAL)**: Durability guarantee
- [ ] **TTL Support**: Key expiration
- [ ] **Snapshots**: Point-in-time recovery
- [ ] **Replication**: Multi-node support
- [ ] **Compression**: Data compression for storage
- [ ] **Encryption**: At-rest encryption

### Phase 10: Documentation & Learning (Optional)
**Goal**: Create comprehensive learning materials

- [ ] **Video Tutorials**: Step-by-step building guide
- [ ] **Blog Posts**: Design decisions and learnings
- [ ] **Interactive Demos**: Live coding sessions
- [ ] **Architecture Diagrams**: SVG diagrams for each component
- [ ] **Performance Analysis**: Benchmarking guide
- [ ] **Interview Prep Guide**: Comprehensive FAQ
- [ ] **Contributing Guide**: For community contributions

---

## 🎯 How to Use This Project

### For Learning
```bash
# Clone the repository
git clone https://github.com/izhan05803/DATABASE_CURATED_IN_GO.git
cd DATABASE_CURATED_IN_GO

# Run the database
go run ./cmd/godb

# Run all tests
go test ./... -v

# View architecture documentation
# Open REPL_ARCHITECTURE.html in your browser
```

### For Interviews
```bash
# Key talking points to prepare:
# 1. B-Tree indexing: "Balanced tree, O(log n) search/insert/delete"
# 2. Buffer pool: "LRU cache for page eviction, minimizes disk I/O"
# 3. Lock strategy: "Dual-lock approach separates metrics from critical path"
# 4. Pattern matching: "DP-based glob matcher, O(n*m) complexity"
# 5. Architecture: "REPL orchestrator with separation of concerns"

# Walk through REPL_ARCHITECTURE.html showing:
# - Component diagram
# - Execution flow (9 steps)
# - Responsibilities table
# - Evolution section (showing reusability)
```

### For Portfolio
```
Perfect for highlighting:
- System design skills
- Go proficiency
- Testing discipline (296+ tests)
- Documentation quality (10,000+ lines)
- Performance optimization (2.3M ops/sec)
- Clean architecture practices
```

---
