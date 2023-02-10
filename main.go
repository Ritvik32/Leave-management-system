package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"

	//"github.com/cosmos/cosmos-sdk/x/ibc/core/client"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type AllLeave struct {
	Message string
	Leaves  []LeaveRequest
}

var JwtKey = "something"

type GetU struct {
	User    User   `json:"user"`
	Status  int    `json:"status" bson:"status"`
	Message string `json:"message" bson:"message"`
	Id      interface{}
}

type User struct {
	Id   string `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
	Role Role   `json:"role" bson:"role"`
}
type Role int

const (
	Student Role = iota
	Admin
)

type Leave struct {
	Id       string `json:"id" bson:"id"`
	From     string `json:"from" bson:"from"`
	To       string `json:"to" bson:"to"`
	Approved bool   `json:"approved" bson:"approved"`
}

type LeaveRequest struct {
	Id   string `json:"id" bson:"id"`
	From string `json:"from" bson:"from"`
	To   string `json:"to" bson:"to"`
}

type LeaveResponse struct {
	Status  int         `json:"status" bson:"status"`
	Message string      `json:"message" bson:"message"`
	Id      interface{} `json:"id" bson:"id"`
}
type CheckStatus struct {
	Status   int  `json:"status" bson:"status"`
	Approved bool `json:"approved" bson:"approved"`
}
type ListLeave struct {
	Id string `json:"id" bson:"id"`
}
type JwtToken struct {
	Token string `json:"token"`
}
type Exception struct {
	Message string `json:"message"`
}

var UserCollection *mongo.Collection
var LeaveCollection *mongo.Collection

func main() {
	r := mux.NewRouter()
	//usersR := r.PathPrefix("/users").Subrouter()
	//usersR.Path("").Methods(http.MethodGet).HandlerFunc(getUsers)
	r.Path("/users").Methods(http.MethodPost).HandlerFunc(createUser)
	//usersR1 := r.PathPrefix("/leaves").Subrouter()
	r.Path("/leaves").Methods(http.MethodPost).HandlerFunc(sendRequest)
	//	usersR1.Path("").Methods(http.MethodPost).HandlerFunc(getResponse)
	r.Path("/users/{id}/leaves").Methods(http.MethodGet).HandlerFunc(listLeaves)
	r.Path("/users/{userId}/leaves/{leaveId}").Methods(http.MethodGet).HandlerFunc(checkStatus)
	r.Path("/authenticate").Methods((http.MethodPost)).HandlerFunc(getJWt)
	r.Path("/approve/{_id}").Methods((http.MethodPost)).HandlerFunc(validateJWT(ApproveLeave))

	clientOptions := options.Client().ApplyURI("mongodb+srv://ritvik:ritvik@cluster0.x1pdgs7.mongodb.net/?retryWrites=true&w=majority")

	// Connect to MongoDB
	dbctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(dbctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(dbctx)

	UserCollection = client.Database("LeaveManagement").Collection("User")
	LeaveCollection = client.Database("LeaveManagement").Collection("Leave")

	//Collection = db.Conn()
	fmt.Println("------------------------------")
	log.Fatal(http.ListenAndServe(":8000", r))

}

func createUser(w http.ResponseWriter, r *http.Request) {
	//u := User{}
	var p User
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		fmt.Println(err)
		http.Error(w, "Error decoidng response object", http.StatusBadRequest)
		return
	}
	id := p.Id
	name := p.Name
	role := p.Role
	u := User{
		Id:   id,
		Name: name,
		Role: role,
	}

	result, err := UserCollection.InsertOne(context.TODO(), u)
	if err != nil {
		json.NewEncoder(w).Encode(GetU{
			Status:  400,
			Message: err.Error(),
		})
		return
	}
	fmt.Println(result)
	json.NewEncoder(w).Encode(GetU{
		Status: 200,
		Id:     result.InsertedID,
	})

}

//students enter leave request
func sendRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("In sedn Req")
	//u := User{}
	var p LeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		fmt.Println(err)
		http.Error(w, "Error decoidng response object", http.StatusBadRequest)
		return
	}
	fmt.Println("LeaveReq====", p)
	id := p.Id
	from := p.From
	to := p.To

	q := Leave{
		Id:       id,
		From:     from,
		To:       to,
		Approved: false,
	}
	result, err := LeaveCollection.InsertOne(context.TODO(), q)
	if err != nil {
		json.NewEncoder(w).Encode(LeaveResponse{
			Status:  400,
			Message: err.Error(),
		})
		return
	}
	//fmt.Println(result)
	json.NewEncoder(w).Encode(LeaveResponse{
		Status: 200,
		Id:     result.InsertedID,
	})

}

//students can check their leave status
func checkStatus(w http.ResponseWriter, r *http.Request) {
	var p Leave
	q := mux.Vars(r)

	//checking with object id of the leave request
	leaveId := q["leaveId"]
	userId := q["userId"]
	oid, err := primitive.ObjectIDFromHex(leaveId)
	if err != nil {
		panic(err)
	}
	fmt.Println("id is", leaveId)
	filter := bson.M{
		"_id": oid,
		"id":  userId,
	}

	err = LeaveCollection.FindOne(context.TODO(), filter).Decode(&p)
	if err != nil {
		json.NewEncoder(w).Encode(LeaveResponse{
			Status:  400,
			Message: err.Error(),
		})
		return

	} else {
		if p.Approved == true {
			json.NewEncoder(w).Encode(CheckStatus{
				Status:   400,
				Approved: p.Approved,
			})
		}
		json.NewEncoder(w).Encode(CheckStatus{
			Status:   200,
			Approved: p.Approved,
		})

	}

}

//allows admin to list all the leaves with respect to a student
func listLeaves(w http.ResponseWriter, r *http.Request) {
	//	var p ListLeave[]
	q := mux.Vars(r)

	id := q["id"]
	fmt.Println(id)

	filter := bson.M{
		"id": id,
	}
	var results []LeaveRequest
	cur, err1 := LeaveCollection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var r LeaveRequest
		err := cur.Decode(&r)
		if err1 != nil {
			log.Fatal("error while decoding resp", err)
		}
		fmt.Println("r=", r)
		results = append(results, r)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(AllLeave{
		Leaves: results,
	})

}

var SECRET = []byte("super-secret-auth-key")
var api_key = "1234"

func CreateJWT(id string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = id
	tokenStr, err := token.SignedString(SECRET)
	if err != nil {
		panic(err)
	}
	return tokenStr, nil

}
func validateJWT(next func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	fmt.Println("i'm in validate")

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("i'm in next")

		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(t *jwt.Token) (interface{}, error) {
				_, ok := t.Method.(*jwt.SigningMethodHMAC)
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("not authorized32133"))
				}
				return SECRET, nil
			})
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("not authorized" + err.Error()))
			}
			if token.Valid {
				data := token.Claims.(jwt.MapClaims)
				fmt.Println("token", data["id"])
				var id1 string
				value, ok := data["id"]
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("header error"))
				}
				id1 = value.(string)

				r.Header["id"] = []string{id1}
				next(w, r)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("not authorizedrfsdgfb"))
		}
	}

}

type getJWtReq struct {
	Id string `json:"id"`
}

type getJWtRes struct {
	Status  int    `json:"status"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

func getJWt(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Body======", r.Body)
	jwtReq := getJWtReq{}
	if err := json.NewDecoder(r.Body).Decode(&jwtReq); err != nil {
		json.NewEncoder(w).Encode(getJWtRes{
			Status:  400,
			Message: "error while decoding body",
		})
		return
	}

	token, err := CreateJWT(jwtReq.Id)
	if err != nil {
		json.NewEncoder(w).Encode(getJWtRes{
			Status:  400,
			Message: "error while creating jwt token",
		})
		return
	}

	json.NewEncoder(w).Encode(getJWtRes{
		Status: 200,
		Token:  token,
	})
}

func ApproveLeave(w http.ResponseWriter, r *http.Request) {
	fmt.Println("approve leave reached", r.Header["id"])
	q := mux.Vars(r)
	leaveid := q["_id"]

	fmt.Println("Leave ID ", leaveid)
	id := r.Header["id"][0]
	fmt.Println("User ID =====", id)
	filter := bson.M{
		"id": id,
	}
	cur, err1 := UserCollection.Find(context.TODO(), filter)
	if err1 != nil {
		log.Fatal("Error while getting user", err1)
	}
	for cur.Next(context.TODO()) {
		fmt.Println("Inside for loop get user")
		var r User
		err := cur.Decode(&r)
		if err1 != nil {
			log.Fatal("error while decoding resp", err)
		} else {
			fmt.Println("User===", r)
			if r.Role == Role(1) {
				oid, err := primitive.ObjectIDFromHex(leaveid)
				if err != nil {
					panic(err)
				}
				fmt.Println("role is 1", r.Role)
				filter := bson.M{
					"_id": oid,
				}
				fmt.Println(filter)

				update := bson.M{
					"$set": bson.M{"approved": true},
				}

				result, err := LeaveCollection.UpdateOne(context.Background(), filter, update)
				fmt.Println("result is", result)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("modified result is ", result.ModifiedCount)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(CheckStatus{
					Status:   200,
					Approved: true,
				})

			}
		}

	}
}
