# Rincewind Log

A terminal logging tool to log studying material or notes quickly and efficiently. Thereafter, one can search the logs in a database. It is also designed to sync to an external directory's markdown note where you can sync with a notetaking tool like Obsidian (and therfore with notes on Cloud). It however, need not be Obsidian.


# Libraries
1. Cobra
2. Google UUID
3. SQLITE3
4. sqlite3-vec
# rincewind_log

# Note for Windows compilation
This uses github.com/mattn/go-sqlite3 that expects a C compiler like `gcc` available. For windows, we might need to use `modernc.org/sqlite`