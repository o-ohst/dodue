package main

import(
	"net/http"
	"strconv"
	"log"
	"encoding/json"
	"time"
)

func getTasks(w http.ResponseWriter, r *http.Request) {

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

func getCategories(w http.ResponseWriter, r *http.Request) {

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
		err := rows.Scan(&category.Category_id, &category.Name, &category.Color, &category.User_id)
		if err != nil {
			logError("getCategories scan", err)
			writeErrorToResponse(w, err)
			return
		}
		categories = append(categories, category)
	}

	json.NewEncoder(w).Encode(categories)

}

func newTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}

	userCookie, _ := r.Cookie("user_id")
	user_id, _ := strconv.Atoi(userCookie.Value)
	decoder := json.NewDecoder(r.Body)
	var task NewTask
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
	done, _ := strconv.ParseBool(r.Header.Get("done"))
	
	log.Print(r.Header.Get("done"))
	log.Print(done)

	err := updateTaskDone(task_id, done)
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
	var category NewCategory
	err := decoder.Decode(&category)
	logError("postCategory", err)

	if category.Name == "" {
		log.Print("empty name")
		writeErrorMessageToResponse(w, "no name provided")
		return
	}

	cColor , _ := strconv.Atoi(category.Color)
	err2 := createCategory(category.Name, cColor, user_id)
	if err2 != nil {
		logError("newCategory newCategory", err2)
		writeErrorToResponse(w, err2)
		return
	}
}

func removeCategory(w http.ResponseWriter, r *http.Request) {

	if r.Method != "DELETE" {
		w.WriteHeader(404)
		return
	}

	cId, _ := strconv.Atoi(r.Header.Get("category_id"))
	
	log.Print(r.Header.Get("category_id"))

	err := deleteCategory(cId)
	if err != nil {
		logError("removeCategory deteleCategory", err)
		writeErrorToResponse(w, err)
		return
	}
}

func removeDoneTasks(w http.ResponseWriter, r *http.Request) {

	if r.Method != "DELETE" {
		w.WriteHeader(404)
		return
	}

	userCookie, _ := r.Cookie("user_id")
	user_id, _ := strconv.Atoi(userCookie.Value)

	err := deleteDoneTasks(user_id)
	if err != nil {
		logError("removeDoneTasks deleteDoneTasks", err)
		writeErrorToResponse(w, err)
		return
	}

}

func login(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}

	username := r.Header.Get("username")
	password := r.Header.Get("password")

	log.Print(username)
	log.Print(password)

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

	err3 := checkPassword(password, user.Password)
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
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			Path: "/",
			Expires: time.Now().AddDate(0, 6, 0),
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "user_id",
			Value: strconv.Itoa(user.User_id),
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			Path: "/",
			Expires: time.Now().AddDate(0, 6, 0),
		})

	}
}

func signup(w http.ResponseWriter, r *http.Request) {

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