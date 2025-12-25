# MKS SQL Parser Reference Guide

## 1. Introduction
The MKS SQL Parser is a PostgreSQL extension that allows dynamic SQL modification using a JSON input payload. It supports block logic, line filtering, and substitutions.

## 2. Installation
### Prerequisites
- PostgreSQL (Development headers)
- GCC/Clang
- Go (1.20+)

### Automated Build (Recommended)
Run the automation script:
```bash
./build_extension.sh
```
This requires `make`. If missing, install it (`sudo apt install make`).

### Manual Build
See `pg_extension/README.md`.

### Database Registration
```sql
CREATE EXTENSION mks_parser;
```

## 3. Usage
The function signature is:
```sql
mks_parser(sql_text text, input_json text) RETURNS text
```

### Example
```sql
SELECT mks_parser(
    'SELECT * FROM users WHERE active = true --<admin AND role = ''admin'' --admin>', 
    '{"admin": true}'
);
```

## 4. Parser Rules

### Block Logic
Conditionally include/exclude chunks of SQL.
- **Syntax**: `--<CONDITION` ... `-->`
- **JSONPath Block**: `--<{ EXPRESSION }` ... `-->`
- **Nested Path Block**: `--< $1 #>> '{a,b}' [= 'val'] >` ... `-->`
- **Simple Key Block**: `--< #'key' = 'val' >` ... `-->`



### Line Filters
Conditionally filter single lines.
- **Syntax**: `... -- #CONDITION`
- **JSONPath Line**: `... -- #{ EXPRESSION }`
- **Nested Path**: `$1 #>> '{a,b}'` (Existence) or `$1 #>> '{a,b}' = 'val'` (Value)
- **Simple Key**: `... -- #'key' = 'val'` (Supports `=,!=,~,~*,~~,%,%>`)





### Substitutions
- **Values**: `%key%` -> Replaced with value or empty string.
- **Parameters**: `:key` / `$key` -> Replaced with value or **line deleted** if missing.

### Minify Mode
If input JSON contains `"minify": true`:
- All comments are stripped.
- Deleted lines are fully removed.
- Output is compact.

For detailed syntax, see [parser_rules.md](parser_rules.md).
