# Development Guidelines

## Build & Test Commands

- Build Go projects: `make build`
- Run tests: `make test`
- Run specific test: `go test -run TestName ./path/to/package`
- Run tests with coverage: `go test -cover ./...`
- Run linting: `make lint`
- Format code: `make format`
- Run code generation: `go generate ./...`
- Coverage report: `make coverage-report` (or `make coverage-report-html` for HTML output)
- On completion, use formatting (`make format`), tests (`make test`), and code generation (`go generate ./...`)
- Never commit without running completion sequence

## Important Workflow Notes

- Always run tests, format, and linter
- For linter use `make lint`
- Run tests, format, and linter after making significant changes to verify functionality
- Go version: 1.24+
- Don't add "Generated with Claude Code" or "Co-Authored-By: Claude" to commit messages or PRs
- Do not include "Test plan" sections in PR descriptions
- Do not add comments that describe changes, progress, or historical modifications. Avoid comments like "new function," "added test," "now we changed this," or "previously used X, now using Y." Comments should only describe the current state and purpose of the code, not its history or evolution.
- Use `go:generate` for generating mocks, never modify generated files manually. Mocks are generated with `github.com/vektra/mockery/v3`.
- After important functionality added, update README.md accordingly
- Always write unit tests instead of manual testing
- Don't manually test by running servers and using curl - write comprehensive unit tests instead

## Code Style Guidelines

- The codebase follows standard Go conventions. When in doubt, check the [Effective Go](https://golang.org/doc/effective_go) guide.
- Follow [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Use snake_case for filenames, camelCase for variables, PascalCase for exported names
- Group imports: standard library, then third-party, then local packages
- Error handling: check errors immediately and return them with context
- Use meaningful variable names; avoid single-letter names except in loops
- Validate function parameters at the start before processing
- Return early when possible to avoid deep nesting
- Prefer composition over inheritance
- Interfaces: Define interfaces in consumer packages
- Function size preferences:
  - Aim for functions around 50-60 lines when possible
  - Don't break down functions too small as it can reduce readability
  - Maintain focus on a single responsibility per function
- Comment style: in-function comments should be lowercase sentences
- Code width: keep lines under 120 characters when possible
- Format: Use `make format`
- Use existing structs from lower-level packages directly, don't duplicate them
  - When a struct is already defined in a lower-level package, use it directly instead of creating a duplicate definition
- Never add comments explaining what interface a struct implements - this is client-side concern
  - Don't write comments like "implements the Fetcher interface" - the consumer of the interface decides what implements it, not the provider
- In any file with structs and methods, order should be:
    1. Structs with methods first
    2. Interfaces after
    3. Data structs after

### Error Handling

- Check errors immediately after function calls
- Return detailed error information through wrapping
- Use `fmt.Errorf("context: %w", err)` to wrap errors with context

### Comments

- All comments inside functions should be lowercase
- Document all exported items with proper casing
- Use inline comments for complex logic
- Start comments with the name of the thing being described
- Comments documenting declarations should be full sentences and begin with the name of the thing being described and end in a period

### Testing

- Use table-driven tests where appropriate
- In table-driven tests, use maps of string to store test cases
- Use subtest with `t.Run()` to make test more structured
- Use `require` for fatal assertions, `assert` for non-fatal ones
- Use mock interfaces for dependency injection
- Test names follow pattern: `Test<Type>_<method>`
- Use separate packages for tests (e.g., `*_test`)

## Libraries

- CLI commands: `github.com/spf13/cobra`
- GitLab API client: `gitlab.com/gitlab-org/api/client-go`
- Testing: `github.com/stretchr/testify`
- Mock generation: `github.com/vektra/mockery/v3`
- Terminal UI: `github.com/briandowns/spinner` for loading indicators
- Table formatting: `github.com/jedib0t/go-pretty/v6` for table, csv, and json outputs
- To access libraries, figure how to use and check their documentation, use `go doc` command and `gh` tool
