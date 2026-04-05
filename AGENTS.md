# AGENTS.md - GoFromScratchDB

A database built from scratch in Go. **Key constraint: ZERO external dependencies.**

## Build/Lint/Test Commands

```bash
# Build
go build -o godb ./cmd/godb

# Test (always use -race flag)
go test ./...                         # Run all tests
go test -race ./...                   # With race detector (REQUIRED)
go test -v -run TestBTreeInsert ./... # Run single test by name
go test -v -run ^TestName$ ./internal/index  # Single test in package

# Benchmarks
go test -bench=. -benchmem ./...

# Lint
go fmt ./...                          # Format code
go vet ./...                          # Static analysis
```

## Project Structure

```
gofromscratchdb/
├── cmd/godb/main.go              # Entry point
├── internal/
│   ├── storage/                  # Page management, file I/O
│   ├── index/                    # B-Tree implementation
│   ├── engine/                   # Core database engine
│   ├── parser/                   # Lexer + command parser
│   └── repl/                     # CLI interface
├── pkg/types/types.go            # Shared types and interfaces
├── data/                         # Database files (gitignored)
└── go.mod
```

## Code Style

### Formatting & Imports
- Use `go fmt` (tabs, not spaces), ~100 char line limit
- Group imports: stdlib first, then internal packages, separated by blank line

```go
import (
    "encoding/binary"
    "fmt"

    "github.com/yourname/gofromscratchdb/internal/storage"
)
```

### Naming Conventions
- **Packages**: lowercase single word (`storage`, `index`)
- **Files**: lowercase with underscores (`btree_test.go`)
- **Exported**: PascalCase (`Record`, `NewEngine`)
- **Unexported**: camelCase (`pageSize`, `insertNode`)
- **Interfaces**: Often `-er` suffix (`Storage`, `Index`)

### Error Handling
- Always check errors explicitly
- Define sentinel errors for common cases
- Wrap with context: `fmt.Errorf("get key %q: %w", key, err)`

```go
var (
    ErrKeyNotFound = errors.New("key not found")
    ErrPageFull    = errors.New("page is full")
)
```

### Concurrency
- Use `sync.RWMutex` for read-heavy access
- All code must pass `go test -race ./...`
- Use channels for coordination (buffer pool LRU)

```go
type Engine struct {
    mu      sync.RWMutex
    storage Storage
}

func (e *Engine) Get(key string) ([]byte, error) {
    e.mu.RLock()
    defer e.mu.RUnlock()
    // ...
}
```

### Testing
- Place tests in same package (`storage_test.go`)
- Use table-driven tests
- Target >80% coverage
- Include benchmarks for performance-critical code

```go
func TestBTreeInsert(t *testing.T) {
    tests := []struct {
        name string
        keys []string
        want int
    }{
        {"empty", []string{}, 0},
        {"single", []string{"foo"}, 1},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tree := NewBTree(3)
            for _, key := range tt.keys {
                tree.Insert(key, 1)
            }
            if got := tree.Size(); got != tt.want {
                t.Errorf("Size() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Interfaces (define first in pkg/types/)

```go
type Storage interface {
    Get(key string) ([]byte, error)
    Set(key string, value []byte) error
    Delete(key string) error
    Scan(prefix string) ([]Record, error)
}

type Index interface {
    Search(key string) (uint32, bool)
    Insert(key string, pageID uint32) error
    Delete(key string) error
}
```

## Core Data Structures

```go
type Record struct {
    Key       string
    Value     []byte
    Timestamp int64
    Deleted   bool
}

