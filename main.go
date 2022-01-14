package main

import (
	"context"
	_ "database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	_ "strconv"

	"github.com/jackc/pgx/v4"
	// "github.com/joho/godotenv"
)

//*****************************DATABASE*****************************************
type Task struct {
	Name        string `json:"name"`
	Done        bool   `json:"done"`
	Task_id     int    `json:"task_id"`
	Category_id int    `json:"category_id"`
}

type Category struct {
	Category_id int `json:"category_id"`
	Name string `json:"name"`
	Color int `json:"color"`
}

var tasks []Task
var categories []Category

var db *pgx.Conn

func addTask(name string) error {
	_, err := db.Exec(context.Background(), "insert into tasks(name, task_id) values($1, $2)", name, 1)
	return err
}

func listTasks() error {
	rows, _ := db.Query(context.Background(), "select * from tasks")

	for rows.Next() {
		var id int32
		var description string
		err := rows.Scan(&id, &description)
		if err != nil {
			return err
		}
		fmt.Printf("%d. %s\n", id, description)
	}

	return rows.Err()
}

//******************************************** Routes *************************************

func handleTasks(w http.ResponseWriter, r *http.Request) {

	tasks = []Task{
		Task{Name: "task 1", Done: false, Task_id: 1, Category_id: 1},
		Task{Name: "task 2", Done: false, Task_id: 2, Category_id: 1},
		Task{Name: "task 3", Done: false, Task_id: 3, Category_id: 2},
	}
	json.NewEncoder(w).Encode(tasks)

}

func handleCategories(w http.ResponseWriter, r *http.Request) {

	categories = []Category{
		Category{Name: "category 1", Category_id: 1, Color: 0},
		Category{Name: "category 2", Category_id: 2, Color: 1},
	}
	json.NewEncoder(w).Encode(categories)

}

//**************************************** Main **************************************
func main() {

	// envErr := godotenv.Load(".env")
	// if envErr != nil {
	// 	log.Fatalf("Error loading .env file")
	// }

	var err error
	db, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close(context.Background())

	// fmt.Println(addTask("hello"))
	fmt.Println(listTasks())

	//**********ROUTES**********
	http.HandleFunc("/tasks", handleTasks)
	http.HandleFunc("/categories", handleCategories)

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, nil)

}
