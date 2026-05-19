# Code Coverage Analysis and Improvement Plan

This report identifies areas in the `sqlfunc` repository where code coverage is lacking and provides a strategic plan for improvement.

## 1. Current Status
**Overall Statement Coverage: 84.9%**

### Coverage by Component
| Package | Coverage | Key Files |
| :--- | :--- | :--- |
| `github.com/dolmen-go/sqlfunc/sqlfunc-gen` | **0.0%** | `main.go` |
| `github.com/dolmen-go/sqlfunc/internal/sqlfuncgen` | **14.0%** (Total) | `fs.go` (71%), `gen.go` (83%), `utils.go` (95%) |
| `github.com/dolmen-go/sqlfunc/internal/genutils` | **2.6%** (Total) | `genutils.go` (67%) |
| `github.com/dolmen-go/sqlfunc` | **77.4%** | `any.go`, `scan.go`, `stmt.go` |

*Note: Percentages vary depending on whether `-coverpkg=./...` is used to include internal logic exercised by root tests.*

---

## 2. Identified Gaps

### A. CLI Tool (`sqlfunc-gen/`)
The main entry point is completely untested.
- **`mainErr`**: Logic for argument validation, calling the generator, and writing to the filesystem.

### B. Generator Logic (`internal/sqlfuncgen/gen.go`)
Core generation logic covers "happy paths" but skips many error conditions:
- **`checkTypeScope`**: Handling of `*types.TypeParam` (generics) and local types.
- **Invalid Signatures**:
    - `ForEach`: Multiple return values, or return types other than `error`/`bool`.
    - `Scan`: Missing `*sql.Rows` or returning values other than `error`.
    - `Stmt`: Missing `context.Context` or incorrect return types for `Exec`/`Query`/`QueryRow`.
- **Template Errors**: Failures in `printFuncDef` during template parsing or execution.

### C. Virtual File System (`internal/sqlfuncgen/fs.go`)
Used to hold generated code in memory. Several standard `fs.FS` methods are untested:
- **`ReadDir`**: Pagination logic (`n > 0`) and `EOF` handling.
- **`Stat`**: Metadata retrieval for files and directories.
- **`addFile`**: Panic conditions for invalid paths or subdirectories.
- **`rootDir.Read`**: Correctly returns `fs.ErrInvalid` but is never called.

### D. Utilities (`internal/genutils/genutils.go`)
- **`WriteFS`**: The "safety belt" check that prevents overwriting manual files (via `IsFileGenerated`) is untested.
- **Error Handling**: Failures during directory reading or file creation.

---

## 3. Proposed Testing Strategy

### Phase 1: Unit Test Expansion
1.  **Virtual FS (`internal/sqlfuncgen/fs_test.go`)**:
    - Use `testing/fstest.TestFS` to validate the implementation against the standard library's requirements.
    - Explicitly test `ReadDir` with various values of `n`.
2.  **Generator Error Handling (`internal/sqlfuncgen/gen_error_test.go`)**:
    - Create table-driven tests using `testdata/` containing invalid Go source code.
    - Verify that `Generate` logs the correct "SKIP" messages for invalid signatures and unsupported types.
3.  **Safety Utilities (`internal/genutils/genutils_test.go`)**:
    - Test `WriteFS` with a mock filesystem to verify that it refuses to overwrite files lacking the `// Code generated` header.

### Phase 2: Integration Testing
1.  **CLI Integration (`sqlfunc-gen/main_test.go`)**:
    - Implement a test that runs `mainErr` in a temporary directory.
    - Assert that files are generated correctly and that invalid flags trigger errors.

### Phase 3: Coverage Verification
- Standardize the use of `go test -coverpkg=./... ./...` to ensure integration tests contribute to the coverage metrics of internal packages.

---

## 4. Future Implementation Steps
1.  **Implement `fs_test.go`** for 100% coverage on the virtual FS.
2.  **Populate `testdata/`** with edge-case signatures to exercise all `SKIP` paths in `gen.go`.
3.  **Add `main_test.go`** to cover the CLI tool.
4.  **Enhance `WriteFS` tests** to ensure safety mechanisms are robust.
