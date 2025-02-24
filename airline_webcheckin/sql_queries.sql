# create database
CREATE DATABASE airline;

# create seats table with 120 rows
CREATE TABLE seats (
	id SERIAL PRIMARY KEY,
	seat_name VARCHAR(5) UNIQUE NOT NULL,
	passenger_name VARCHAR(20)
);

# query to simulate seat numbers in the table
INSERT INTO seats (seat_name)
SELECT 
    row_number || seat_letter
FROM 
    generate_series(1, 50) AS row_number, 
    unnest(ARRAY['A', 'B', 'C', 'D', 'E', 'F']) AS seat_letter;