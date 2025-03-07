package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// 定义结构体
type Person struct {
	M_Name   string `json:"M_Name"`
	M_Age    int    `json:"M_Age"`
	M_Sex    string `json:"M_Sex"`
	M_Number string `json:"M_Number"`
}

// 添加员工信息
func AddPerson() {
	var (
		name, sex, number string
		age               int
	)
	fmt.Println("请输入员工姓名：")
	fmt.Scanln(&name)
	fmt.Println("请输入员工年龄：")
	fmt.Scanln(&age)
	fmt.Println("请输入员工性别：")
	fmt.Scanln(&sex)
	fmt.Println("请输入员工电话：")
	fmt.Scanln(&number)

	person := Person{
		M_Name:   name,
		M_Age:    age,
		M_Sex:    sex,
		M_Number: number,
	}
	jsonData, err := json.Marshal(person)
	if err != nil {
		fmt.Println("json编码失败！", err)
		return
	}

	resp, err := http.Post("http://127.0.0.1:8080/add", "application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("请求失败！", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("添加员工信息响应：", resp.Status)
}

// 显示员工信息
func ShowPerson() {
	resp, err := http.Get("http://127.0.0.1:8080/show")
	if err != nil {
		fmt.Println("请求失败！", err)
		return
	}
	defer resp.Body.Close()

	var persons []Person
	err = json.NewDecoder(resp.Body).Decode(&persons)
	if err != nil {
		fmt.Println("解析响应失败！", err)
		return
	}

	if len(persons) == 0 {
		fmt.Println("当前没有员工信息！")
		return
	}
	for _, person := range persons {
		fmt.Printf("姓名：%s\t年龄：%d\t性别：%s\t电话：%s\n",
			person.M_Name, person.M_Age, person.M_Sex, person.M_Number)
	}

	fmt.Println("显示员工信息响应：", resp.Status)
}

// 查找员工信息
func FindPerson() {
	var name string
	fmt.Println("请输入需要查找的员工姓名：")
	fmt.Scanln(&name)

	resp, err := http.Get("http://127.0.0.1:8080/find?name=" + name)
	if err != nil {
		fmt.Println("请求失败！", err)
		return
	}
	defer resp.Body.Close()

	var person Person
	err = json.NewDecoder(resp.Body).Decode(&person)
	if err != nil {
		fmt.Println("解析响应失败！", err)
		return
	}

	if person.M_Name == "" {
		fmt.Println("未找到相关信息！")
	} else {
		fmt.Printf("姓名：%s\t年龄：%d\t性别：%s\t电话：%s\n",
			person.M_Name, person.M_Age, person.M_Sex, person.M_Number)
	}
	fmt.Println("查找员工信息响应：", resp.Status)
}

// 修改员工信息
func ModifyPerson() {
	var (
		oldname, name, sex, number string
		age                        int
	)
	fmt.Println("请输入要修改信息的员工姓名：")
	fmt.Scanln(&oldname)
	resp, err := http.Get("http://127.0.0.1:8080/find?name=" + oldname)
	if err != nil {
		fmt.Println("查询员工信息失败！")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("未找到员工信息！", err)
		return
	}

	fmt.Println("请输入新的员工姓名：")
	fmt.Scanln(&name)
	fmt.Println("请输入新的员工年龄：")
	fmt.Scanln(&age)
	fmt.Println("请输入新的员工性别：")
	fmt.Scanln(&sex)
	fmt.Println("请输入新的员工电话：")
	fmt.Scanln(&number)

	person := Person{
		M_Name:   name,
		M_Age:    age,
		M_Sex:    sex,
		M_Number: number,
	}
	jsonData, err := json.Marshal(person)
	if err != nil {
		fmt.Println("json编码失败！")
		return
	}
	request, err := http.NewRequest("PUT", "http://127.0.0.1:8080/modify?name="+oldname,
		bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("创建请求失败！", err)
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(request)
	if err != nil {
		fmt.Println("请求失败！", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("修改员工信息响应：", resp.Status)
}

// 删除员工信息
func DeletePerson() {
	var name string
	fmt.Println("请输入要删除的员工姓名：")
	fmt.Scanln(&name)

	request, err := http.NewRequest("DELETE", "http://127.0.0.1:8080/delete?name="+name, nil)
	if err != nil {
		fmt.Println("创建请求失败！", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println("请求失败！", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("删除员工信息响应：", resp.Status)
}

// 清空员工信息
func ClearPerson() {
	resp, err := http.Get("http://127.0.0.1:8080/clear")
	if err != nil {
		fmt.Println("请求失败！", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("清空员工响应：", resp.Status)
}

// 显示操作菜单
func ShowMenu() {
	fmt.Println("##################################")
	fmt.Println("#----------公司HR管理系统---------#")
	fmt.Println("#----------1. 添加员工信息--------#")
	fmt.Println("#----------2. 显示员工信息--------#")
	fmt.Println("#----------3. 查询员工信息--------#")
	fmt.Println("#----------4. 修改员工信息--------#")
	fmt.Println("#----------5. 删除员工信息--------#")
	fmt.Println("#----------6. 清空员工信息--------#")
	fmt.Println("#----------0. 退出程序-----------#")
	fmt.Println("##################################")
	fmt.Print("请选择操作：")
}

func main() {
	for {
		var op int
		ShowMenu()
		fmt.Scanln(&op)

		switch op {
		case 1:
			AddPerson()
		case 2:
			ShowPerson()
		case 3:
			FindPerson()
		case 4:
			ModifyPerson()
		case 5:
			DeletePerson()
		case 6:
			ClearPerson()
		case 0:
			fmt.Println("退出程序，欢迎下次使用！")
			return
		default:
			fmt.Println("输入错误，请重新输入！")
		}
	}
}
