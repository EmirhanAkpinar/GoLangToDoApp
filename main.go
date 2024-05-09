package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var mySigningKey = []byte("privia")

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
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
	Task      string    `json:"task"`
	Completed bool      `json:"completed"`
	Deleted   bool      `json:"deleted"`
}

var ToDoLists = []ToDoList{
	{ID: 1, UserID: 1, Title: "ToDoList 1", CompletePercent: 0, Items: []ToDoItem{{ID: 1, Task: "Task 1", Completed: false, Deleted: false}, {ID: 2, Task: "Task 2", Completed: false, Deleted: true}}, Deleted: false},
	{ID: 2, UserID: 1, Title: "ToDoList 2", CompletePercent: 0, Items: []ToDoItem{{ID: 1, Task: "Task 1", Completed: false, Deleted: false}, {ID: 2, Task: "Task 2", Completed: false, Deleted: true}}, Deleted: false},
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
	//Create a new list with the title in the request body
	var list ToDoList
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		w.Write([]byte("Error reading request body"))
		return
	}

	err = json.Unmarshal(body, &list)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error decoding JSON"))

		return
	}
	tokenString := r.Header.Get("Authorization")
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	claims := token.Claims.(jwt.MapClaims)

	Username := claims["username"].(string)
	var userid uint
	for _, user := range users {
		if user.Username == Username {
			userid = user.ID
			break
		}
	}
	list.ID = uint(len(ToDoLists) + 1)
	list.UserID = userid
	list.CreatedAt = time.Now()
	list.UpdatedAt = time.Now()
	list.Deleted = false
	ToDoLists = append(ToDoLists, list)

	w.Write([]byte("List created successfully"))
}

