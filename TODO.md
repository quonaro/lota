# TODO / Ideas

## Config Import / Modular Task Definitions

### Problem Statement

Large projects inevitably outgrow a single `lota.yml`. When the file exceeds a few hundred lines, navigation becomes painful and collaboration on different functional areas (e.g., infra vs. build) creates unnecessary merge conflicts.

### Proposed Feature

Allow importing groups, commands, or entire configurations from external `*.yml` files into the main `lota.yml`.

### Syntax Options (Draft)

#### Option A: Explicit `import:` field on root level

```yaml
# lota.yml
import:
  deploy: ./tasks/deploy.yml
  infra:   ./tasks/infra.yml

vars:
  - APP_NAME=myapp

build:
  desc: Build the application
  script: go build .
```

Imported commands/groups become namespaced: `lota deploy up`, `lota infra docker logs`.

#### Option B: Implicit `lota.d/` directory

All `*.yml` files inside `lota.d/` are automatically loaded and merged into the root configuration as flat groups/commands.

```
project/
├── lota.yml
└── lota.d/
    ├── deploy.yml
    └── infra.yml
```

No syntax changes to `lota.yml` required.

#### Option C: Inline `import:` tag (consistent with existing `import:yaml` for vars)

```yaml
# lota.yml
build:
  desc: Build
  script: go build .

deploy:
  import: ./tasks/deploy.yml
```

### Open Questions / Design Challenges

1. **Path resolution for `dir`**: Should relative `dir` values inside imported files resolve against the imported file's directory or the main `lota.yml` directory? Both choices have surprising edge cases.

2. **Variable merge order**: When imported configs define their own `vars` / `args`, what is the merge priority? Does the main file override imports, or vice versa?

3. **Name collisions**: Two imported files both define a group named `test`. Should this error, merge, or be namespaced?

4. **Circular imports**: If `a.yml` imports `b.yml` and `b.yml` imports `a.yml`, the parser must detect the cycle and fail with a clear message.

5. **Upward config search**: How does `findConfigFile` behave when the user runs `lota` from a subdirectory and the main config imports relative paths?

6. **Dry-run semantics**: In dry-run mode, imported files should be listed as "would load from X" without executing.

### Prior Art & Lessons Learned

- **Task (Taskfile)** supports `includes` with namespaces. It is a frequent source of user confusion regarding variable scoping and path resolution.
- **Just** intentionally avoids imports entirely. Simplicity is a feature.
- **npm scripts** has no native import, but `lerna` / `nx` solve the monorepo problem at a different layer.

### Recommendation

**Deferred.** Current nested groups (`dev > run > test`) cover the vast majority of use cases. Re-evaluate when a real-world `lota.yml` demonstrably becomes unmaintainable (e.g., >300 lines). If implemented, prefer **Option A (explicit namespaces)** because it keeps the mental model explicit and avoids silent merges.

### Related Code Paths

- `config/loader.go` — `findConfigFile`, upward search logic.
- `config/parser.go` — `ParseConfig`, root-level key iteration.
- `config/types.go` — `AppConfig`, `Group`, `Command` structs.
