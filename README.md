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
1. **Project Setup**: Repo, Go module, structure, README, Makefile
2. **In-Memory Store**: Record struct, map, basic commands, REPL
3. **File Persistence**: File format, save/load, persistence
4. **Page-Based Storage**: Pages, buffer pool, paged file
5. **B-Tree Index**: B-Tree, search/insert/delete, integration
6. **Enhanced CLI**: KEYS, INFO, HELP, history, formatting
7. **Testing & Docs**: >80% coverage, usage docs, architecture

---

## 📚 What You'll Learn
- Data structures (B-Tree), file I/O, binary encoding, memory management, algorithms, concurrency
- Go: structs, interfaces, error handling, testing, CLI, file ops
- Software engineering: design, modularity, testing, git, docs
- Interview prep: "How does a DB work?", "Explain B-Trees", "How is data persisted?"

---

## ✅ Quick Checklist
- [x] Clear goal
- [x] User stories
- [x] Data model
- [x] MVP features
- [x] CLI wireframe
- [x] Architecture
- [x] Tech stack
- [x] Dev phases

---

## 🏁 Next Steps
- Start with **Phase 1: Project Setup**
- Build incrementally, test constantly, document learnings, share progress!
