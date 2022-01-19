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

	"github.com/joho/godotenv" //DEV
	"golang.org/x/crypto/bcrypt"
)

func initLocalEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func writeErrorToResponse(w http.ResponseWriter, err error) {
	if err != nil {
		e := Error{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(e)
	}
}

func writeErrorMessageToResponse(w http.ResponseWriter, errorMessage string) {
	e := Error{
		Error: errorMessage,
	}
	json.NewEncoder(w).Encode(e)
}

func logError(msg string, err error) {
	if err != nil {
		log.Print(fmt.Errorf(msg+"; %w", err))
	}
}

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	logError("hashPassword", err)
	return string(bytes)
}

func checkPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	logError("checkPasswordHash", err)

	return err
}

func GetJWT(username string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["username"] = username
	claims["iss"] = "dodue"

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	logError("JWT", err)

	return tokenString
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
	User_id     int    `json:"user_id"`
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
	_, err := db.Exec(context.Background(), "update tasks set done=$1 where task_id=$2", done, task_id)

	return err
}

func deleteTask(task_id int) error {
	_, err := db.Exec(context.Background(), "delete from widgets where id=$1", task_id)

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

//************************************ User CRUD ****************************************

func createUser(username string, password string) error {

	_, err := db.Exec(context.Background(), "insert into users(username, password) values($1, $2)", username, hashPassword(password))

	return err
}

func readUser(username string) (pgx.Rows, error) {
	log.Print(username)

	rows, err := db.Query(context.Background(), "select * from users where username=$1", username)

	return rows, err
}

//******************************************* Middlewares ***********************************

func authApi(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("api_key")

		if apiKey == "" {
			log.Print("no api key provided")
			w.WriteHeader(403)
			return
		}

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

		tokenCookie, err := r.Cookie("token")

		if err == nil {

			token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf(("invalid signing method"))
				}

				iss := "dodue"
				checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, true)
				if !checkIss {
					return nil, fmt.Errorf(("invalid iss"))
				}

				return []byte(os.Getenv("JWT_SECRET")), nil

			})

			if err != nil {
				logError("authJwt Parse", err)
			}

			if token.Valid {
				next.ServeHTTP(w, r)
			}

		} else {
			writeErrorMessageToResponse(w, "no authorization token provided")
			return
		}
	})
}

func checkUserCookie(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("user_id")
		if err != nil {
			writeErrorMessageToResponse(w, "no user_id cookie provided")
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func cors(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
    	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, token, user_id, api_key")
		next.ServeHTTP(w, r)
	})
}


 

//******************************************** Endpoints *************************************

func getTasks(w http.ResponseWriter, r *http.Request) { //req: user_id

	if r.Method != "GET" {
		w.WriteHeader(404)
		return
	}

	var tasks []Task
	var user_id int

	userCookie, _ := r.Cookie("user_id")
	user_id, _ = strconv.Atoi(userCookie.Value)

	rows, err := readTasks(user_id)

	if err != nil {
		writeErrorToResponse(w, err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.Task_id, &task.Name, &task.Category_id, &task.Done, &task.User_id)
		if err != nil {
			logError("getTasks scan", err)
			writeErrorToResponse(w, err)
			return
		}
		tasks = append(tasks, task)
	}

	json.NewEncoder(w).Encode(tasks)

}

func getCategories(w http.ResponseWriter, r *http.Request) { //req: user_id

	if r.Method != "GET" {
		w.WriteHeader(404)
		return
	}

	var categories []Category

	userCookie, _ := r.Cookie("user_id")
	user_id, _ := strconv.Atoi(userCookie.Value)
	rows, err := readCategories(user_id)

	if err != nil {
		writeErrorToResponse(w, err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var category Category
		err := rows.Scan(&category.Category_id, &category.Name, &category.Color)
		if err != nil {
			logError("getCategories scan", err)
			writeErrorToResponse(w, err)
			return
		}
		categories = append(categories, category)
	}

	json.NewEncoder(w).Encode(categories)

}

func newTask(w http.ResponseWriter, r *http.Request) { //req: user_id, name

	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}

	userCookie, _ := r.Cookie("user_id")
	user_id, _ := strconv.Atoi(userCookie.Value)
	decoder := json.NewDecoder(r.Body)
	var task Task
	err := decoder.Decode(&task)
	logError("newTask", err)

	if task.Name == "" {
		log.Print("empty name")
		writeErrorMessageToResponse(w, "no name provided")
		return
	}

	if task.Category_id == 0 {
		log.Print("empty category id")
		writeErrorMessageToResponse(w, "no category_id provided")
		return
	}

	err2 := createTask(task.Name, user_id, task.Category_id)
	logError("newTask createTask", err2)
	if err2 != nil {
		writeErrorToResponse(w, err2)
		return
	}

}

func doneTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != "PUT" {
		w.WriteHeader(404)
		return
	}

	task_id, _ := strconv.Atoi(r.Header.Get("task_id"))

	err := updateTaskDone(task_id, true)
	if err != nil {
		logError("doneTask updateTaskDone", err)
		writeErrorToResponse(w, err)
		return
	}

}

func newCategory(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}

	userCookie, _ := r.Cookie("user_id")
	user_id, _ := strconv.Atoi(userCookie.Value)
	decoder := json.NewDecoder(r.Body)
	var category Category
	err := decoder.Decode(&category)
	logError("postCategory", err)

	if category.Name == "" {
		log.Print("empty name")
		writeErrorMessageToResponse(w, "no name provided")
		return
	}

	// if category.Color == 0 {
	// 	log.Print("empty category id")
	// 	writeErrorMessageToResponse(w, "no category_id provided")
	// 	return
	// }

	err2 := createCategory(category.Name, category.Color, user_id)
	if err2 != nil {
		logError("newCategory newCategory", err2)
		writeErrorToResponse(w, err2)
		return
	}

}

func login(w http.ResponseWriter, r *http.Request) { //req: username, password

	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}

	username := r.Header.Get("username")
	password := r.Header.Get("password")

	if username == "" {
		writeErrorMessageToResponse(w, "no username provided")
		return
	}

	if password == "" {
		writeErrorMessageToResponse(w, "no password provided")
		return
	}

	rows, err := readUser(username)
	if err != nil {
		logError("login readUser", err)
		writeErrorToResponse(w, err)
		return
	}
	defer rows.Close()
	if !rows.Next() {
		log.Print("login readUser; user not found")
		writeErrorMessageToResponse(w, "user not found")
		return
	}

	var user User
	err2 := rows.Scan(&user.User_id, &user.Username, &user.Password)
	if err2 != nil {
		logError("login scan", err2)
		return
	}

	err3 := checkPasswordHash(password, user.Password)
	if err3 != nil {

		logError("login checkPasswordHash", err3)
		writeErrorMessageToResponse(w, "wrong password")
		return

	} else { //login success!

		token := GetJWT(username)
		log.Print("got jwt")

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
			Secure:   true, //DEV
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "user_id",
			Value: strconv.Itoa(user.User_id),
		})

	}
}

func signup(w http.ResponseWriter, r *http.Request) { //req: username, password

	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}

	username := r.Header.Get("username")
	password := r.Header.Get("password")

	if username == "" {
		writeErrorMessageToResponse(w, "no username provided")
		return
	}

	if password == "" {
		writeErrorMessageToResponse(w, "no password provided")
		return
	}

	err := createUser(username, password)

	if err != nil {
		logError("signup createUser", err)
		writeErrorToResponse(w, err)
		return
	}

}

//**************************************** Main **************************************
func main() {

	//DEV
	// initLocalEnv()

	initDb()
	defer db.Close(context.Background())

	//************************* Endpints *************************
	mux := http.NewServeMux()

	getTasksHandler := http.HandlerFunc(getTasks)
	newTaskHandler := http.HandlerFunc(newTask)
	doneTaskHandler := http.HandlerFunc(doneTask)
	getCategoriesHandler := http.HandlerFunc(getCategories)
	newCategoryHandler := http.HandlerFunc(newCategory)
	signupHandler := http.HandlerFunc(signup)
	loginHandler := http.HandlerFunc(login)

	mux.Handle("/tasks", cors(authApi(authJwt(checkUserCookie(getTasksHandler)))))      //GET
	mux.Handle("/tasks/new", cors(authApi(authJwt(checkUserCookie(newTaskHandler)))))  //POST
	mux.Handle("/tasks/done", cors(authApi(authJwt(checkUserCookie(doneTaskHandler))))) //PUT

	mux.Handle("/categories", cors(authApi(authJwt(checkUserCookie(getCategoriesHandler)))))   //GET
	mux.Handle("/categories/new", cors(authApi(authJwt(checkUserCookie(newCategoryHandler))))) //POST

	mux.Handle("/signup", cors(authApi(signupHandler))) //POST
	mux.Handle("/login", cors(authApi(loginHandler)))   //POST

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, mux)

}
