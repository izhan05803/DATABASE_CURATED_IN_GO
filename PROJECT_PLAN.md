# GoFromScratchDB - Project Plan

A database built from scratch in Go, designed as a learning project for tech freshers.

---

## Step 1: Define the Goal

### Why am I making this project?
- To deeply understand how databases actually work under the hood
- To learn Go by building something meaningful
- To have a portfolio project that demonstrates systems programming skills
- To understand concepts that are often "black boxes" (storage engines, indexing, query parsing)

### Who is it for?
- Tech freshers who want to go beyond CRUD applications
- Developers curious about database internals
- Anyone preparing for systems design interviews
- Self-learners who learn best by building

### What makes it valuable?
- Hands-on understanding of: file I/O, data structures, memory management, concurrency
- Demonstrates ability to work with low-level systems
- Teaches concepts used in real databases (SQLite, PostgreSQL, Redis)
- Impressive portfolio piece that stands out

### What problem does it solve?
- Bridges the gap between "using a database" and "understanding a database"
- Provides practical Go experience beyond web APIs

---

## Step 2: User Stories

### Core Stories (What a fresher would want to do)

**As a learner, I want to:**
1. Store key-value pairs so that I understand basic data persistence
2. Retrieve data by key so that I understand lookups
3. Delete data so that I understand data removal and cleanup
4. Update existing data so that I understand in-place modifications
5. See my data persist after restart so that I understand file storage
6. Run simple queries so that I understand query parsing
7. See how indexing speeds up searches so that I understand why indexes matter
8. Handle multiple requests so that I understand concurrency basics
9. Use a simple CLI to interact with the database so that I can test it easily
10. Understand each component through clear code so that I can learn from it

### Stretch Stories (After basics work)

**As an advanced learner, I want to:**
- Execute basic SQL-like commands (SELECT, INSERT, UPDATE, DELETE)
- See transaction basics (BEGIN, COMMIT, ROLLBACK)
- Connect via TCP so that I understand network protocols
- Benchmark performance so that I understand optimization

---

## Step 3: Define Data Models

### Internal Data Structures

```
Structure: Record
- key        (string, required, unique)
- value      ([]byte, required)
- timestamp  (int64, auto-generated)
- deleted    (bool, for soft deletes)

Structure: Page (disk storage unit)
- page_id    (uint32)
- page_type  (leaf/internal/overflow)
- records    ([]Record)
- next_page  (uint32, pointer)

Structure: Index
- name       (string)
- field      (string)
- tree       (B-Tree root pointer)

Structure: Table (for SQL-like features)
- name       (string)
- columns    ([]Column)
- primary_key (string)
- indexes    ([]Index)

Structure: Column
- name       (string)
- data_type  (string: int, string, bool, float)
- nullable   (bool)
```

### File Format (On Disk)

```
[Header: 100 bytes]
  - magic_number (4 bytes): "GODB"
  - version (4 bytes)
  - page_size (4 bytes)
  - total_pages (4 bytes)
  - free_list_head (4 bytes)
  - root_page (4 bytes)

[Page 0: Metadata]
[Page 1: Root of B-Tree]
[Page 2..N: Data Pages]
```

---

## Step 4: Nail the MVP

### Feature Prioritization

| Feature | Priority | Reason |
|---------|----------|--------|
| In-memory key-value store | MUST HAVE | Foundation - teaches maps, structs |
| File persistence | MUST HAVE | Core learning - teaches file I/O |
| Basic CRUD operations | MUST HAVE | Essential functionality |
| Simple CLI interface | MUST HAVE | Needed to interact with DB |
| B-Tree indexing | MUST HAVE | Key learning - teaches data structures |
| Page-based storage | MUST HAVE | Teaches how real DBs work |
| WAL (Write-Ahead Log) | NICE TO HAVE | Durability concept |
| SQL parser | NICE TO HAVE | Good learning but complex |
| TCP server | NICE TO HAVE | Networking basics |
| Transactions | FUTURE | Complex, after basics solid |
| Query optimizer | FUTURE | Advanced topic |
| Replication | FUTURE | Distributed systems |

### MVP Definition

**Version 0.1 - The Absolute Minimum:**
1. In-memory key-value store with GET, SET, DELETE
2. Persistence to a single file
3. CLI to run commands
4. Data survives restart

**Version 0.2 - Real Database Feels:**
1. B-Tree based storage
2. Page-based file format
3. Basic indexing

**Version 0.3 - SQL Flavor:**
1. Simple SQL parser (CREATE TABLE, INSERT, SELECT, DELETE)
2. Multiple tables support
3. WHERE clause filtering

---

## Step 5: Wireframe (CLI Interface)

Since this is a CLI database, the "wireframe" is the command interface:

