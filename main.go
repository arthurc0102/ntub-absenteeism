package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	osUser "os/user"
	"path"
	"strings"
	"syscall"

	"github.com/PuerkitoBio/goquery"

	"golang.org/x/crypto/ssh/terminal"
)

const debug = true
const tag = "ntub-attendance"
const baseURL = "http://ntcbadm.ntub.edu.tw"
const loginURL = baseURL + "/login.aspx"
const loginSuccessURL = baseURL + "/Portal/indexSTD.aspx"

var filePath = func() string {
	u, e := osUser.Current()
	check(e, true)
	return path.Join(u.HomeDir, "."+tag+".json")
}()

type user struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *user) init() {
	u.load()

	if u.Username != "" && u.Password != "" {
		return
	}

	u.Username = input("Account: ")
	u.Password = inputPassword()

	if input("Do you want to save your info? [Y/n] ") == "n" {
		return
	}

	u.export()
}

func (u user) toJSON(pretty bool) string {
	var bytes []byte
	var err error

	if pretty {
		bytes, err = json.MarshalIndent(u, "", "    ")
	} else {
		bytes, err = json.Marshal(u)
	}

	check(err, true)
	return string(bytes)
}

func (u *user) load() {
	if _, err := os.Stat(filePath); err != nil {
		return
	}

	raw, err := ioutil.ReadFile(filePath)
	json.Unmarshal(raw, &u)
	check(err, true)
	return
}

func (u user) export() {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if os.IsNotExist(err) {
		file, err = os.Create(filePath)
		check(err, true)
	}

	defer file.Close()
	file.WriteString(u.toJSON(true) + "\n")
}

func main() {
	var currentUser user
	currentUser.init()

	client, loginResult := login(currentUser.Username, currentUser.Password)
	if !loginResult {
		fmt.Println("登入失敗")
		os.Exit(0)
	}

	getAttendance(client)
}

func login(username string, password string) (*http.Client, bool) {
	doc, err := goquery.NewDocument(loginURL)
	check(err, true)

	data := url.Values{}
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		name := s.AttrOr("name", "")
		if name == "" {
			return
		}

		value := s.AttrOr("value", "")
		data.Add(name, value)
	})

	data.Set("UserID", username)
	data.Set("PWD", password)

	client := &http.Client{}
	client.Jar, err = cookiejar.New(nil)
	check(err, true)

	res, err := client.PostForm(loginURL, data)
	check(err, true)

	defer res.Body.Close()
	return client, res.Request.URL.String() == loginSuccessURL
}

func getAttendance(client *http.Client) {

}

func check(err error, leave bool) {
	if err == nil {
		return
	}

	fmt.Println(err)

	if leave {
		os.Exit(0)
	}
}

func input(ask string) string {
	fmt.Print(ask)
	reader := bufio.NewReader(os.Stdin)
	result, err := reader.ReadString('\n')

	if err != nil {
		log.Fatalln(err)
	}

	return strings.Replace(result, "\n", "", -1)
}

func inputPassword() string {
	fmt.Print("Password: ")
	bytesPassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")
	check(err, true)
	return string(bytesPassword)
}
