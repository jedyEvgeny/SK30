// В приложении изменены типы запросов в разделе Body в месте где вводим числа: со стрингов на инты
// см. примеры запросов перед функциями ниже
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi"
)

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Friends []int  `json:"friends"`
}

var (
	users      = make(map[int]User)
	usersMutex sync.RWMutex
	userID     int
)

func main() {
	r := chi.NewRouter()

	r.Post("/create", CreateUser)
	r.Post("/make_friends", MakeFriends)
	r.Delete("/user", DeleteUser)
	r.Get("/friends/{id}", GetFriends)
	r.Put("/{id}", UpdateAge)
	fmt.Println("Сервер запущен")

	http.ListenAndServe(":8080", r)
}

// {"name":"Анжелика","age":24,"friends":[]}
func CreateUser(w http.ResponseWriter, r *http.Request) {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID++
	newUser.ID = userID
	users[newUser.ID] = newUser

	w.WriteHeader(http.StatusCreated)
	response := fmt.Sprint("id:", newUser.ID)
	w.Write([]byte(response))
}

// {"source_id":1,"target_id":2}
func MakeFriends(w http.ResponseWriter, r *http.Request) {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	var data struct {
		SourceID int `json:"source_id"`
		TargetID int `json:"target_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sourceUser, sourceExists := users[data.SourceID]
	targetUser, targetExists := users[data.TargetID]

	if !sourceExists || !targetExists {
		http.Error(w, "Один или оба пользователя не существуют", http.StatusBadRequest)
		return
	}

	sourceUser.Friends = append(sourceUser.Friends, data.TargetID)
	targetUser.Friends = append(targetUser.Friends, data.SourceID)

	users[data.SourceID] = sourceUser
	users[data.TargetID] = targetUser

	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf("%s и %s теперь друзья", sourceUser.Name, targetUser.Name)
	w.Write([]byte(response))
}

// {"target_id":3}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	var data struct {
		TargetID int `json:"target_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, exists := users[data.TargetID]
	if !exists {
		http.Error(w, "Пользователь не существует", http.StatusBadRequest)
		return
	}

	delete(users, data.TargetID)

	//Удаляем пользователя из списка друзей всех других пользователей
	for index, u := range users {
		var newFriends []int
		for _, friendID := range u.Friends {
			if friendID != data.TargetID {
				newFriends = append(newFriends, friendID)
			}
		}
		users[index] = User{
			ID:      u.ID,
			Name:    u.Name,
			Age:     u.Age,
			Friends: newFriends,
		}
	}

	w.WriteHeader(http.StatusOK)
	response := fmt.Sprint("Сообщение:", user.Name)
	w.Write([]byte(response))
}

func GetFriends(w http.ResponseWriter, r *http.Request) {
	usersMutex.RLock()
	defer usersMutex.RUnlock()

	userID := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, exists := users[id]
	if !exists {
		http.Error(w, "Пользователь не существует", http.StatusBadRequest)
		return
	}

	friendNames := make([]string, 0, len(user.Friends))
	for _, friendID := range user.Friends {
		friend, friendExists := users[friendID]
		if friendExists {
			friendNames = append(friendNames, friend.Name)
		}
	}

	w.WriteHeader(http.StatusOK)
	response := fmt.Sprint("Друзья:", friendNames)
	w.Write([]byte(response))
}

// {"new age":19}
func UpdateAge(w http.ResponseWriter, r *http.Request) {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	userID := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var data struct {
		NewAge int `json:"new_age"`
	}
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, exists := users[id]
	if !exists {
		http.Error(w, "Пользователь не найден", http.StatusBadRequest)
		return
	}

	user.Age = data.NewAge
	users[id] = user

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Возраст пользователя успешно обновлён"))
}
