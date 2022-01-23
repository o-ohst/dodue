package main

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

type NewCategory struct {
	Name        string `json:"name"`
	Color       string    `json:"color"`
}

type NewTask struct {
	Name        string `json:"name"`
	Category_id int   `json:"category_id"`
}
