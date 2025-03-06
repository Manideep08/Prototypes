/* Creating two tables to dump 1million entries with. 
    a. created 1m rows with auto inc as primary column.
        time taken:
            without index - 
                INSERT 0 1000000
                Time: 2042.227 ms (00:02.042)
            with index - reason is we need to sort things in index as well
                INSERT 0 1000000
                Time: 3139.795 ms (00:03.140)

    b. created 1m rows with uuid as primary key
        time taken:
            without index - 
                INSERT 0 1000000
                Time: 3595.414 ms (00:03.595)
            with index - 
                INSERT 0 1000000
                Time: 4943.787 ms (00:04.944)


    index size: 
            auto_inc_index_size 
                ---------------------
                    6456 kB
                    (1 row)

*/

-- Table with 4-byte auto-increment integer as primary key
CREATE TABLE table_auto_inc (
    ID SERIAL PRIMARY KEY,
    Age INT
);

-- Insert 1 million rows into the table with auto-increment
INSERT INTO table_auto_inc (Age)
SELECT (random() * 100)::int -- Random age between 0 and 99
FROM generate_series(1, 1000000);


-- Table with 16-byte UUID as primary key
CREATE TABLE table_uuid (
    ID UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    Age INT
);

-- Insert 1 million rows into the table with UUID
INSERT INTO table_uuid (Age)
SELECT (random() * 100)::int -- Random age between 0 and 99
FROM generate_series(1, 1000000);

-- Create index on the 'Age' column for both tables
CREATE INDEX idx_age_auto_inc ON table_auto_inc (Age);
CREATE INDEX idx_age_uuid ON table_uuid (Age);

-- Get the index size for the auto-increment table
SELECT pg_size_pretty(pg_relation_size('idx_age_auto_inc')) AS auto_inc_index_size;

-- Get the index size for the UUID table
SELECT pg_size_pretty(pg_relation_size('idx_age_uuid')) AS uuid_index_size;


-- Drop the index for the auto-increment table
DROP INDEX IF EXISTS idx_age_auto_inc;

-- Drop the index for the UUID table
DROP INDEX IF EXISTS idx_age_uuid;

