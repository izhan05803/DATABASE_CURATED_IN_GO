# Windows Demo Quick Start Guide

**Get started immediately on Windows**

---

## ⚡ Quick Commands

### 1. Build
```powershell
go build -o godb.exe ./cmd/godb
```

### 2. Clean Database
```powershell
del database.godb
```

### 3. Run Demo
```powershell
godb.exe
```

---

## 📋 Copy-Paste Demo (Just paste into the terminal)

```
SET user:1 "Alice Johnson"
SET user:2 "Bob Smith"
SET product:100 "Laptop"
GET user:1
KEYS user:*
DELETE user:2
GET user:2
INFO
EXIT
```

---

## 🎯 Expected Output

```
✅ OK
✅ OK
✅ OK
"Alice Johnson"
(2 keys matched)
 1) user:1
 2) user:2
✅ OK
❌ GET key not found

[Beautiful INFO display with metrics]

👋 Goodbye!
```

---

## ✅ What Works

- ✅ Builds on Windows
- ✅ Runs as `godb.exe`
- ✅ All commands work
- ✅ Pretty output with Unicode
- ✅ Error handling works
- ✅ Metrics displayed correctly
- ✅ Data persists

---

## 🚀 Ready to Demo!

```powershell
# 1. Build
go build -o godb.exe ./cmd/godb

# 2. Clean
del database.godb

# 3. Run
godb.exe

# 4. Paste demo commands above
```

---

