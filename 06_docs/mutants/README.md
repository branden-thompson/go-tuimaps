# The mutants (v0.3.0 NFR-2, M4)

Every v0.3.0 requirement row is anchored to a mutant its test kills (D-7). A test that exists but cannot fail is caught here.

`mutants.json` is the table. Each row is one mutant:

| Field | What it holds |
|---|---|
| `id` | its own name, `m001` and on |
| `row` | the requirement it holds (`L-1.1`) |
| `what` | what the mutant breaks, in words |
| `file` | the file it changes |
| `old` | text found in that file exactly once |
| `new` | what replaces it |
| `package` | the package its test is in (`.` or `./internal/render`) |
| `test` | the test that must fail with the mutant in place |

Two tests read it:

- **`TestEveryMutantFindsItsLine`** (`mutants_test.go`) runs on every `go test`. Each mutant's text is in its file once, it changes something, its id is its own, and its test is declared in its package. A change that moves a mutant's line fails here; re-point the mutant in the same commit.
- **`TestEveryMutantIsKilled`** (`mutants_kill_test.go`, tag `mutants`) applies each mutant to a copy of the tree in a temporary directory, never to the tree itself. The copy must vet with the mutant in place: a mutant that does not compile is no evidence either way. The mutant's test must then fail. `scripts/gate`'s full lane runs it, and fails when no mutant ran. To run it alone: `go test -tags mutants -run TestEveryMutantIsKilled .`
