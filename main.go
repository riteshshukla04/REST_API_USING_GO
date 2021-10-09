package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)



var client *mongo.Client


//<---------Data Models ----------- >
type User struct{
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
	Email string `json:"email" bson:"email"`
	Password []byte `json:"password" bson:"password"`

	
}
type Post struct{
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Caption string `json:"caption" bson:"captions"`
	ImageURL string `json:"imageURL" bson:"imageURL"`
	Posted_Timestamp time.Time `json:"postedTimestamp" bson:"postedTimestamp"`
	User *User  `json:"user" bson:"user"`
}




//<-------End Point Functions ----------- >
func getUser(Response http.ResponseWriter,Request *http.Request){
	Response.Header().Set("Content-Type","application/json")
	var user []User
	collection:=client.Database("Appointy").Collection("User")
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err :=collection.Find(ctx,bson.M{})
	if err!=nil{
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx){
		var user_cursor User
		cursor.Decode(&user_cursor)
		user=append(user, user_cursor)
	}
	if err :=cursor.Err();err!=nil{
		return
	}
	json.NewEncoder(Response).Encode(user)

}



func getUserByID(Response http.ResponseWriter,Request *http.Request){
	Response.Header().Set("Content-Type","application/json")
	var user []User
	params:=mux.Vars(Request)
	id,_:=primitive.ObjectIDFromHex(params["id"])
	collection:=client.Database("Appointy").Collection("User")
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err :=collection.Find(ctx,bson.M{})
	if err!=nil{
		fmt.Println(err)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx){
		var user_cursor User
		cursor.Decode(&user_cursor)
		user=append(user, user_cursor)
	}
	if err :=cursor.Err();err!=nil{
		fmt.Println(err)
		return
	}

	for i:=0;i<len(user);i++{
		fmt.Println(id)
		if user[i].Id==id{
			json.NewEncoder(Response).Encode(user[i])
			return
		}
	}
}



func getPost(Response http.ResponseWriter,Request *http.Request){
	Response.Header().Set("Content-Type","application/json")
	var post []Post
	params:=mux.Vars(Request)
	id,_:=primitive.ObjectIDFromHex(params["id"])
	collection:=client.Database("Appointy").Collection("Post")
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err :=collection.Find(ctx,bson.M{})
	if err!=nil{
		fmt.Println(err)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx){
		var post_cursor Post
		cursor.Decode(&post_cursor)
		post=append(post, post_cursor)
	}
	if err :=cursor.Err();err!=nil{
		fmt.Println(err)
		return
	}

	for i:=0;i<len(post);i++{
		fmt.Println(id)
		if post[i].Id==id{
			json.NewEncoder(Response).Encode(post[i])
			return
		}
	}
}



func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}



func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}




func createUser(Response http.ResponseWriter,Request *http.Request){
	Response.Header().Set("Content-Type","application/json")
	fmt.Println(os.Getenv("p"))
	var user User


	_=json.NewDecoder(Request.Body).Decode(&user)
	user.Password=encrypt(user.Password,os.Getenv(PASSWORD_SSH))
	collection:=client.Database("Appointy").Collection("User")
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	result,_:=collection.InsertOne(ctx,user)
	json.NewEncoder(Response).Encode(result)
}





func createPost(Response http.ResponseWriter,Request *http.Request){
	Response.Header().Set("Content-Type","application/json")
	var post Post
	_=json.NewDecoder(Request.Body).Decode(&post)
	post.Posted_Timestamp=time.Now()
	collection:=client.Database("Appointy").Collection("Post")
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	result,_:=collection.InsertOne(ctx,post)
	json.NewEncoder(Response).Encode(result)
}




func getPostByUser(Response http.ResponseWriter,Request *http.Request){
	Response.Header().Set("Content-Type","application/json")
	var post []Post
	params:=mux.Vars(Request)
	id,_:=primitive.ObjectIDFromHex(params["id"])
	collection:=client.Database("Appointy").Collection("Post")
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	cursor,err :=collection.Find(ctx,bson.M{})
	if err!=nil{
		fmt.Println(err)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx){
		var post_cursor Post
		cursor.Decode(&post_cursor)
		post=append(post, post_cursor)
	}
	if err :=cursor.Err();err!=nil{
		fmt.Println(err)
		return
	}
	var userPosts []Post
	for i:=0;i<len(post);i++{
		fmt.Println(post[i].User.Id)
		if post[i].User.Id==id{
			userPosts = append(userPosts, post[i])
			
		}
		
	}
	json.NewEncoder(Response).Encode(userPosts)
}



func main(){
	ctx,_:=context.WithTimeout(context.Background(),10*time.Second)
	client,_=mongo.Connect(ctx,options.Client().ApplyURI("mongodb://localhost:27017"))
	Request :=mux.NewRouter()
	Request.HandleFunc("/users",createUser).Methods("POST")
	Request.HandleFunc("/users",getUser).Methods("GET")
	Request.HandleFunc("/users/{id}",getUserByID).Methods("GET")
	Request.HandleFunc("/posts",createPost).Methods("POST")
	Request.HandleFunc("/posts/{id}",getPost).Methods("GET")
	Request.HandleFunc("/posts/users/{id}",getPostByUser).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000",Request))
}