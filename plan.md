1.  **Create Directory:** Create `wails-app/adapters/llm` if it does not exist.
2.  **Create `openai_adapter.go`:** Create `wails-app/adapters/llm/openai_adapter.go` defining `OpenAIAdapter` which wraps the existing `wails-app/backend/llm` package functions behind the `driven.LLMPort` interface.
3.  **Create `mock_adapter.go`:** Create `wails-app/adapters/llm/mock_adapter.go` defining `MockLLMAdapter` which is a test double that replays scripted responses and checks inputs.
4.  **Create `openai_adapter_test.go`:** Create `wails-app/adapters/llm/openai_adapter_test.go` to test `OpenAIAdapter` and/or `MockLLMAdapter` (based on the instruction, the tests use `MockLLMAdapter` to verify things like `StreamCompletion` calls `onToken`, `onDone`, records calls, returns `Err`, and tests `ChatMessage` conversion for `OpenAIAdapter`). We'll make sure there is 100% coverage for the `adapters/llm` package.
5.  **Pre-commit steps:** Run tests and format the code.
6.  **Submit:** Commit and submit the code.
