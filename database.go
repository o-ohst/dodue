package main

import (	
	"log"
	"os"
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool")

var db *pgxpool.Pool

func initDb() {
	var err error
	db, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	logError("initDb", err)
}

//************************************ Task CRUD ****************************************

func createTask(name string, user_id int, category_id int) error {
	_, err := db.Exec(context.Background(), "insert into tasks(name, user_id, category_id) values($1, $2, $3)", name, user_id, category_id)
	return err
}

func readTasks(user_id int) (pgx.Rows, error) {
	rows, err := db.Query(context.Background(), "select * from tasks where user_id=$1 and done=$2", user_id, false)

	return rows, err
}

func updateTaskDone(task_id int, done bool) error {
	log.Print(done)
	_, err := db.Exec(context.Background(), "update tasks set done=$1 where task_id=$2", done, task_id)

	return err
}

// func deleteTask(task_id int) error { //unused for now
// 	_, err := db.Exec(context.Background(), "delete from tasks where task_id=$1", task_id)

// 	return err
// }

func deleteDoneTasks(user_id int) error {
	_, err := db.Exec(context.Background(), "delete from tasks where done=$1 and user_id=$2", true, user_id)

	return err
}

//************************************ Category CRUD ****************************************

func createCategory(name string, color int, user_id int) error {
	_, err := db.Exec(context.Background(), "insert into categories(name, color, user_id) values($1, $2, $3)", name, color, user_id)

	return err
}

func readCategories(user_id int) (pgx.Rows, error) {
	rows, err := db.Query(context.Background(), "select * from categories where user_id=$1", user_id)

	return rows, err
}

func deleteCategory(category_id int) error {
	_, err := db.Exec(context.Background(), "delete from categories where category_id=$1", category_id)

	return err
}

//************************************ User CRUD ****************************************

func createUser(username string, password string) error {

	_, err := db.Exec(context.Background(), "insert into users(username, password) values($1, $2)", username, hash(password))

	return err
}

func readUser(username string) (pgx.Rows, error) {
	log.Print(username)

	rows, err := db.Query(context.Background(), "select * from users where username=$1", username)

	return rows, err
}