func todoListFunc(w http.ResponseWriter, r *http.Request, listID string) {
	id, err := strconv.Atoi(listID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, list := range ToDoLists {
		if list.ID == uint(id) {
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
}

func deleteListFunc(w http.ResponseWriter, r *http.Request, listID string) {
	listID = strings.TrimSuffix(listID, "/delete")
	id, err := strconv.Atoi(listID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for i, list := range ToDoLists {
		if list.ID == uint(id) {
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
					if list.ID == uint(id) {
						if list.UserID == userid {
							ToDoLists[i].Deleted = true
							ToDoLists[i].DeletedAt = time.Now()
							//Update the deleted field of the items to true
							for j := range ToDoLists[i].Items {
								ToDoLists[i].Items[j].Deleted = true
								ToDoLists[i].Items[j].DeletedAt = time.Now()
							}
							w.Write([]byte("List deleted successfully"))
							return
						} else {
							// 404
							w.WriteHeader(http.StatusNotFound)
							return
						}
					}
				}
			} else if userType == "2" {
				if list.ID == uint(id) {
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
func updateListFunc(w http.ResponseWriter, r *http.Request, listID string) {
	listID = strings.TrimSuffix(listID, "/update")
	id, err := strconv.Atoi(listID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, list := range ToDoLists {
		if list.ID == uint(id) {
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
					if list.ID == uint(id) {
						if list.UserID == userid {
							//Update the task.Task field with the new value in the request body
							var task ToDoList
							body, err := ioutil.ReadAll(r.Body)
							if err != nil {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("Error reading request body"))
								return
							}

							err = json.Unmarshal(body, &task)
							if err != nil {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("Error decoding JSON"))
								return
							}
							list.Title = task.Title
							list.UpdatedAt = time.Now()
							w.Write([]byte("List updated successfully"))
							return
						} else {
							// 404
							w.WriteHeader(http.StatusNotFound)
							return
						}
					}
				}
			} else if userType == "2" {
				if list.ID == uint(id) {
					//Update the task.Task field with the new value in the request body
					var task ToDoList
					body, err := ioutil.ReadAll(r.Body)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("Error reading request body"))
						return
					}

					err = json.Unmarshal(body, &task)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("Error decoding JSON"))
						return
					}
					list.Title = task.Title
					list.UpdatedAt = time.Now()
					w.Write([]byte("List updated successfully"))
					return
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

func taskHandler(w http.ResponseWriter, r *http.Request) {
	listID := strings.TrimPrefix(r.URL.Path, "/list/")
	listIDParts := strings.Split(listID, "/")
	if len(listIDParts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// I messed up the listID and id variables. I will fix it later.
	id, err := strconv.Atoi(listIDParts[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	taskIDParts := strings.Split(r.URL.Path, "/")
	if len(taskIDParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	taskID := strings.Join(taskIDParts[4:], "/")
	if strings.HasSuffix(listID, "/create") {
		taskCreateFunc(w, r, id, err)
		return

	}

	taskID2 := taskIDParts[4]
	taskid, err := strconv.Atoi(taskID2)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if strings.HasSuffix(taskID, "/delete") {
		taskDeleteFunc(w, r, taskID, taskid, id)
		return
	}

	// If url /list/{listID}/task/{taskid}/complete then complete task with ID {taskid}
	if strings.HasSuffix(taskID, "/complete") {
		taskCompleteFunc(w, r, taskID, taskid, id)
		return
	}
	// If url /list/{listID}/task/{taskid}/update then update task with ID {taskid}
	if strings.HasSuffix(taskID, "/update") {
		taskUpdateFunc(w, r, taskID, taskid, id)
		return
	}

	// If url /list/{listID}/task/create then create a new task in the list with ID {listID}

	// If url /list/{listID}/task/{taskid} then return task details
	taskDetailFunc(w, r, taskID, taskid, id)
	w.WriteHeader(http.StatusNotFound)
	return
}

func taskDetailFunc(w http.ResponseWriter, r *http.Request, taskID string, taskid int, id int) {
	for _, list := range ToDoLists {
		if list.ID == uint(id) {
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
							for _, item := range list.Items {
								if item.ID == uint(taskid) {
									response, _ := json.Marshal(item)
									w.Write(response)
									return
								}
							}
						}
					}

				}
			} else if userType == "2" {
				for _, list := range ToDoLists {
					if list.ID == uint(id) {
						for _, item := range list.Items {
							if item.ID == uint(taskid) {
								response, _ := json.Marshal(item)
								w.Write(response)
								return
							}
						}
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
}

func taskUpdateFunc(w http.ResponseWriter, r *http.Request, taskID string, taskid int, id int) {
	taskID = strings.TrimSuffix(taskID, "/update")
	taskid, err := strconv.Atoi(taskID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for i, list := range ToDoLists {
		if list.ID == uint(id) {
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
							for j, item := range list.Items {
								if item.ID == uint(taskid) {
									//Update the task.Task field with the new value in the request body
									var task ToDoItem
									body, err := ioutil.ReadAll(r.Body)
									if err != nil {
										w.WriteHeader(http.StatusBadRequest)
										w.Write([]byte("Error reading request body"))
										return
									}

									err = json.Unmarshal(body, &task)
									if err != nil {
										w.WriteHeader(http.StatusBadRequest)
										w.Write([]byte("Error decoding JSON"))
										return
									}
									ToDoLists[i].Items[j].Task = task.Task
									ToDoLists[i].Items[j].UpdatedAt = time.Now()
									ToDoLists[i].UpdatedAt = time.Now()
									w.Write([]byte("Task updated successfully"))
									return
								}
							}
						}
					}
				}
			} else if userType == "2" {
				for _, list := range ToDoLists {
					if list.ID == uint(id) {
						for j, item := range list.Items {
							if item.ID == uint(taskid) {
								//Update the task.Task field with the new value in the request body
								var task ToDoItem
								body, err := ioutil.ReadAll(r.Body)
								if err != nil {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("Error reading request body"))
									return
								}

								err = json.Unmarshal(body, &task)
								if err != nil {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("Error decoding JSON"))
									return
								}
								ToDoLists[i].Items[j].Task = task.Task
								ToDoLists[i].Items[j].UpdatedAt = time.Now()
								ToDoLists[i].UpdatedAt = time.Now()
								w.Write([]byte("Task updated successfully"))
								return
							}
						}
					}
				}
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
	return
}

func taskCompleteFunc(w http.ResponseWriter, r *http.Request, taskID string, taskid int, id int) {
	taskID = strings.TrimSuffix(taskID, "/complete")
	taskid, err := strconv.Atoi(taskID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for i, list := range ToDoLists {
		if list.ID == uint(id) {
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
							for j, item := range list.Items {
								if item.ID == uint(taskid) {
									ToDoLists[i].Items[j].Completed = true
									ToDoLists[i].Items[j].UpdatedAt = time.Now()
									ToDoLists[i].UpdatedAt = time.Now()
									//Calculate the percentage of completed tasks
									var completedCount int
									for _, item := range list.Items {
										if item.Completed {
											completedCount++
										}
									}
									ToDoLists[i].CompletePercent = (completedCount * 100) / len(list.Items)

									w.Write([]byte("Task completed successfully"))
									return
								}

							}
						}
					}

				}
			} else if userType == "2" {
				for _, list := range ToDoLists {
					if list.ID == uint(id) {
						for j, item := range list.Items {
							if item.ID == uint(taskid) {
								ToDoLists[i].Items[j].Completed = true
								ToDoLists[i].Items[j].UpdatedAt = time.Now()
								ToDoLists[i].UpdatedAt = time.Now()
								//Calculate the percentage of completed tasks
								var completedCount int
								for _, item := range list.Items {
									if item.Completed {
										completedCount++
									}
								}
								ToDoLists[i].CompletePercent = (completedCount * 100) / len(list.Items)
								w.Write([]byte("Task completed successfully"))
								return
							}
						}
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
func taskDeleteFunc(w http.ResponseWriter, r *http.Request, taskID string, taskid int, id int) {
	taskID = strings.TrimSuffix(taskID, "/delete")
	taskid, err := strconv.Atoi(taskID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for i, list := range ToDoLists {
		if list.ID == uint(id) {
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
							for j, item := range list.Items {
								if item.ID == uint(taskid) {
									ToDoLists[i].Items[j].Deleted = true

									ToDoLists[i].Items[j].DeletedAt = time.Now()
									w.Write([]byte("Task deleted successfully"))
									return
								}
							}
						}
					}
				}
			} else if userType == "2" {
				for _, list := range ToDoLists {
					if list.ID == uint(id) {

						for j, item := range list.Items {
							if item.ID == uint(taskid) {
								ToDoLists[i].Items[j].Deleted = true
								ToDoLists[i].Items[j].DeletedAt = time.Now()
								w.Write([]byte("Task deleted successfully"))
								return
							}
						}
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
func taskCreateFunc(w http.ResponseWriter, r *http.Request, id int, err error) {
	//Declare the taskID variable
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for i, list := range ToDoLists {
		if list.ID == uint(id) {
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
							//Create a new task with the title in request body
							var task ToDoItem
							body, err := ioutil.ReadAll(r.Body)
							if err != nil {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("Error reading request body"))
								return
							}

							err = json.Unmarshal(body, &task)
							if err != nil {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("Error decoding JSON"))
								return
							}
							task.ID = uint(len(list.Items) + 1)
							task.CreatedAt = time.Now()
							task.UpdatedAt = time.Now()
							task.Deleted = false
							ToDoLists[i].Items = append(ToDoLists[i].Items, task)
							ToDoLists[i].UpdatedAt = time.Now()
							w.Write([]byte("Task created successfully"))
							return
						}
					}
				}
			} else if userType == "2" {
				for _, list := range ToDoLists {
					if list.ID == uint(id) {
						//Create a new task with the title in request body
						var task ToDoItem
						body, err := ioutil.ReadAll(r.Body)
						if err != nil {
							w.WriteHeader(http.StatusBadRequest)
							w.Write([]byte("Error reading request body"))
							return
						}

						err = json.Unmarshal(body, &task)
						if err != nil {
							w.WriteHeader(http.StatusBadRequest)
							w.Write([]byte("Error decoding JSON"))
							return
						}
						task.ID = uint(len(list.Items) + 1)
						task.CreatedAt = time.Now()
						task.UpdatedAt = time.Now()
						task.Deleted = false

						ToDoLists[i].Items = append(ToDoLists[i].Items, task)

						ToDoLists[i].UpdatedAt = time.Now()
						w.Write([]byte("Task created successfully"))
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

func listHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/list" {
		listDetailFunc(w, r)
		return
	}

	listID := r.URL.Path[len("/list/"):]
	if listID == "" {
		listDetailFunc(w, r)
		return
	}
	if listID == "create" {
		createListFunc(w, r)
		return
	}

	if strings.HasSuffix(listID, "/delete") && !strings.Contains(listID, "/task/") {
		deleteListFunc(w, r, listID)
		return
	}
	if strings.HasSuffix(listID, "/update") && !strings.Contains(listID, "/task/") {
		updateListFunc(w, r, listID)
		return
	}

	// If url /list/{listID}/task/{taskid} then return task details
	if strings.Contains(listID, "/task/") {
		taskHandler(w, r)
		return
	}

	// If url /list/{listID} then return list details
	todoListFunc(w, r, listID)
	w.WriteHeader(http.StatusNotFound)
	return

}

func handleRequests() {

	http.HandleFunc("/login", login)
	//
	http.HandleFunc("/list/", isAuthenticated(listHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
