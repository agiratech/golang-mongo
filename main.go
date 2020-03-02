package main

import (
  "fmt"
  "net/http"
  "os"
  "time"

  "github.com/gin-gonic/gin"
  mgo "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
)

// database and collection names are statically declared
const database, collection = "go-mongo-practice", "user"

// User structure
type User struct {
  ID        bson.ObjectId `bson:"_id"`
  Name      string        `bson:"name"`
  Address   string        `bson:"address"`
  Age       int           `bson:"age"`
  CreatedAt time.Time     `bson:"created_at"`
  UpdatedAt time.Time     `bson:"updated_at"`
}

// Users list
type Users []User

// DB connection
func connect() *mgo.Session {
  session, err := mgo.Dial("localhost")
  if err != nil {
    fmt.Println("session err:", err)
    os.Exit(1)
  }
  return session
}

// getUser function
func getUser(id bson.ObjectId) (User, error) {
  user := User{}
  session := connect()
  defer session.Close()
  err := session.DB(database).C(collection).Find(bson.M{"_id": &id}).One(&user)
  return user, err
}

// StartService function
func StartService() {
  router := gin.Default()
  api := router.Group("/api")
  {
    // List users
    api.GET("/users", func(c *gin.Context) {
      users := Users{}
      session := connect()
      defer session.Close()
      err := session.DB(database).C(collection).Find(bson.M{}).All(&users)

      if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Users are not exist"})
        return
      }
      c.JSON(http.StatusOK, gin.H{"status": "success", "users": &users})
    })
    // Create user record
    api.POST("/users", func(c *gin.Context) {
      user := User{}
      err := c.Bind(&user)

      if err != nil {
        c.JSON(http.StatusBadRequest,
          gin.H{
            "status": "failed",
            "message": "Invalid request body",
          })
        return
      }
      user.ID = bson.NewObjectId()
      user.CreatedAt, user.UpdatedAt = time.Now(), time.Now()
      session := connect()
      defer session.Close()
      err = session.DB(database).C(collection).Insert(user)

      if err != nil {
        c.JSON(http.StatusBadRequest,
          gin.H{
            "status": "failed",
            "message": "Error in the user insertion",
          })
        return
      }
      c.JSON(http.StatusOK,
        gin.H{
          "status": "success",
          "user": &user,
        })
    })
    // Retrieve user recor
    api.GET("/users/:id", func(c *gin.Context) {
      var id bson.ObjectId = bson.ObjectIdHex(c.Param("id")) // Get Param
      user, err := getUser(id)

      if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Invalid ID"})
        return
      }
      c.JSON(http.StatusOK, gin.H{"status": "success", "user": &user})
    })
    // Update user record
    api.PUT("/users/:id", func(c *gin.Context) {
      var id bson.ObjectId = bson.ObjectIdHex(c.Param("id")) // Get Param
      existingUser, err := getUser(id)

      if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid ID"})
        return
      }
      err = c.Bind(&existingUser)

      if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request body"})
        return
      }
      existingUser.UpdatedAt = time.Now()
      session := connect()
      defer session.Close()
      err = session.DB(database).C(collection).Update(bson.M{"_id": &id}, existingUser)
      if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Error in the user updation"})
        return
      }
      c.JSON(http.StatusOK, gin.H{"status": "success", "user": &existingUser})
    })
    // Remove user record
    api.DELETE("/users/:id", func(c *gin.Context) {
      var id bson.ObjectId = bson.ObjectIdHex(c.Param("id")) // Get Param
      session := connect()
      defer session.Close()
      err := session.DB(database).C(collection).Remove(bson.M{"_id": &id})
      if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Error in the user deletion"})
        return
      }
      c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User deleted successfully"})
    })
  }

  router.NoRoute(func(c *gin.Context) {
    c.AbortWithStatus(http.StatusNotFound)
  })
  router.Run(":8000")
}

func main() {
  StartService()
}
