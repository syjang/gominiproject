package main

import (
	"database/sql"
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"
	_ "gopkg.in/goracle.v2"
)

type DatabaseInfo struct {
	DBUserPW string `json:"dbuserpw"`
	DBName   string `json:"dbname"`
}

func initDB(dbInfo DatabaseInfo) error {
	log.Println("Initializing Database....")
	db, err := dbConn(dbInfo)
	if err != nil {
		log.Println("Can't not connect to DB....")
		log.Println("DB Info : ", dbInfo.DBUserPW, dbInfo.DBName)
		//panic(err.Error())
		return err
	}
	defer db.Close()

	// check open
	err = db.Ping()
	if err != nil {
		log.Println("Ping Test Error....")
		log.Println("DB Info : ", dbInfo.DBUserPW, dbInfo.DBName)
		//panic(err.Error())
		return err
	}
	log.Println("Database connected.")
	return nil
}

func dbConn(dbInfo DatabaseInfo) (*sql.DB, error) {
	dbDriver := "goracle" // oracle
	// dbUser := "root"
	// dbPass := "face"
	// dbName := "tcp(172.17.0.2:3306)/face"
	db, err := sql.Open(dbDriver, dbInfo.DBUserPW+"@"+dbInfo.DBName)
	if err != nil {
		panic(err.Error())
	}
	return db, err
}

func main() {
	dbinfo := DatabaseInfo{
		DBName:   "127.0.0.1:1521/ORCLCDB.localdomain",
		DBUserPW: "face/face",
	}
	initDB(dbinfo)

	db, err := dbConn(dbinfo)
	if err != nil {
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Print(err)
		return
	}

	u1 := uuid.Must(uuid.NewV4())
	test := make([]byte, 512)
	for i := 0; i < 512; i++ {
		test[i] = (byte)(i)
	}

	temp := fmt.Sprintf("INSERT INTO tbl_face_feature (UUID, NAME ,FEATURE) VALUES ('%s' , '%s' , hextoraw('%x'))", u1, "Human", test)
	// log.Println(temp)
	_, err = db.Exec(temp)
	if err != nil {
		log.Println(err)
		return
	}

	seldb, err := db.Query("SELECT uuid ,name ,feature FROM tbl_face_feature")
	if err != nil {
		log.Println(err)
		return
	}

	for seldb.Next() {
		var id string
		var name string
		var feature []byte
		err = seldb.Scan(&id, &name, &feature)
		if err != nil {
			log.Println(err)
		}

		log.Println(id)
		log.Println(name)
		log.Println(len(feature))
	}

}
