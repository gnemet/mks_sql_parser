/**
 * MKS SQL Parser - API Tester
 * A standalone script to test database connection and SQL execution via the backend REST API.
 *
 * Usage: node api_tester.js
 */

const CONFIG = {
  baseUrl: "http://localhost:8080",
  db: {
    name: "Manual",
    host: "localhost",
    port: "5432",
    user: "postgres",
    password: "password",
    database: "postgres",
    schema: "public",
  },
};

async function testConnection() {
  console.log(`Connecting to ${CONFIG.baseUrl}...`);
  try {
    const response = await fetch(`${CONFIG.baseUrl}/connect`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(CONFIG.db),
    });
    const data = await response.json();
    if (data.error) {
      console.error("❌ Connection Failed:", data.error);
      return false;
    }
    console.log("✅ Connected Successfully!");
    return true;
  } catch (err) {
    console.error("❌ Network Error:", err.message);
    return false;
  }
}

async function runQuery(sql, params = {}) {
  console.log(`\nExecuting Query: ${sql}`);
  try {
    const response = await fetch(`${CONFIG.baseUrl}/query`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sql: sql,
        params: params,
        order: Object.keys(params),
      }),
    });
    const data = await response.json();
    if (data.error) {
      console.error("❌ Query Error:", data.error);
      if (data.substituted_sql) {
        console.log("Attempted SQL:", data.substituted_sql);
      }
    } else {
      console.log("--- Substituted SQL ---");
      console.log(data.substituted_sql);
      console.log("\n--- Results ---");
      console.table(data.rows);
    }
  } catch (err) {
    console.error("❌ Network Error:", err.message);
  }
}

async function main() {
  const connected = await testConnection();
  if (!connected) return;

  // Sample Query matching mks-processing workflow
  const sampleSql = `
        SELECT 1 as id, 
        $1 ? 'key' as key_exists, 
        $1#>>'{key}' as key_value
    `;
  const sampleParams = { key: "api test value" };

  await runQuery(sampleSql, sampleParams);
}

main();
