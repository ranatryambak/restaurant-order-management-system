package database

import (
	"database/sql"
	"fmt"
	"log"
)

func ConnectDb(dbname string) (*sql.DB,error){

	connectionString := fmt.Sprintf("user=pqgotest dbname=%s sslmode=verify-full",dbname)
	db,err := sql.Open("postgres",connectionString)
	if err!=nil{
		log.Fatal(err)
	}

	return db,nil
}