package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

func getDBConn() *sql.DB {

	connStr := "user=allenkimanideep password=postgres dbname=airline sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %s", err)
	}
	defer db.Close()

	// Verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping the database: %s", err)
	}

	return db
}

func assignSeatUsingGeneralApproach(passengerName string) {
	db := getDBConn()
	defer db.Close()

	var seat_name string

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("error while begin transac %f", err)
	}

	err = tx.QueryRow("SELECT seat_name from seats where passenger_name is null order by id asc limit 1").Scan(&seat_name)
	if err != nil {
		log.Printf("Error while selecting row %v", err)
	}

	fmt.Printf("Assigning seat_name %v for passemger_name %v", seat_name, passengerName)

	_, err = tx.Exec("UPDATE seats set passenger_name = $1 where seat_name = $2", passengerName, seat_name)
	if err != nil {
		tx.Rollback()
		log.Printf("Error while updating seat for passenger %v error %v", passengerName, err)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error while commiting the transaction %v", err)
	}

}

func assignSeatUsingSelectUpdate(passengerName string) {
	db := getDBConn()
	defer db.Close()

	var seat_name string

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("error while begin transac %f", err)
	}

	err = tx.QueryRow("SELECT seat_name from seats where passenger_name is null order by id asc limit 1 FOR UPDATE").Scan(&seat_name)
	if err != nil {
		log.Printf("Error while selecting row %v", err)
	}

	fmt.Printf("Assigning seat_name %v for passemger_name %v\n", seat_name, passengerName)

	_, err = tx.Exec("UPDATE seats set passenger_name = $1 where seat_name = $2", passengerName, seat_name)
	if err != nil {
		tx.Rollback()
		log.Printf("Error while updating seat for passenger %v error %v", passengerName, err)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error while commiting the transaction %v", err)
	}

}

func assignSeatUsingSkipLock(passengerName string) {

	/*
		why use limit 1?
		so basically queryRows is selecting all rows with lock and returning only 1 row to update for me.
		initially when some go routines are spawned, all the rows are getting locked and when rest are spawned,
		they are unable to get any in the result set
	*/
	db := getDBConn()
	defer db.Close()

	var seat_name string

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("error while begin transac %f", err)
	}

	err = tx.QueryRow("SELECT seat_name from seats where passenger_name is null order by id asc limit 1 FOR UPDATE SKIP LOCKED").Scan(&seat_name)
	if err != nil {
		log.Printf("Error while selecting row %v", err)
	}

	fmt.Printf("Assigning seat_name %v for passemger_name %v\n", seat_name, passengerName)

	_, err = tx.Exec("UPDATE seats set passenger_name = $1 where seat_name = $2", passengerName, seat_name)
	if err != nil {
		tx.Rollback()
		log.Printf("Error while updating seat for passenger %v error %v", passengerName, err)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error while commiting the transaction %v", err)
	}

}

func main() {

	start := time.Now()

	file, err := os.Open("random_names.csv")
	if err != nil {
		log.Printf("error reading file bro %v", err)
		return
	}

	defer file.Close()

	// getDBConn()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("bro unable to convert import in csv reader %v", err)
		return
	}
	var wg sync.WaitGroup

	// tried with 500 records, db is resetting connections as it is overhelmed, hence took 120. can use connection pool for better things or singleton.
	for _, record := range records[:120] {
		wg.Add(1)
		go func(record []string) {
			defer wg.Done()

			/* in 1st approach we assigned only 48 rows(it can be random), reason is it tries to update same record as multiple transactions pick same record */
			// assignSeatUsingGeneralApproach(record[1])

			// in below approach, all seats will be assigned and sql internally re-evaluates query every time a locked transaction is released.
			// assignSeatUsingSelectUpdate(record[1])

			// in below approach, this is also same as above but lil faster as locked rows are skipped explicitly
			assignSeatUsingSkipLock(record[1])

		}(record)

	}

	wg.Wait()

	elapsed := time.Since(start)

	fmt.Printf("Time elapsed: %s\n", elapsed)

}
