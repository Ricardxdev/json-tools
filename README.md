# JSON Tools 🛠️

**A Go library for parsing, filtering, and analyzing JSON/YAML data. Extract typed collections, generate value catalogs, and infer schema details dynamically.
Features**

- Parse & Filter JSON/YAML:
    Convert unstructured map[string]any entries into strongly typed collections (map[string]T) with custom filters.

- Value Catalog Generation:
    Generate YAML files listing all unique values for each field in nested JSON/YAML structures (supports collections and index paths).

- Type Catalog Generation:
    Output YAML files describing observed Go types (e.g., string, int, []any) for every field in your data.

## Installation
```bash
go get github.com/your-username/json-tools
```
## Usage

### 1. Parsing and Filtering
```go
type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
}

rawData := map[string]any{
    "alice": map[string]any{"name": "Alice", "age": 30},
    "bob":   map[string]any{"name": "Bob", "age": "invalid"},
}

filters := map[string]func(any) bool{
    "age": func(v any) bool { _, ok := v.(int); return ok },
}

parsed, _ := jsonTools.ParseFilter[User](rawData, filters)
// Output: map[string]User with "alice" (valid age)
```

### 2. Generate Value Catalog
```go
data := map[string]any{
    "users": []any{
        map[string]any{"role": "admin", "active": true},
        map[string]any{"role": "user", "active": false},
    },
}

jsonTools.GenerateValueCatalog(data, "output/values.yaml", "users")
```

**Example values.yaml Output:**
```yaml
role:
  - admin
  - user
active:
  - true
  - false
```

### 3. Generate Type Catalog
```go
jsonTools.GenerateTypeCatalog(data, "output/types.yaml", "users")
```

**Example types.yaml Output:**
```yaml
role:
  - string
active:
  - bool
```

### Example Use Cases

- Validate and sanitize API payloads.

- Automate documentation for dynamic JSON/YAML configurations.

- Audit datasets for type consistency.