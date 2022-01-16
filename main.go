package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v4"
	// "github.com/joho/godotenv" //DEV
	"golang.org/x/crypto/bcrypt"
)

// func initLocalEnv() {
// 	err := godotenv.Load(".env")
// 	if err != nil {
// 		log.Fatalf("Error loading .env file")
// 	}
// 	return
// }

func handleFatalError(msg string, err error) {
	if err != nil {
		log.Fatal(fmt.Errorf(msg+"; %w", err))
	}
}

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	handleFatalError("hashPassword", err)
	return string(bytes)
}

func checkPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	handleFatalError("checkPasswordHash", err)

	return err
}

//************************************ Struct ***************************************

type Error struct {
	Error string `json:"error"`
}

type Task struct {
	Task_id     int    `json:"task_id"`
	Name        string `json:"name"`
	Category_id int    `json:"category_id"`
	Done        bool   `json:"done"`
	User_id     int    `json:"user_id"`
}

type Category struct {
	Category_id int    `json:"category_id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
}

type User struct {
	User_id  int    `json:"user_id"`
	Username string `json:"username"`
	Password string `josn:"password"`
}

//************************************ Database ****************************************

var db *pgx.Conn

func initDb() {
	var err error
	db, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	handleFatalError("initDb", err)
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

func deleteTask(task_id int) error {
	_, err := db.Exec(context.Background(), "delete from widgets where id=$1", task_id)

	return err
}

//************************************ Category CRUD ****************************************

func createCategory(name string, color int) error {
	_, err := db.Exec(context.Background(), "insert into categories(name, color) values($1, $2)", name, color)

	return err
}

func readCategories(user_id int) (pgx.Rows, error) {
	rows, err := db.Query(context.Background(), "select * from tasks where user_id=", user_id)

	return rows, err
}

//************************************ User CRUD ****************************************

func createUser(username string, password string) error {

	_, err := db.Exec(context.Background(), "insert into users(username, password) values($1, $2)", username, hashPassword(password))

	return err
}

func readUser(username string) (pgx.Rows, error) {

	rows, err := db.Query(context.Background(), "select * from users where username=$1", username)

	return rows, err
}

//******************************************* Middlewares ***********************************

func authApi(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("api-key")
		if apiKey != os.Getenv("API_SECRET") {
			w.WriteHeader(403)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func authJwt(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Token"] != nil {

			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf(("Invalid Signing Method"))
				}
				iss := "dodue"
				checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
				if !checkIss {
					return nil, fmt.Errorf(("invalid iss"))
				}
				return os.Getenv("JWT_SECRET"), nil

			})

			if err != nil {
				fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				next.ServeHTTP(w, r)
			}

		} else {
			fmt.Fprintf(w, "No Authorization Token provided")
		}
	})
}

//JWT

func GetJWT(username string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = username
	claims["iss"] = "dodue"

	tokenString, err := token.SignedString(os.Getenv("JWT_SECRET"))
	handleFatalError("JWT", err)

	return tokenString
}

//******************************************** Routes *************************************

func getTasks(w http.ResponseWriter, r *http.Request) { //req: user_id

	var tasks []Task

	user_id, _ := strconv.Atoi(r.Header.Get("user_id"))
	rows, err := readTasks(user_id)

	if err != nil {
		e := Error{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.Task_id, &task.Name, &task.Category_id, &task.Done, &task.User_id)
		handleFatalError("getTasks scan", err)
		tasks = append(tasks, task)
	}

	json.NewEncoder(w).Encode(tasks)

}

func getCategories(w http.ResponseWriter, r *http.Request) { //req: user_id

	var categories []Category

	user_id, _ := strconv.Atoi(r.Header.Get("user_id"))
	rows, err := readCategories(user_id)

	if err != nil {
		e := Error{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var category Category
		err := rows.Scan(&category.Category_id, &category.Name, &category.Color)
		handleFatalError("getCategories scan", err)
		categories = append(categories, category)
	}

	json.NewEncoder(w).Encode(categories)

}

func newTask(w http.ResponseWriter, r *http.Request) {

	user_id, _ := strconv.Atoi(r.Header.Get("user_id"))
	decoder := json.NewDecoder(r.Body)
	var task Task
	err := decoder.Decode(&task)
	handleFatalError("newTask", err)
	err2 := createTask(task.Name, user_id, task.Category_id)
	handleFatalError("newTask createTask", err2)

}

func newCategory(w http.ResponseWriter, r *http.Request) {

	user_id, _ := strconv.Atoi(r.Header.Get("user_id"))
	decoder := json.NewDecoder(r.Body)
	var category Category
	err := decoder.Decode(&category)
	handleFatalError("postCategory", err)
	err2 := createTask(category.Name, user_id, category.Category_id)
	handleFatalError("newCategory newCategory", err2)

}

func login(w http.ResponseWriter, r *http.Request) { //req: username, password

	username := r.Header.Get("username")
	password := r.Header.Get("password")
	log.Print("logging in")
	rows, err := readUser(username)
	log.Print("finished readUser")

	if err != nil {
		e := Error{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	defer rows.Close()

	var user User
	err2 := rows.Scan(&user.User_id, &user.Username, &user.Password)
	handleFatalError("login scan", err2)
	log.Print("finished scan")

	err3 := checkPasswordHash(password, user.Password)
	log.Print("finished check hash")

	if err3 != nil {
		e := Error{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(e)
		return
	} else {

		token := GetJWT(username)
		log.Print("got jwt")
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
			Secure:   true,
		})
		log.Print("set cookie")
		http.SetCookie(w, &http.Cookie{
			Name:  "username",
			Value: user.Username,
		})
	}
}

func signup(w http.ResponseWriter, r *http.Request) { //req: username, password

	username := r.Header.Get("username")
	password := r.Header.Get("password")
	err := createUser(username, password)

	if err != nil {
		e := Error{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(e)
		return
	}

}

//**************************************** Main **************************************
func main() {

	//DEV
	// initLocalEnv()

	initDb()
	defer db.Close(context.Background())

	//************************* Routes *************************
	mux := http.NewServeMux()

	tasksHandler := http.HandlerFunc(getTasks)
	categoriesHandler := http.HandlerFunc(getCategories)
	signupHandler := http.HandlerFunc(signup)
	loginHandler := http.HandlerFunc(login)

	mux.Handle("/tasks", authJwt(authApi(tasksHandler)))
	mux.Handle("/categories", authJwt(authApi(categoriesHandler)))
	mux.Handle("/signup", authApi(signupHandler))
	mux.Handle("/login", authApi(loginHandler))
	
	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, mux)

}
