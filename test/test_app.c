#include <stdio.h>
#include <stdlib.h>
#include "../mks_so/mks_sql.h"

char* read_file(const char* filename) {
    FILE *f = fopen(filename, "rb");
    if (!f) return NULL;
    fseek(f, 0, SEEK_END);
    long fsize = ftell(f);
    fseek(f, 0, SEEK_SET);

    char *string = malloc(fsize + 1);
    if (!string) {
        fclose(f);
        return NULL;
    }
    fread(string, 1, fsize, f);
    fclose(f);
    string[fsize] = 0;
    return string;
}

int main(int argc, char **argv) {
    if (argc < 3) {
        printf("Usage: %s <sql_file> <input_json_file>\n", argv[0]);
        return 1;
    }

    char *sql = read_file(argv[1]);
    if (!sql) {
        printf("Error reading SQL file: %s\n", argv[1]);
        return 1;
    }

    char *input = read_file(argv[2]);
    if (!input) {
        printf("Error reading JSON file: %s\n", argv[2]);
        free(sql);
        return 1;
    }

    // Call the shared library function
    // ensure mks_sql.h is available by building the .so first
    char *result = mksSql(sql, input);
    printf("%s\n", result);

    free(sql);
    free(input);
    free(result);

    return 0;
}
