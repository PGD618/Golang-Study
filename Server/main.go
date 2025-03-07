package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

// 结构体定义
type Person struct {
	M_Name   string `json:"M_Name"`
	M_Age    int    `json:"M_Age"`
	M_Sex    string `json:"M_Sex"`
	M_Number string `json:"M_Number"`
}

// 全局变量
var (
	db  *sql.DB
	rdb *redis.Client
)

// 初始化数据库
func init() {
	// 初始化 MySQL
	var err error
	db, err = sql.Open("mysql", "root:123abc@tcp(127.0.0.1:3306)/text")
	if err != nil {
		log.Println("MySQL 连接失败！", err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 5)

	err = db.Ping()
	if err != nil {
		log.Println("MySQL 连接失败！", err)
	}
	fmt.Println("MySQL 连接成功！")

	// 初始化 Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		PoolSize: 100,
	})

	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Println("Redis 连接失败！", err)
	}
	fmt.Println("Redis 连接成功！")
}

// 添加员工信息到 MySQL
func InsertToMySQL(person Person) error {
	query := "INSERT INTO employees (Name, Age, Sex, Number) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, person.M_Name, person.M_Age, person.M_Sex, person.M_Number)
	return err
}

// 添加员工信息到 Redis
func InsertToRedis(person Person) error {
	ctx := context.Background()
	jsonData, err := json.Marshal(person)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, person.M_Name, jsonData, 0).Err()
}

// 从 MySQL 删除员工信息
func DeleteFromMySQL(name string) error {
	query := "DELETE FROM employees WHERE Name = ?"
	_, err := db.Exec(query, name)
	return err
}

// 从 Redis 删除员工信息
func DeleteFromRedis(name string) error {
	ctx := context.Background()
	return rdb.Del(ctx, name).Err()
}

// 从 MySQL 查询单个员工信息
func GetFromMySQL(name string) (Person, error) {
	var person Person
	query := "SELECT Name, Age, Sex, Number FROM employees WHERE Name = ?"
	err := db.QueryRow(query, name).Scan(&person.M_Name, &person.M_Age,
		&person.M_Sex, &person.M_Number)
	if err != nil {
		return Person{}, err
	}
	return person, nil
}

// 从 Redis 中查询员工信息
func GetFromRedis(name string) (Person, error) {
	ctx := context.Background()
	val, err := rdb.Get(ctx, name).Result()
	if err != nil {
		return Person{}, err
	}
	var person Person
	if err := json.Unmarshal([]byte(val), &person); err != nil {
		return Person{}, err
	}
	return person, nil
}

// 添加员工信息
func AddPerson(w http.ResponseWriter, r *http.Request) {
	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go func(person Person) {
		if err := InsertToMySQL(person); err != nil {
			log.Println("数据存入 MySQL 失败！", err)
		}
		if err := InsertToRedis(person); err != nil {
			log.Println("数据存入 Redis 失败！", err)
		}
	}(person)

	fmt.Fprintf(w, "添加成功！")
}

// 显示员工信息
func ShowPerson(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT Name, Age, Sex, Number FROM employees")
	if err != nil {
		log.Println("查询失败！", err)
		return
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		var person Person
		err := rows.Scan(&person.M_Name, &person.M_Age, &person.M_Sex, &person.M_Number)
		if err != nil {
			log.Println("解析数据失败！", err)
			return
		}
		persons = append(persons, person)
	}

	json.NewEncoder(w).Encode(persons)
}

// 查询员工信息
func FindPerson(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	// 使用 Channel 接收结果
	resultChan := make(chan Person)
	errChan := make(chan error)

	// 先查 Redis，如果 Redis 中没有，查 MySQL
	go func() {
		person, err := GetFromRedis(name)
		if err == nil {
			resultChan <- person
			return
		}

		person, err = GetFromMySQL(name)
		if err != nil {
			errChan <- err
			return
		}

		if err := InsertToRedis(person); err != nil {
			log.Println("数据写入 Redis 失败！", err)
		}

		resultChan <- person
	}()

	// 监听结果
	select {
	case person := <-resultChan:
		json.NewEncoder(w).Encode(person)
	case err := <-errChan:
		if err == sql.ErrNoRows {
			http.Error(w, "没有找到相关信息！", http.StatusNotFound)
		} else {
			http.Error(w, "查询失败！", http.StatusInternalServerError)
		}
	case <-time.After(2 * time.Second):
		http.Error(w, "查询超时！", http.StatusRequestTimeout)
	}
}

// 修改员工信息
func ModifyPerson(w http.ResponseWriter, r *http.Request) {
	oldname := r.URL.Query().Get("name")
	if oldname == "" {
		http.Error(w, "缺少员工姓名参数", http.StatusBadRequest)
		return
	}

	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go func(person Person, oldname string) {
		query := "UPDATE employees SET Name = ?, Age = ?, Sex = ?, Number = ? WHERE Name = ?"
		_, err := db.Exec(query, person.M_Name, person.M_Age, person.M_Sex,
			person.M_Number, oldname)
		if err != nil {
			log.Println("MySQL 更新失败！", err)
		}
		if err := DeleteFromRedis(oldname); err != nil {
			log.Println("Redis 更新失败！", err)
		}
		if err := InsertToRedis(person); err != nil {
			log.Println("Redis 更新失败！", err)
		}
	}(person, oldname)

	fmt.Fprintf(w, "修改成功！")
}

// 删除员工信息
func DeletePerson(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	go func(name string) {
		if err := DeleteFromMySQL(name); err != nil {
			log.Println("MySQL 删除失败！", err)
		}
		if err := DeleteFromRedis(name); err != nil {
			log.Println("Redis 删除失败！", err)
		}
	}(name)

	fmt.Fprintf(w, "删除成功！")
}

// 清空员工信息
func ClearPerson(w http.ResponseWriter, r *http.Request) {
	go func() {
		if _, err := db.Exec("DELETE FROM employees"); err != nil {
			log.Println("MySQL 清空失败！", err)
		}
		ctx := context.Background()
		if err := rdb.FlushDB(ctx).Err(); err != nil {
			log.Println("Redis 清空失败！", err)
		}
	}()

	fmt.Fprintf(w, "员工信息已清空！")
}

func main() {
	http.HandleFunc("/add", AddPerson)
	http.HandleFunc("/show", ShowPerson)
	http.HandleFunc("/find", FindPerson)
	http.HandleFunc("/modify", ModifyPerson)
	http.HandleFunc("/delete", DeletePerson)
	http.HandleFunc("/clear", ClearPerson)

	log.Println("服务器启动，端口：8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("服务器启动失败！", err)
	}
}