type Page struct {
    PageID   uint32
    PageType PageType  // leaf/internal/overflow
    Records  []Record
    NextPage uint32
}
```

### Binary File Format
```
[Header: 100 bytes] magic="GODB", version, page_size(4096), total_pages, free_list_head, root_page
[Page 0: Metadata]
[Page 1: B-Tree Root]
[Page 2..N: Data Pages]
```

## CLI Commands
```
SET key value | GET key | DELETE key | KEYS pattern | INFO | HELP
```

## Architecture
```
CLI/REPL → Parser → Engine → Index(B-Tree) + BufferPool(LRU) → Storage → database.godb
```

## Common Gotchas

1. **Race detector**: Always test with `-race` flag
2. **File handles**: Use `defer f.Close()` after opening
3. **Binary encoding**: Use `encoding/binary` with explicit byte order
4. **Page size**: 4KB (4096 bytes)
5. **No external deps**: `go.mod` should have no `require` block

## Implementation Workflow

1. Define interface in `pkg/types/types.go`
2. Implement in `internal/<package>/`
3. Write tests alongside implementation
4. Run `go test -race ./...` frequently
5. Add benchmarks for hot paths

---

## Agent Skills

The following skills are available for planning, design, and refactoring workflows.

### /write-a-prd
Create a PRD through user interview, codebase exploration, and module design.

**Process:**
1. Ask for detailed problem description and solution ideas
2. Explore repo to verify assertions and understand current state
3. Interview relentlessly until reaching shared understanding
4. Sketch major modules - look for deep modules (encapsulate complexity behind simple interfaces)
5. Write PRD and submit as GitHub issue

**PRD Template sections:** Problem Statement, Solution, User Stories (extensive numbered list), Implementation Decisions, Testing Decisions, Out of Scope, Further Notes

### /prd-to-plan
Turn a PRD into a multi-phase implementation plan using tracer-bullet vertical slices.

**Process:**
1. Confirm PRD is in context
2. Explore codebase for architecture and patterns
3. Identify durable architectural decisions (routes, schema, models, auth)
4. Draft vertical slices - each cuts through ALL layers end-to-end
5. Quiz user on granularity and iterate
6. Write plan to `./plans/<feature>.md`

**Vertical slice rules:**
- Each slice delivers COMPLETE path through every layer (schema, API, UI, tests)
- Completed slice is demoable on its own
- Prefer many thin slices over few thick ones
- Include durable decisions, NOT specific file/function names

### /prd-to-issues
Break a PRD into independently-grabbable GitHub issues using vertical slices.

**Process:**
1. Fetch PRD from GitHub issue
2. Explore codebase if needed
3. Draft vertical slices (mark as HITL or AFK)
4. Quiz user on breakdown, dependencies, granularity
5. Create GitHub issues in dependency order

**Issue types:**
- **HITL**: Requires human interaction (architectural decision, design review)
- **AFK**: Can be implemented and merged without human interaction (prefer these)

**Issue template:** Parent PRD link, What to build, Acceptance criteria, Blocked by, User stories addressed

### /grill-me
Get relentlessly interviewed about a plan or design until every branch is resolved.

**Behavior:**
- Walk down each branch of the design tree
- Resolve dependencies between decisions one-by-one
- Provide recommended answer for each question
- Ask questions ONE at a time
- If answerable by exploring codebase, explore instead of asking

Use when you want to stress-test a plan or design before implementation.

### /design-an-interface
Generate multiple radically different interface designs using parallel sub-agents.

**Process:**
1. Gather requirements (problem, callers, operations, constraints)
2. Spawn 3+ parallel sub-agents with different constraints:
   - "Minimize method count (1-3 methods)"
   - "Maximize flexibility"
   - "Optimize for most common case"
   - "Take inspiration from [paradigm/library]"
3. Present each design: signature, usage example, what it hides
4. Compare on: simplicity, general-purpose vs specialized, efficiency, depth
5. Synthesize best elements

**Evaluation criteria (from "A Philosophy of Software Design"):**
- **Deep module** (good): Small interface hiding significant complexity
- **Shallow module** (avoid): Large interface with thin implementation

### /request-refactor-plan
Create a detailed refactor plan with tiny commits, filed as GitHub issue.

**Process:**
1. Ask for detailed problem description and solution ideas
2. Explore repo to verify assertions
3. Present alternative options
4. Interview about implementation details
5. Hammer out exact scope (what changes, what doesn't)
6. Check test coverage; discuss testing plans if insufficient
7. Break into tiny commits (each leaves codebase working)
8. Create GitHub issue with plan

**Template sections:** Problem Statement, Solution, Commits (detailed tiny-commit plan), Decision Document, Testing Decisions, Out of Scope, Further Notes

**Key principle:** "Make each refactoring step as small as possible, so that you can always see the program working." - Martin Fowler
