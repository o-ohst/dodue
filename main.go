package main

import (
	"net/http"
	"os"
)

//**************************************** Main **************************************
func main() {

	//DEV
	// initLocalEnv()

	//init
	initDb()
	defer db.Close()
	mux := http.NewServeMux()

	//handlers
	getTasksHandler := http.HandlerFunc(getTasks)
	newTaskHandler := http.HandlerFunc(newTask)
	doneTaskHandler := http.HandlerFunc(doneTask)
	getCategoriesHandler := http.HandlerFunc(getCategories)
	newCategoryHandler := http.HandlerFunc(newCategory)
	deleteCategoryHandler := http.HandlerFunc(removeCategory)
	signupHandler := http.HandlerFunc(signup)
	loginHandler := http.HandlerFunc(login)
	deleteDoneTasksHandler := http.HandlerFunc(removeDoneTasks)

	//endpoints
	mux.Handle("/tasks", cors(authApi(authJwt(checkUserCookie(getTasksHandler)))))                   //GET
	mux.Handle("/tasks/new", cors(authApi(authJwt(checkUserCookie(newTaskHandler)))))                //POST
	mux.Handle("/tasks/done", cors(authApi(authJwt(checkUserCookie(doneTaskHandler)))))              //PUT
	mux.Handle("/tasks/deletedone", cors(authApi(authJwt(checkUserCookie(deleteDoneTasksHandler))))) //DELETE

	mux.Handle("/categories", cors(authApi(authJwt(checkUserCookie(getCategoriesHandler)))))         //GET
	mux.Handle("/categories/new", cors(authApi(authJwt(checkUserCookie(newCategoryHandler)))))       //POST
	mux.Handle("/categories/delete", cors(authApi(authJwt(checkUserCookie(deleteCategoryHandler))))) //DELETE

	mux.Handle("/signup", cors(authApi(signupHandler))) //POST
	mux.Handle("/login", cors(authApi(loginHandler)))   //POST

	//serve
	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, mux)
}
