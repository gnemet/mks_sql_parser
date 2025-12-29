-- ============================================================
-- MKS SQL Parser Comprehensive Test Case
-- Covers all rules defined in config.yaml (IDs 1-12)
-- Input: test/all_rules.json
-- ============================================================

SELECT 'START_TEST' as status;

-- ============================================================
-- GROUP 1: BLOCK LOGIC (IDs 1, 2, 3, 4, 5)
-- ============================================================

-- [ID 1] Standard Block Logic
-- 1.1 Exists (Should be KEPT)
--<basic.exists
SELECT 'Block 1.1 KEPT' as rule_1_exists;
-->

-- 1.2 Not Exists (Should be KEPT)
--<!basic.missing_key
SELECT 'Block 1.2 KEPT' as rule_1_not_exists;
-->

-- 1.3 Value Match (Should be KEPT)
--<basic.value_match:match
SELECT 'Block 1.3 KEPT' as rule_1_value_match;
-->

-- 1.4 Value Mismatch (Should be REMOVED)
--<basic.value_match:wrong_value
SELECT 'Block 1.4 REMOVED' as rule_1_value_mismatch;
-->

-- [ID 2] JSONPath Block Logic
-- 2.1 Complex Expression (Should be KEPT)
--<{ $.jsonpath.a + $.jsonpath.b == 30 }
SELECT 'Block 2.1 KEPT' as rule_2_jsonpath_math;
-->

-- 2.2 Array Check (Should be KEPT)
--<{ $.jsonpath.list[0].active == true }
SELECT 'Block 2.2 KEPT' as rule_2_jsonpath_array;
-->

-- 2.3 False Expression (Should be REMOVED)
--<{ $.jsonpath.a > 100 }
SELECT 'Block 2.3 REMOVED' as rule_2_jsonpath_false;
-->

-- [ID 4] Nested Block (Postgres Syntax)
-- 4.1 Nested Exists (Should be KEPT)
--< $1 #> '{pg,nested,deep_key}' >
SELECT 'Block 4.1 KEPT' as rule_4_nested_exists;
-->

-- 4.2 Nested Value Match (Should be KEPT)
--< $1 #> '{pg,nested,number}' = '42' >
SELECT 'Block 4.2 KEPT' as rule_4_nested_value;
-->

-- [ID 5] Simple Extended Block Logic
-- 5.1 Regex Match (Should be KEPT)
--< #'ops.email' ~ '@gmail\.com$' >
SELECT 'Block 5.1 KEPT' as rule_5_regex;
-->

-- 5.2 LIKE Match (Should be KEPT)
--< #'ops.name' ~~ 'J%' >
SELECT 'Block 5.2 KEPT' as rule_5_like;
-->

-- 5.3 Not Equal (Should be KEPT)
--< #'ops.status' != 'inactive' >
SELECT 'Block 5.3 KEPT' as rule_5_not_equal;
-->

-- ============================================================
-- GROUP 2: LINE FILTERS (IDs 6, 7, 8, 10)
-- ============================================================

-- [ID 8] Standard Line Filters
SELECT 'Line 8.1 KEPT' as rule_8_exists;
-- #basic.exists
SELECT 'Line 8.2 REMOVED' as rule_8_exists_fail;
-- #basic.missing
SELECT 'Line 8.3 KEPT' as rule_8_not_exists;
-- #!basic.missing
SELECT 'Line 8.4 KEPT' as rule_8_val_match;
-- #basic.value_match:match
SELECT 'Line 8.5 REMOVED' as rule_8_val_fail;
-- #basic.value_match:wrong

-- [ID 6] JSONPath Line Filters
SELECT 'Line 6.1 KEPT' as rule_6_path_true;
-- #{ $.jsonpath.b > 10 }
SELECT 'Line 6.2 REMOVED' as rule_6_path_false;
-- #{ $.jsonpath.b < 10 }

-- [ID 7] Simple Extended Line Filters
SELECT 'Line 7.1 KEPT' as rule_7_regex;
-- #'ops.email' ~ 'test'
SELECT 'Line 7.2 REMOVED' as rule_7_regex_fail;
-- #'ops.email' !~ 'test'
SELECT 'Line 7.3 KEPT' as rule_7_like;
-- #'ops.desc' ~~ '%long%'

-- [ID 10] Nested Line Filters (Postgres Syntax)
SELECT 'Line 10.1 KEPT' as rule_10_nest_ex;
-- $1 #> '{pg,nested,deep_key}'
SELECT 'Line 10.2 KEPT' as rule_10_nest_val;
-- $1 #> '{pg,nested,number}' = '42'
SELECT 'Line 10.3 REMOVED' as rule_10_val_fail;
-- $1 #> '{pg,nested,number}' = '99'

-- ============================================================
-- GROUP 3: POSTGRES FILTERS (ID 9)
-- ============================================================

-- 9.1 Simple Key Access (Result must not be null/missing)
SELECT 1 FROM tmptbl WHERE id = $1 -> 'pg';
-- Kept because 'pg' exists

-- 9.2 Path Access
SELECT 1 FROM tmptbl WHERE val = $1#>>'{pg,simple_key}'; -- Kept

-- 9.3 Missing Key (Should be REMOVED)
SELECT 1 FROM tmptbl WHERE id = $1 -> 'non_existent_key';
-- REMOVED

-- ============================================================
-- GROUP 4: SUBSTITUTIONS (IDs 11, 12)
-- ============================================================

-- [ID 12] Replace or Empty (%key%)
SELECT '%sub.replace%' as rule_12_replaced;
-- Should be 'replaced_value'
SELECT '%sub.missing%' as rule_12_empty;
-- Should be ''

-- [ID 11] Replace or Delete (:key, $key)
SELECT:sub.replace as rule_11_colon_kept;
-- Should be replaced
SELECT:sub.missing as rule_11_colon_del;
-- Should be REMOVED
SELECT $sub.replace as rule_11_dollar_kept;
-- Should be replaced
SELECT $sub.missing as rule_11_dollar_del;
-- Should be REMOVED

SELECT 'END_TEST' as status;