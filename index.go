package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-sql-driver/mysql"
)

type vehicle struct {
	TYPE  string
	PLATE string
	Entry string
	Exit  string
}

type parking struct {
	FLOOR   int
	NUMBER  int
	SECTION string
	state   bool
	PLATE   string
}

var vehicles [500]vehicle
var parkings [500]parking
var maxpark = 500
var vacantpark = 500
var occupark = 0

var db *sql.DB

func entry(w http.ResponseWriter, r *http.Request) {
	var park parking
	var nullPlate sql.NullString

	row := db.QueryRow(`SELECT FLOOR, SECTION, NUMBER, STATE, NUMBERPLATE FROM PARKING WHERE STATE=0 LIMIT 1;`)
	err := row.Scan(&park.FLOOR, &park.SECTION, &park.NUMBER, &park.state, &nullPlate)
	
	if err == sql.ErrNoRows {
		fmt.Fprintf(w, "No parking available!")
		return
	} else if err != nil {
		fmt.Fprintf(w, "Database error")
		return
	}

	park.PLATE = nullPlate.String

	if occupark < maxpark {
		r.ParseForm()
		plate := r.PostFormValue("plate")
		vehicleType := r.PostFormValue("type")

		var v vehicle
		v.PLATE = plate
		v.TYPE = vehicleType
		v.Entry = time.Now().Format("15:04:05")

		db.Exec("INSERT INTO VEHICLE (TYPE, PLATE, ENTRY, EXIT) VALUES (?, ?, ?, ?)", v.TYPE, v.PLATE, v.Entry, nil)
		db.Exec(`UPDATE PARKING SET STATE=1, NUMBERPLATE=? WHERE FLOOR=? AND SECTION=? AND NUMBER=?`, v.PLATE, park.FLOOR, park.SECTION, park.NUMBER)
		
		occupark = occupark + 1
		fmt.Fprintf(w, "Parking at Floor:%d Section: %s Number: %d Plate: %s ", park.FLOOR, park.SECTION, park.NUMBER, v.PLATE)
	} else {
		fmt.Fprintf(w, "No parking available!")
	}
}

func exit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	plate := r.PostFormValue("plate")

	var vType, vPlate, vEntry, vExit sql.NullString
	row := db.QueryRow(`SELECT TYPE, PLATE, ENTRY, EXIT FROM VEHICLE WHERE PLATE=?`, plate)
	err := row.Scan(&vType, &vPlate, &vEntry, &vExit)
	
	if err != nil {
		fmt.Fprintf(w, "Vehicle not found!")
		return
	}

	entryTimeStr := vEntry.String
	currentTimeStr := time.Now().Format("15:04:05")

	entryTime, _ := time.Parse("15:04:05", entryTimeStr)
	currentTime, _ := time.Parse("15:04:05", currentTimeStr)

	h := int(currentTime.Sub(entryTime).Hours())
	if h < 1 {
		h = 1
	}

	cost := 0
	if h == 1 {
		cost = 100
	} else {
		cost = 100 + 20*(h-1)
	}

	db.Exec("DELETE FROM VEHICLE WHERE PLATE = ?", plate)
	db.Exec(`UPDATE PARKING SET STATE=0, NUMBERPLATE=NULL WHERE NUMBERPLATE = ?`, plate)
	
	if occupark > 0 {
		occupark = occupark - 1
	}

	fmt.Fprintf(w, "The parking price is %d", cost)
}

func main() {
	cfg := mysql.NewConfig()
	cfg.User = "root"
	cfg.Passwd = "1q2w3e4r"
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "parking"

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	
	db.Exec("CREATE TABLE IF NOT EXISTS VEHICLE (TYPE varchar(255),PLATE varchar(255),ENTRY TIME,EXIT TIME);")
	db.Exec("CREATE TABLE IF NOT EXISTS PARKING (FLOOR INT, SECTION CHAR, NUMBER INT,STATE BOOLEAN, NUMBERPLATE varchar(255));")

	var count int
	db.QueryRow("SELECT COUNT(*) FROM PARKING").Scan(&count)
	if count == 0 {
		tx, _ := db.Begin()
		for i := 1; i < 251; i++ {
			tx.Exec("INSERT INTO PARKING VALUES (?,?,?,?,?)", 1, "A", i, 0, nil)
		}
		for i := 1; i < 251; i++ {
			tx.Exec("INSERT INTO PARKING VALUES (?,?,?,?,?)", 2, "B", i, 0, nil)
		}
		tx.Commit()
	}

	db.QueryRow("SELECT COUNT(*) FROM PARKING WHERE STATE = 1").Scan(&occupark)

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	mux := http.NewServeMux()
	mux.HandleFunc("/entry", entry)
	mux.HandleFunc("/exit", exit)

	serverAddr := ":8080"
	fmt.Printf("Server is running on http://localhost%s/entry\n", serverAddr)

	err = http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}