```
┌─────────────────────────────────────────────────┐
│  GoFromScratchDB v0.1                           │
│  Type 'help' for commands, 'exit' to quit       │
├─────────────────────────────────────────────────┤
│                                                 │
│  godb> SET user:1 "John Doe"                    │
│  OK                                             │
│                                                 │
│  godb> GET user:1                               │
│  "John Doe"                                     │
│                                                 │
│  godb> DELETE user:1                            │
│  OK                                             │
│                                                 │
│  godb> GET user:1                               │
│  (nil)                                          │
│                                                 │
│  godb> KEYS user:*                              │
│  1) user:2                                      │
│  2) user:3                                      │
│                                                 │
│  godb> INFO                                     │
│  Records: 2                                     │
│  Storage: 4.2 KB                                │
│  Uptime: 00:05:32                               │
│                                                 │
└─────────────────────────────────────────────────┘
```

### Command Flow

```
User Input
    │
    v
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│  Parse  │───>│ Execute │───>│ Storage │───>│ Response│
│ Command │    │  Logic  │    │  Layer  │    │  Format │
└─────────┘    └─────────┘    └─────────┘    └─────────┘
```

---

## Step 6: Understand the Project's Future

### Project Type: **Learning Project / Portfolio Piece**

| Question | Answer |
|----------|--------|
| Quick prototype or production? | Learning project, but with production-quality code practices |
| Scale for many users? | No, single-user is fine |
| Expected lifespan? | Ongoing learning, can expand over months |
| Other developers? | Solo, but code should be readable for portfolio |

### Approach Decision: **Educational Production MVP**
- Clean, well-documented code (others will read it)
- Proper error handling (learn good habits)
- Unit tests for core components (learn testing)
- No premature optimization (understand basics first)
- Modular design (easy to extend for learning)

---

## Step 7: Determine Components & Architecture

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────┐
│                      CLI Interface                        │
│                   (REPL / Command Line)                   │
└─────────────────────────┬────────────────────────────────┘
                          │
                          v
┌──────────────────────────────────────────────────────────┐
│                    Command Parser                         │
│              (Tokenize & Parse Commands)                  │
└─────────────────────────┬────────────────────────────────┘
                          │
                          v
┌──────────────────────────────────────────────────────────┐
│                   Execution Engine                        │
│            (Execute Commands, Business Logic)             │
└─────────────────────────┬────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          v               v               v
┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│   Index     │   │   Buffer    │   │    Query    │
│  Manager    │   │    Pool     │   │  Processor  │
│  (B-Tree)   │   │  (Caching)  │   │ (Filtering) │
└──────┬──────┘   └──────┬──────┘   └─────────────┘
       │                 │
       v                 v
┌──────────────────────────────────────────────────────────┐
│                    Storage Engine                         │
│              (Page Management, File I/O)                  │
└─────────────────────────┬────────────────────────────────┘
                          │
                          v
