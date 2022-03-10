package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db, err = sql.Open("mysql", "root:manoj@123@tcp(127.0.0.1:3306)/student_db")

type Student struct {
	Id       int    `db:"id"`
	Name     string `db:"name" form:"name"`
	Email    string `db:"email" form:"email"`
	Gender   string `db:"gender" form:"gender"`
	Mobile   string `db:"mobile" form:"mobile"`
	Password string `db:"mobile" form:"password"`
}
type User struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Error struct {
	Message string `json:"message"`
}
type JWT struct {
	Token string `json:"token"`
}

func getall(c *gin.Context) {
	var (
		student  Student
		students []Student
	)
	rows, err := db.Query("select id,name,email,gender,mobile from student;")
	if err != nil {
		fmt.Println(err.Error())
	}
	for rows.Next() {
		rows.Scan(&student.Id, &student.Name, &student.Email, &student.Gender, &student.Mobile)
		students = append(students, student)
	}
	defer rows.Close()
	c.JSON(http.StatusOK, students)
}
func getbyid(c *gin.Context) {
	var student Student
	id := c.Param("id")
	row := db.QueryRow("select id,name,email,gender,mobile from student where id = ?;", id)
	err = row.Scan(&student.Id, &student.Name, &student.Email, &student.Gender, &student.Mobile)
	if err != nil {
		c.JSON(http.StatusOK, nil)
	} else {
		c.JSON(http.StatusOK, student)
	}
}
func delete(c *gin.Context) {
	id := c.Param("id")
	stmt, err := db.Prepare("delete from student where id=?;")
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = stmt.Exec(id)
	if err != nil {
		fmt.Println(err.Error())
	}
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("sucessfully delete the student id :%s", id),
	})
}
func InsertAdd(c *gin.Context) {
	Name := c.PostForm("name")
	Email := c.PostForm("email")
	Gender := c.PostForm("gender")
	Mobile := c.PostForm("mobile")
	stmt, err := db.Prepare("insert into student(name,email,gender,mobile)values(?,?,?,?);")
	if err != nil {
		fmt.Print(err.Error())
	}
	_, err = stmt.Exec(Name, Email, Gender, Mobile)
	if err != nil {
		fmt.Print(err.Error())
	}
	log.Println("INSERT: Name: " + Name + " | Email: " + Email)
	defer stmt.Close()
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("successfully created"),
	})
}
func update(c *gin.Context) {
	Id, err := strconv.Atoi(c.Param("id"))
	Name := c.PostForm("Name")
	Email := c.PostForm("email")
	Gender := c.PostForm("gender")
	Mobile := c.PostForm("mobile")
	stmt, err := db.Prepare("update student set name= ?,email= ?,gender=?, mobile= ? where id= ?;")
	if err != nil {
		fmt.Print(err.Error())
	}
	_, err = stmt.Exec(Name, Email, Gender, Mobile, Id)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer stmt.Close()
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Successfully updated"),
	})
}
func Login(c *gin.Context) {
	var student Student
	Email := c.PostForm("email")
	Password := c.PostForm("Password")
	stmt, err := db.Prepare("SELECT name, email,gender,mobile FROM student WHERE email=? AND password=?;")
	if err != nil {
		fmt.Print(err.Error())
	}
	defer stmt.Close()
	row := stmt.QueryRow(Email, Password)
	err = row.Scan(&student.Id, &student.Name, &student.Email, &student.Gender, &student.Mobile)
	if err != nil {
		fmt.Print(err.Error())
	}
	c.JSON(http.StatusOK, student)
}
func GenrateTokenn(Email string) (string, error) {
	var err error
	secret := "secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"email":      Email,
		"expiresIn":  time.Now().Add(time.Minute * 25).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Fatal(err)
	}
	return tokenString, nil
}
func Login1(c *gin.Context) {
	var u User
	json.NewDecoder(c.Request.Body).Decode(&u)
	var user User
	stmt, serr := db.Prepare("select email,password from user1 where email = ?;")
	if serr != nil {
		fmt.Println("s   ", serr)
	}
	defer stmt.Close()
	result := stmt.QueryRow(u.Email)
	rerr := result.Scan(&user.Email, &user.Password)
	if serr != nil {
		fmt.Println("r   ", rerr)
	}

	if u.Email != user.Email || u.Password != user.Password {
		fmt.Println("Invalid credentials")
	} else {
		token, err := GenrateTokenn(user.Email)
		if err != nil {
			fmt.Println("tok      ", err)
		}
		c.JSON(http.StatusOK, gin.H{"token": token, "user": user.Email})

	}
}
func middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, "Bearer ")

		if len(bearerToken) <2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated."})
			c.Abort()
			return
		}
		authToken := bearerToken[1]
		valid:=verifyToken(authToken)
		if valid==false{
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated."})
			c.Abort()
		}else{
			c.Next()
		}

	}
}
func verifyToken(tokenString string) bool {
	token, error := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}

		return []byte("secret"), nil
	})
	if error != nil {
		fmt.Println(error)
		return false
	}
	if token.Valid {
		return true
	} else {
		return false
	}

}
func createTable() {
	stmt, err := db.Prepare("CREATE TABLE user (id int NOT NULL AUTO_INCREMENT, name varchar(40),email varchar(60),gender varchar(60), mobile varchar(15), PRIMARY KEY (id));")
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Table is successfully created....")
	}
}
func main() {
	//createTable()
	router := gin.Default()
	router.GET("/api/student/:id", getbyid)
	router.GET("/api/student", middleware(), getall)
	router.POST("/api/student", InsertAdd)
	router.PUT("/api/student/:id", update)
	router.DELETE("/api/student/:id", delete)
	router.GET("/api/student_login", Login)
	router.POST("api/user_Login", Login1)
	// router.GET("api/valid_token",verifyToken)
	router.Run(":8028")
}

//queryInsertUser = "INSERT INTO users(first_name, last_name, email, date_created) VALUES(?, ?, ?, ?);"
