# MKS SQL Parser Rules

This document outlines the syntax and usage of the MKS SQL Parser features implemented in the `mks_sql.so` shared library. The parser processes SQL text using a provided JSON input to conditionally include, exclude, or modify lines of code.

## 1. Block Logic
Blocks allow you to conditionally include or exclude multiple lines of SQL based on the presence or value of keys in the input JSON.

**Syntax:**
- Start: `--<CONDITION`
- End: `-->` (or `-->TAG`)

**Conditions:**
| Syntax          | Logic           | Description                                                       |
| :-------------- | :-------------- | :---------------------------------------------------------------- |
| `--<key`        | **Exists**      | Block is **kept** if `key` exists in JSON.                        |
| `--<!key`       | **Not Exists**  | Block is **kept** if `key` does **NOT** exist.                    |
| `--<key:value`  | **Equals**      | Block is **kept** if `key` exists AND equals `value`.             |
| `--<key:!value` | **Not Equals*** | Block is **kept** if `key` exists AND does **NOT** equal `value`. |

*> Note: Strict equality check. If `key` is missing, equivalence checks usually fail (block skipped).*

**Example:**
```sql
--<version:4
    , $1->>'version' as version  -- Included only if input "version" is "4"
--version>

--<!banana
    , 'banana not exists' as "banana" -- Included only if input "banana" is MISSING
--banana>
```

---

## 2. Line Filter Tags
Line tags allow you to conditionally filter a single line of SQL. Tags are placed as comments at the end of the line.

**Syntax:**
- `... SQL CODE ... -- #CONDITION`

**Conditions:**
| Syntax       | Logic          | Description                                   |
| :----------- | :------------- | :-------------------------------------------- |
| `#key`       | **Exists**     | Line is **kept** if `key` exists.             |
| `#!key`      | **Not Exists** | Line is **kept** if `key` does **NOT** exist. |
| `#key:value` | **Equals**     | Line is **kept** if `key` equals `value`.     |

**Example:**
```sql
, 'longKey'   as "keyCheck"      -- #kulcs   (Kept if "kulcs" exists)
, true        as "existsAddress" -- #checkAddress:true (Kept if "checkAddress" is "true")
, false       as "noAddress"     -- #!checkAddress (Kept if "checkAddress" is MISSING)
```

---

## 3. Substitutions
The parser can replace placeholders in the SQL text with values from the JSON input.

### A. Value or Empty (`%key%`)
Replaces the placeholder with the value. If the key is missing, it is replaced with an empty string (or simple space/removal), but the **line is kept**.

| Syntax  | Description                                           |
| :------ | :---------------------------------------------------- |
| `%key%` | Replaces with `input[key]`. If missing, becomes `""`. |

**Example:**
```sql
, '%value%' as "replaceValue" 
-- If input["value"] = "foo", becomes: 'foo' as "replaceValue"
-- If input["value"] is missing, becomes: '' as "replaceValue"
```

### B. Value or Delete (`:key` / `$key`)
Replaces the placeholder with the value. If the key is missing, the **entire line is deleted**. This is useful for SQL parameters like `:limit` or `:offset`.

| Syntax | Description                                              |
| :----- | :------------------------------------------------------- |
| `:key` | Replaces with `input[key]`. If missing, **Delete Line**. |
| `$key` | Same behavior. *(Note: `$1`, `$2` etc. are ignored)*     |

**Example:**
```sql
limit :limit
offset :offset
-- If "limit" is 10 but "offset" is missing:
-- limit 10
-- -- deleted: offset :offset
```

---

## 4. PostgreSQL JSON Filters (Legacy)
The parser supports legacy PostgreSQL-style JSON operator checks.

**Syntax:**
- `$1->'key'` or `$1->>'key'`

**Logic:**
If the referenced key (e.g., `key` inside `$1->'key'`) is **missing** from the input JSON, the entire line is **deleted**.

**Example:**
```sql
and mk.id = ( $1->>'id' )
-- If "id" is missing in input, this line is commented out (deleted).
```

---

## 5. JSONPath Logic
The parser supports conditional logic using boolean expressions with JSONPath selectors. This is similar to PostgreSQL's `jsonb_path_match`.

**Syntax:**
- Block: `--<{ EXPRESSION }` ... `-->`
- Line: `... -- #{ EXPRESSION }`

**Expression Logic:**
- Uses standard operators: `&&` (AND), `||` (OR), `==` (Equals), `!=` (Not Equals), etc.
- `$` represents the root input JSON object.
- `exists($.key)` checks if a key exists in the input.

**Example:**
```sql
--<{ exists($.key1) || exists($.key2) }
    , 'key1 or key2' as "isExistOrBlock"
-->

, 'burgonya "ok" exists' as "burgonya"   -- #{ $.burgonya == "ok" }
```

---


## Summary of Action Priorities
1.  **Block Logic** (Standard & JSONPath): Checked first.
2.  **Line Tags** (Standard & JSONPath): Checked next. If condition fails, line is deleted.
3.  **Substitutions**: 
    -   `:key` / `$key`: If key missing -> Delete line.
    -   `%key%`: If key missing -> Empty string (Line kept).
4.  **Legacy Filters**: `$1->'key'` missing -> Delete line.

---

## 6. Minify Mode
If the input JSON contains `"minify": true`, the output SQL will be cleaned:
- Deleted lines are completely removed (not commented out).
- Comments (starting with `--`) are stripped from kept lines.
- Empty lines are skipped.

This produces a compact SQL result suitable for execution.
