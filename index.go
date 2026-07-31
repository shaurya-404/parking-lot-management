package main

import (
	"database/sql"
	"fmt"
	_ "fmt"
	"log"
	"net/http"
	"strconv"
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

func entry(w http.ResponseWriter, r *http.Request) {
	var park parking
	row, err := db.Query(`SELECT * FROM PARKING WHERE STATE="FALSE";`)
	_ = err
	if occupark < 500 && row.Next() {
		row.Scan(park.FLOOR, park.SECTION, park.NUMBER, park.state, park.PLATE)
		occupark = occupark + 1
		var v vehicle
		r.ParseForm()
		plate := r.PostFormValue("plate")
		vehicleType := r.PostFormValue("type")
		v.PLATE = plate
		v.Entry = time.Now().Format("15:04:05")
		v.TYPE = vehicleType
		db.Exec("INSERT INTO VEHICLE VALUES (?,?,?,?)", v.TYPE, v.PLATE, v.Entry, 0)
		db.Exec(`UPDATE PARKING SET STATE="TRUE" WHERE FLOOR = ? AND SECTION=? AND NUMBER = ?`, park.FLOOR, park.SECTION, park.NUMBER)
		db.Exec(`UPDATE PARKING SET PLATE="?" WHERE FLOOR = ? AND SECTION=? AND NUMBER = ?`, park.FLOOR, park.SECTION, park.NUMBER)
		fmt.Println("Parking at Floor:%s Section: %s Number: %d Plate: %s ", park.FLOOR, park.SECTION, park.NUMBER, park.PLATE)
	} else {
		fmt.Println("No parking available!")
	}
}

func exit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	plate := r.PostFormValue("plate")
	row, err := db.Query(`SELECT * FROM VEHICLE WHERE PLATE="?"`, plate)
	_ = err
	var a, b, c, d string
	row.Scan(&a, &b, &c, &d)
	c = c[:2]
	timer := time.Now().Format("15:04:05")
	a1, err := strconv.Atoi(timer[:2])
	a2, err := strconv.Atoi(c)
	h := a1 - a2
	cost := 0
	if h == 1 {
		cost = 100
	} else {
		cost = 100 + 20*(h-1)
	}
	fmt.Println("The parking price is %d", cost)
}

func delete(w http.ResponseWriter, r *http.Request) {
	var v vehicle
	r.ParseForm()
	plate := r.PostFormValue("plate")
	v.PLATE = plate
	db.Exec("DELETE FROM VEHICLE WHERE PLATE = ?", v.PLATE)
	db.Exec(`UPDATE PARKING SET STATE="FALSE" where PLATE = ?`, v.PLATE)
	db.Exec(`UPDATE PARKING SET PLATE=NULL where PLATE = ?`, v.PLATE)
	fmt.Println("Parking deleted!")
}

// func helloHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/hello" {
// 		http.NotFound(w, r)
// 		return
// 	}

// 	fmt.Fprint(w, "Hello, welcome to my Go HTTP server!")
// }
var db *sql.DB

func main() {
	cfg := mysql.NewConfig()
	cfg.User = "root"
	cfg.Passwd = "1q2w3e4r"
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "parking"

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	_ = err
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("CREATE TABLE IF NOT EXISTS VEHICLE (TYPE varchar(255),PLATE varchar(255),ENTRY TIME,EXIT TIME);")
	db.Exec(("CREATE TABLE IF NOT EXISTS PARKING (FLOOR INT, SECTION CHAR, NUMBER INT,STATE BOOLEAN, NUMBERPLATE varchar(255));"))
	for i := 1; i < 251; i++ {
		db.Exec("INSERT INTO PARKING VALUES (?,?,?,?,?)", 1, "A", i, "FALSE", "NULL")
	}
	for i := 1; i < 251; i++ {
		db.Exec("INSERT INTO PARKING VALUES (?,?,?,?,?)", 2, "B", i, "FALSE", "NULL")
	}
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
	mux := http.NewServeMux()

	//mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/entry", entry)
	mux.HandleFunc("/exit", exit)
	mux.HandleFunc("/delete", delete)

	serverAddr := ":8080"
	fmt.Printf("Server is running on http://localhost%s/entry\n", serverAddr)

	err = http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
