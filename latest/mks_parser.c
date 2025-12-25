#include "postgres.h"
#include "fmgr.h"
#include "utils/builtins.h"
#include "../mks_so/mks_sql.h"

PG_MODULE_MAGIC;

PG_FUNCTION_INFO_V1(mks_parser);

Datum
mks_parser(PG_FUNCTION_ARGS)
{
    text *sql_text_t = PG_GETARG_TEXT_PP(0);
    text *json_input_t = PG_GETARG_TEXT_PP(1);
    
    char *sql_str;
    char *json_str;
    char *result_str;
    text *result_t;
    
    // Convert TEXT to C string
    sql_str = text_to_cstring(sql_text_t);
    json_str = text_to_cstring(json_input_t);
    
    // Call the Go shared library function
    // Note: mksSql returns a C string allocated by C.CString in Go.
    // It is malloc'd, so we must free it.
    result_str = mksSql(sql_str, json_str);
    
    if (result_str == NULL) {
        PG_RETURN_NULL();
    }
    
    // Convert back to TEXT
    result_t = cstring_to_text(result_str);
    
    // Free the result from Go
    free(result_str);
    
    // text_to_cstring allocates memory using palloc, which is automatically freed by Postgres.
    // However, if we want to be explicit or if we were using malloc, we'd handle it.
    // Postgres memory context handles palloc'd memory.
    
    PG_RETURN_TEXT_P(result_t);
}
