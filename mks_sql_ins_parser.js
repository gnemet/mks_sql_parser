// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.

/**
 * Parser for mks_sql_ins SQL wrapper files.
 */
class MksSqlInsParser {
    constructor() {
    }

    /**
     * Parses the content of an SQL file to find mks_sql_ins( info, sql ) call.
     * @param {string} content - The full file content
     * @returns {Object|null} - { info: Object, sql: string, fullMatch: string } or null
     */
    parse(content) {
        // Simple state machine to find mks_sql_ins(
        const startMarker = 'select mks_sql_ins(';
        const startIdx = content.indexOf(startMarker);
        if (startIdx === -1) return null;

        let idx = startIdx + startMarker.length;
        let balance = 1;
        let inString = false;
        let inDollarString = false;
        let dollarTag = '';
        let currentArg = '';
        let args = [];

        let startArgIdx = idx;

        while (idx < content.length && balance > 0) {
            const char = content[idx];
            const nextChar = content[idx + 1];

            // Handle strings
            if (!inString && !inDollarString) {
                if (char === "'") {
                    inString = true;
                } else if (char === '$') {
                    // Possible dollar quote start
                    const match = content.substring(idx).match(/^\$([a-zA-Z0-9_]*)\$/);
                    if (match) {
                        inDollarString = true;
                        dollarTag = match[0];
                        idx += dollarTag.length - 1; // Skip tag
                    }
                } else if (char === '(') {
                    balance++;
                } else if (char === ')') {
                    balance--;
                    if (balance === 0) {
                        // End of mks_sql_ins
                        args.push(content.substring(startArgIdx, idx));
                        break;
                    }
                } else if (char === ',' && balance === 1) {
                    // Argument separator
                    args.push(content.substring(startArgIdx, idx));
                    startArgIdx = idx + 1;
                }
            } else if (inString) {
                if (char === "'") {
                    // Check for escaped quote ''
                    if (nextChar === "'") {
                        idx++; // skip next
                    } else {
                        inString = false;
                    }
                }
            } else if (inDollarString) {
                // Check for closing dollar tag
                if (content.substring(idx).startsWith(dollarTag)) {
                    inDollarString = false;
                    idx += dollarTag.length - 1;
                }
            }
            idx++;
        }

        if (args.length < 2) return null;

        const infoRaw = args[0].trim();
        // The second argument is the SQL text, which might be followed by other arguments we typically ignore or don't use yet
        const sqlRaw = args[1].trim();

        return {
            info: this.parseInfo(infoRaw),
            sql: this.cleanSql(sqlRaw),
            infoRaw: infoRaw,
            sqlRaw: sqlRaw
        };
    }

    parseInfo(raw) {
        // Expect json_build_object( ... )
        const marker = 'json_build_object';
        const startIdx = raw.indexOf(marker);
        if (startIdx === -1) return {};

        // Extract content inside first level (...)
        let idx = startIdx + marker.length;
        // Find opening (
        while (idx < raw.length && raw[idx] !== '(') idx++;
        if (idx >= raw.length) return {};

        idx++; // skip (

        // Similar parsing logic for arguments of json_build_object
        let balance = 1;
        let inString = false;
        let inDollarString = false;
        let dollarTag = '';

        const args = [];
        let startArgIdx = idx;

        while (idx < raw.length && balance > 0) {
            const char = raw[idx];
            const nextChar = raw[idx + 1];

            if (!inString && !inDollarString) {
                if (char === "'") {
                    inString = true;
                } else if (char === '$') {
                    const match = raw.substring(idx).match(/^\$([a-zA-Z0-9_]*)\$/);
                    if (match) {
                        inDollarString = true;
                        dollarTag = match[0];
                        idx += dollarTag.length - 1;
                    }
                } else if (char === '(') {
                    balance++;
                } else if (char === ')') {
                    balance--;
                    if (balance === 0) {
                        args.push(raw.substring(startArgIdx, idx));
                        break;
                    }
                } else if (char === ',' && balance === 1) {
                    args.push(raw.substring(startArgIdx, idx));
                    startArgIdx = idx + 1;
                }
            } else if (inString) {
                if (char === "'") {
                    if (nextChar === "'") idx++;
                    else inString = false;
                }
            } else if (inDollarString) {
                if (raw.substring(idx).startsWith(dollarTag)) {
                    inDollarString = false;
                    idx += dollarTag.length - 1;
                }
            }
            idx++;
        }

        // Construct object from key-value pairs
        // args[0] = key, args[1] = value, args[2] = key, ...
        const result = {};
        for (let i = 0; i < args.length; i += 2) {
            let key = args[i].trim();
            let val = (args[i + 1] || '').trim();

            if (key.startsWith("'") && key.endsWith("'")) {
                key = key.substring(1, key.length - 1);
            }

            result[key] = this.parseValue(val);
        }

        return result;
    }

    parseValue(val) {
        // Check for literal string
        if (val.startsWith("'") && val.endsWith("'")) {
            // Basic unescape
            let s = val.substring(1, val.length - 1);
            return s.replace(/''/g, "'");
        }

        // Check for dollar quoted string (e.g. $json$ ... $json$)
        // It might be followed by cast or operators: $json${...}$json$::jsonb || ...
        // We only want the dollar-quoted part if it looks like a JSON block

        // Regex to match starting $tag$
        const dollarMatch = val.match(/^\$([a-zA-Z0-9_]*)\$/);
        if (dollarMatch) {
            const tag = dollarMatch[0];
            // Find closing tag
            // We need to be careful about not matching start tag again if nested? 
            // Postgres dollar quotes don't really nest with same tag.
            // But we should just look for the next occurrence of tag
            const contentStart = tag.length;
            const contentEnd = val.indexOf(tag, contentStart);

            if (contentEnd !== -1) {
                const rawContent = val.substring(contentStart, contentEnd);
                // Try parsing as JSON if the tag hints it or if it looks like JSON
                // The user specifically asks for default object extraction
                if (tag.includes('json') || rawContent.trim().startsWith('{')) {
                    try {
                        return JSON.parse(rawContent);
                    } catch (e) {
                        return rawContent; // Return raw string if parse fails
                    }
                }
                return rawContent;
            }
        }

        return val;
    }

    cleanSql(raw) {
        // Identify if it is wrapped in dollar quotes
        const m = raw.match(/^\$([a-zA-Z0-9_]*)\$([\s\S]*)\$\1\$/);
        if (m) {
            return m[2].trim();
        }
        // Handle single quotes
        if (raw.startsWith("'") && raw.endsWith("'")) {
            let s = raw.substring(1, raw.length - 1);
            return s.replace(/''/g, "'");
        }
        return raw;
    }
}
