package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var mySigningKey = []byte("privia")

//print my signing key

type Credentials struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `json:"username"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var users = []Credentials{
	{ID: 1, Username: "user1", Password: "password1", Type: "1"},
	{ID: 2, Username: "user2", Password: "password2", Type: "2"},
}

func login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid := false
	for _, user := range users {
		if creds.Username == user.Username && creds.Password == user.Password {
			valid = true
			break
		}
	}

	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(tokenString))
}

func isAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return mySigningKey, nil
		})
		if err != nil || !token.Valid || token == nil || tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

type ToDoList struct {
	ID              uint       `gorm:"primaryKey"`
	UserID          uint       `json:"user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       time.Time  `json:"deleted_at"`
	Title           string     `json:"title"`
	CompletePercent int        `json:"percent"`
	Items           []ToDoItem `json:"items"`
	Deleted         bool       `json:"deleted"`
}

type ToDoItem struct {
	ID         uint      `gorm:"primaryKey"`
	ToDoListID uint      `json:"todo_list_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  time.Time `json:"deleted_at"`
	Task       string    `json:"task"`
	Completed  bool      `json:"completed"`
	Deleted    bool      `json:"deleted"`
}

var ToDoLists = []ToDoList{
	{ID: 1, UserID: 1, Title: "ToDoList 1", CompletePercent: 0, Items: []ToDoItem{{ID: 1, ToDoListID: 1, Task: "Task 1", Completed: false, Deleted: false}, {ID: 2, ToDoListID: 1, Task: "Task 2", Completed: false, Deleted: true}}, Deleted: false},
}

func listDetailFunc(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	claims := token.Claims.(jwt.MapClaims)

	Username := claims["username"].(string)
	var userType string
	var userid uint
	for _, user := range users {
		if user.Username == Username {
			userType = user.Type
			userid = user.ID
			break
		}
	}

	if userType == "1" {
		for _, list := range ToDoLists {
			if list.UserID == userid {
				response, _ := json.Marshal(list)
				w.Write(response)
			}
		}
	} else if userType == "2" {
		for _, list := range ToDoLists {
			response, _ := json.Marshal(list)
			w.Write(response)
		}
	} else {
		// Geçersiz kullanıcı türü durumunda hata mesajı döndürülür.
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid user type"))
		return
	}
}
func createListFunc(w http.ResponseWriter, r *http.Request) {
	var list ToDoList
	err := json.NewDecoder(r.Body).Decode(&list)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ToDoLists = append(ToDoLists, list)
	w.Write([]byte("List created successfully"))
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	// If url not have a / then return error
	if r.URL.Path == "/list" {
		listDetailFunc(w, r)
		return
	}

	listID := r.URL.Path[len("/list/"):]
	if listID == "" {
		listDetailFunc(w, r)
		return
	}
	// If url /list/create then return error
	if listID == "create" {
		createListFunc(w, r)
		return
	}

	// If url /list/{listID}/delete then delete list with ID {listID}
	// After list id
	if strings.HasSuffix(listID, "/delete") {
		listID = strings.TrimSuffix(listID, "/delete")
		id, err := strconv.Atoi(listID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for i, list := range ToDoLists {
			if list.ID == uint(id) {
				//Update the deleted field of the list to true
				ToDoLists[i].Deleted = true
				ToDoLists[i].DeletedAt = time.Now()
				//Update the deleted field of the items to true
				for j := range ToDoLists[i].Items {
					ToDoLists[i].Items[j].Deleted = true
					ToDoLists[i].Items[j].DeletedAt = time.Now()
				}
				w.Write([]byte("List deleted successfully"))
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// If url /list/{listID} then return list details
	id, err := strconv.Atoi(listID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, list := range ToDoLists {
		if list.ID == uint(id) {
			// If userType is 1 then return only user1's list if userType is 2 then return all lists
			tokenString := r.Header.Get("Authorization")
			token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return mySigningKey, nil
			})
			claims := token.Claims.(jwt.MapClaims)

			Username := claims["username"].(string)
			var userType string
			var userid uint
			for _, user := range users {
				if user.Username == Username {
					userType = user.Type
					userid = user.ID
					break
				}
			}
			if userType == "1" {
				for _, list := range ToDoLists {
					if list.UserID == userid {
						if list.ID == uint(id) {
							response, _ := json.Marshal(list)
							w.Write(response)
							return
						}
					}
				}
			} else if userType == "2" {
				for _, list := range ToDoLists {
					if list.ID == uint(id) {
						response, _ := json.Marshal(list)
						w.Write(response)
						return
					}
				}
			} else {
				// Geçersiz kullanıcı türü durumunda hata mesajı döndürülür.
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid user type"))
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
	return

}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Creating a new task..."))
}

func completeTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Completing task..."))
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Updating task..."))
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Deleting task..."))
}

func handleRequests() {

	http.HandleFunc("/login", login)
	//
	http.HandleFunc("/list/", isAuthenticated(listHandler))
	//
	http.HandleFunc("/list/task/create", isAuthenticated(createTaskHandler))
	http.HandleFunc("/list/task/update", isAuthenticated(updateTaskHandler))
	http.HandleFunc("/list/task/complete", isAuthenticated(completeTaskHandler))
	http.HandleFunc("/list/task/delete", isAuthenticated(deleteTaskHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
