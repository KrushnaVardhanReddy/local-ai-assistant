1. **Fix `go.mod` Go Version**
   - Revert Go version in `go.mod` from `1.26.0` to `1.25.0`.

2. **Fix Cross-Compilation Regression**
   - We need to handle `cpu.X86.HasAVX2` cleanly.
   - Refactor `profiler.go` to expose a generic interface or variable.
   - Create `profiler_amd64.go` with `//go:build amd64` that sets `hasAVX2 = cpu.X86.HasAVX2`.
   - Create `profiler_notamd64.go` with `//go:build !amd64` that sets `hasAVX2 = false`.
   - Update `profiler.go` and its tests to use this structured approach.

3. **Improve Robustness of Background Downloader**
   - Instead of checking if `parakeetDir` exists, check for a specific final file, e.g., `filepath.Join(parakeetDir, "model.onnx")`.
   - To make extraction atomic, we can extract to a temporary folder `parakeet.tmp` and then rename it to `parakeet`. Alternatively, just check for the specific file which is acceptable for this MVP. Let's check for `model.onnx`.

4. **Verify Tests**
   - Run all tests to ensure the changes are correct and coverage remains 100%.

5. **Complete pre-commit steps**
   - Rerun pre-commit instructions, request code review, and complete the check.
