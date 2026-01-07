# How to use the Tester

1.  **Select a Test Case**: Use the dropdown in the header to load pre-configured examples. This populates both SQL and JSON Editors.
2.  **Edit Code**:
    - **SQL Text**: Write your SQL query with MKS Parser syntax (e.g., `--<condition` blocks).
    - **Input JSON**: Define the variables/context for the parser.
3.  **Run Tools**:
    - **<i class="fas fa-play"></i> Run**: Processes the SQL via Wasm or Server API.
    - **<i class="fas fa-database"></i> Run & Query**: Processes the SQL and immediately executes the result against the connected database.
4.  **View Results**:
    - **Result**: Shows the final SQL after processing.
    - **SQL Result**: Displays tabular data from direct database queries.

### Components

#### Database Connection

- Select a predefined database from the dropdown or choose **Manual Connection**.
- Active connections enable the **Run & Query** button and the **SQL Result** pane.

#### SQL Log

- Tracks all operations, including "Launched SQL" (the exact SQL sent to Postgres).
- Use the <i class="fas fa-trash"></i> to clear and <i class="fas fa-copy"></i> to copy the history.

### Advanced Features

#### JSON to SQL Type Mapping

The parser supports direct mapping of JSON keys to Postgres types using special operators:

- `$1->^'key'`: Integer
- `$1->#'key'`: Numeric
- `$1->&'key'`: Boolean
- `$1->@'key'`: Timestamp
- `$1->^^'key'`: Integer Array
- `$1->##'key'`: Numeric Array
- `$1->&&'key'`: Boolean Array
- `$1->@@'key'`: Timestamp Array

#### Execution Strategies

- **EXECUTE Mode**: Fully parameterized query using `$1::jsonb` for the payload. Ideal for high performance and security.
- **COPY Mode**: Simulates Postgres COPY behavior for high-volume data ingestion tests.

#### Result Persistence

Your workspace state (SQL and JSON) is automatically saved to local storage, allowing you to resume work across page refreshes.

### Features

- **Minify**: Toggle the <i class="fas fa-compress-alt"></i> button to remove unnecessary whitespace from the output.
- **Execution Mode**: Configurable via `config.yaml` to use either `EXECUTE` (parameterized) or `COPY` strategy.
- **Drag & Drop**: Drag `.sql` or `.json` files directly into the editors.
- **Docs & Help**: Use the <i class="fas fa-book"></i> and <i class="fas fa-question-circle"></i> icons for guidance.
