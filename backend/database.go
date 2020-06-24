package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	_ "github.com/lib/pq"
)

func insertDomain(domain Domain) {
	// connect to database
	db, err := sql.Open("postgres", "postgresql://root@localhost:26257/servers_project?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("Connected to the database")
	if !isDomainPresent(db, domain.Name) {
		//insert
		fmt.Println("inserting into domain the domain named", domain.Name)
		ins := "INSERT INTO domains (name, servers_changed, ssl_grade, previous_ssl_grade, logo, title, is_down, time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
		values := domain.getValues()
		_, err = db.Exec(ins, values[0], "N/A", values[2], values[3], values[4], values[5], "false", time.Now().Hour())
		if err != nil {
			fmt.Println(err)
		}

	} else {
		fmt.Println("jk, the domain is already there")
		//update
		// check last time and see if more than one hour has passed
		ins := "SELECT time, ssl_grade from domains WHERE name = $1"
		t, err := db.Query(ins, domain.Name)
		if err != nil {
			log.Fatal(err)
		}
		defer t.Close()
		for t.Next() {
			var prevTime int
			var prev_ssl string
			if err := t.Scan(&prevTime, &prev_ssl); err != nil {
				log.Fatal(err)
			}
			// calculate time difference
			// currentTime := time.Now().Hour() // returns int, and prevtime is string
			currentTime := 3
			diff := math.Abs(float64(currentTime - prevTime))
			if diff >= 1 {
				fmt.Println("time difference is greater than one hour")
				// get old servers
				if prev_ssl != domain.Ssl_grade {
					// change values of domain we are inserting
					fmt.Println("ssl grade has changed")
					domain.Servers_changed = "true"
					domain.Previous_ssl_grade = prev_ssl
					// update query to update serverschanged, prev ssl grade, and ssl grade, and time
					ins := "UPDATE domains SET servers_changed = $1, ssl_grade = $2, previous_ssl_grade = $3, time = $4 WHERE name = $5"
					_, err := db.Exec(ins, domain.Servers_changed, domain.Ssl_grade, domain.Previous_ssl_grade, currentTime, domain.Name)
					if err != nil {
						log.Fatal(err)
					}

				} else {
					// only update time and set servers_changed to false
					fmt.Println("ssl grade has not changed")
					ins := "UPDATE domains SET servers_changed = $1, time = $2 WHERE name = $3"
					_, err := db.Exec(ins, "false", currentTime, domain.Name)
					if err != nil {
						log.Fatal(err)
					}

				}
			} else {
				// don't do anything
				fmt.Println("time difference is less than one hour")
			}
		}

	}

	// insert Endpoints
	for _, endpoint := range domain.Endpoints {
		if !isEndpointPresent(db, domain.Name, endpoint.IpAddress) {
			fmt.Println(domain.Name, " is not present")
			//insert
			ins := "INSERT INTO endpoints (name, ip_address, grade, country, owner) VALUES ($1, $2, $3, $4, $5)"
			values := endpoint.getValues()
			_, err = db.Exec(ins, domain.Name, values[0], values[1], values[2], values[3])
			if err != nil {
				fmt.Println(err)
			}

		} else {
			// only update if grade is not the same
			ins := "SELECT grade from endpoints WHERE name = $1 AND ip_address = $2"
			g, err := db.Query(ins, domain.Name, endpoint.IpAddress)
			if err != nil {
				log.Fatal(err)
			}
			defer g.Close()
			for g.Next() {
				var grade string
				if err := g.Scan(&grade); err != nil {
					log.Fatal(err)
				}
				if grade != endpoint.Grade {
					//update
					fmt.Println("grade endpoint has changed")
					ins = "UPDATE endpoints SET grade = $1 WHERE name = $2 AND ip_address = $3"
					_, err := db.Exec(ins, endpoint.Grade, domain.Name, endpoint.IpAddress)
					if err != nil {
						log.Fatal(err)
					}
				}
				fmt.Println("grade endpoint has not changed")
			}

		}
	}
	db.Close()

}

func isDomainPresent(db *sql.DB, name string) bool {
	// check if domain already exists
	rows, err := db.Query(`SELECT COUNT(*) FROM domains WHERE name = $1`, name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rows)
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf(name)
		if name == "0" {
			return false
		}
	}
	return true

}

func isEndpointPresent(db *sql.DB, name string, ip string) bool {
	rows, err := db.Query(`SELECT COUNT(*) FROM endpoints WHERE name = $1 AND ip_address = $2`, name, ip)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rows)
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf(name)
		if name == "0" {
			return false
		}
	}
	return true

}

// fetch information about servers
func fetchSSL(name string) (string, string, string) { // fetches servers_changed, ssl_grade, previous_ssl_grade
	var servers_changed, ssl_grade, previous_ssl_grade string
	db, err := sql.Open("postgres", "postgresql://root@localhost:26257/servers_project?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	ins := "SELECT servers_changed, ssl_grade, previous_ssl_grade from domains WHERE name = $1"
	g, err := db.Query(ins, name)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	for g.Next() {
		if err := g.Scan(&servers_changed, &ssl_grade, &previous_ssl_grade); err != nil {
			log.Fatal(err)
		}
	}
	db.Close()
	return servers_changed, ssl_grade, previous_ssl_grade

}

func fetchAll() []string {
	names := []string{}
	db, err := sql.Open("postgres", "postgresql://root@localhost:26257/servers_project?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	ins := "SELECT name from domains"
	n, err := db.Query(ins)
	if err != nil {
		log.Fatal(err)
	}
	defer n.Close()
	for n.Next() {
		var name string
		if err := n.Scan(&name); err != nil {
			log.Fatal(err)
		}
		names = append(names, name)
	}
	db.Close()
	return names
}