┌──────────────────────────────────────────────────────────┐
│                     Data File                             │
│                   (database.godb)                         │
└──────────────────────────────────────────────────────────┘
```

### Component Breakdown

| Component | Responsibility | What Fresher Learns |
|-----------|---------------|---------------------|
| CLI/REPL | Read user input, display output | I/O handling, user experience |
| Parser | Tokenize commands, validate syntax | String parsing, state machines |
| Execution Engine | Route commands, coordinate components | Architecture patterns |
| Index Manager | B-Tree operations, key lookups | Data structures, algorithms |
| Buffer Pool | Cache pages in memory | Memory management, LRU cache |
| Storage Engine | Read/write pages to disk | File I/O, binary encoding |
| Query Processor | Filter and project data | Logic implementation |

---

## Step 8: Pick Your Stack

| Layer | Technology | Reason |
|-------|------------|--------|
| Language | Go | Learning goal, great for systems programming, simple concurrency |
| Storage | Custom binary file | Learn file formats, no dependencies |
| Encoding | encoding/binary, encoding/gob | Learn serialization |
| CLI | bufio + fmt (stdlib) | Keep it simple, no external deps |
| Testing | testing (stdlib) | Learn Go testing patterns |
| Build | go build | Standard Go toolchain |
| Version Control | Git + GitHub | Portfolio visibility |

### Why Go is Perfect for This

1. **Simple syntax** - Focus on concepts, not language quirks
2. **Great stdlib** - File I/O, encoding, testing built-in
3. **Compiled** - Understand how compiled languages work
4. **Concurrency** - Goroutines for future multi-client support
5. **Industry relevant** - Used in real databases (CockroachDB, InfluxDB)

### Dependencies: ZERO external packages for MVP
This forces learning the fundamentals without framework magic.

---

## Step 9: Development Process

### Phase 1: Project Setup (Day 1)

- [ ] Create GitHub repository
- [ ] Initialize Go module: `go mod init github.com/username/gofromscratchdb`
- [ ] Create folder structure (see below)
- [ ] Set up .gitignore
- [ ] Create README.md with project description
- [ ] Set up basic Makefile

**Folder Structure:**
```
gofromscratchdb/
├── cmd/
│   └── godb/
│       └── main.go          # Entry point
├── internal/
│   ├── storage/
│   │   ├── page.go          # Page structure
│   │   ├── file.go          # File operations
│   │   └── storage_test.go
│   ├── index/
│   │   ├── btree.go         # B-Tree implementation
│   │   └── btree_test.go
│   ├── engine/
│   │   ├── engine.go        # Core database engine
│   │   └── engine_test.go
│   ├── parser/
│   │   ├── lexer.go         # Tokenizer
│   │   ├── parser.go        # Command parser
│   │   └── parser_test.go
│   └── repl/
│       └── repl.go          # CLI interface
├── pkg/
│   └── types/
│       └── types.go         # Shared types
├── data/                    # Database files (gitignored)
├── Makefile
├── go.mod
└── README.md
```

### Phase 2: In-Memory Store (Day 2-3)

- [ ] Implement basic Record struct
- [ ] Create in-memory store using Go map
- [ ] Implement SET command
- [ ] Implement GET command
- [ ] Implement DELETE command
- [ ] Write unit tests for each operation
- [ ] Create simple REPL loop

### Phase 3: File Persistence (Day 4-5)

- [ ] Design file header format
- [ ] Implement file creation
- [ ] Implement write to file (serialize records)
- [ ] Implement read from file (deserialize records)
- [ ] Add SAVE command
- [ ] Add LOAD on startup
- [ ] Test data survives restart

### Phase 4: Page-Based Storage (Day 6-8)

- [ ] Define Page struct (4KB pages)
- [ ] Implement page serialization
- [ ] Implement page deserialization
- [ ] Create page manager (allocate, free, read, write)
- [ ] Implement buffer pool (cache pages in memory)
- [ ] Replace simple file with paged storage
- [ ] Write tests for page operations

### Phase 5: B-Tree Index (Day 9-12)

- [ ] Implement B-Tree node structure
- [ ] Implement B-Tree search
- [ ] Implement B-Tree insert
- [ ] Implement B-Tree delete
- [ ] Implement B-Tree split/merge
- [ ] Integrate B-Tree with storage engine
- [ ] Write comprehensive B-Tree tests
- [ ] Benchmark lookups (before/after indexing)

### Phase 6: Enhanced CLI & Commands (Day 13-14)

- [ ] Implement KEYS command (list keys with pattern)
- [ ] Implement INFO command (stats)
- [ ] Implement HELP command
- [ ] Add command history
- [ ] Improve error messages
- [ ] Add pretty output formatting

### Phase 7: Testing & Documentation (Day 15)

- [ ] Achieve >80% test coverage
- [ ] Write README with usage examples
- [ ] Document architecture in /docs
- [ ] Add inline code comments explaining concepts
- [ ] Create example scripts

---

## What a Tech Fresher Learns

### Core Computer Science Concepts

| Concept | Where in Project | Real-World Application |
|---------|-----------------|----------------------|
| Data Structures | B-Tree implementation | Every database uses trees |
| File I/O | Storage engine | All persistent systems |
| Binary Encoding | Page serialization | Network protocols, file formats |
| Memory Management | Buffer pool | Caching, performance |
| Algorithms | Search, insert, delete | Everywhere |
| Concurrency | Future: multi-client | Web servers, distributed systems |

### Go-Specific Skills

- Structs and interfaces
- Error handling patterns
- Package organization
- Testing with `go test`
- Working with bytes and binary data
- File operations with `os` package
- Building CLI applications

### Software Engineering Skills

- Designing before coding
- Breaking problems into components
- Writing testable code
- Git workflow
- Documentation
- Debugging complex systems

### Interview Preparation

This project gives you talking points for:
- "How does a database work internally?"
- "Explain B-Trees and why databases use them"
- "How is data persisted to disk?"
- "What is a buffer pool?"
- "How would you design a key-value store?"

---

## Quick Reference Checklist

- [x] Clear goal statement
- [x] User stories covering core functionality
- [x] Data model sketch
- [x] MVP feature list (ruthlessly trimmed)
- [x] Basic wireframes (CLI interface)
- [x] Understanding of project scale/longevity
- [x] Architecture diagram
- [x] Tech stack decisions with rationale
- [x] Development phases outlined

---

## Next Steps

1. **Start with Phase 1** - Set up the project structure
2. **Build incrementally** - Get each phase working before moving on
3. **Test constantly** - Write tests as you build
4. **Document learnings** - Keep notes on what you learn
5. **Share progress** - Commit often, write good commit messages

**Ready to start coding? Begin with Phase 1: Project Setup!**
