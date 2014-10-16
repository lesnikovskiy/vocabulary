// To use the server install mgo.v2 driver for MongoDB
// go get gopkg.in/mgo.v2
// go get github.com/gorilla/mux
// To use JSON Web Tokens use 'go get github.com/dgrijalva/jwt-go'
package main

import (
	"encoding/json"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	DbServer   = "mongodb://localhost:8000"
	Database   = "vocabulary"
	Collection = "entries"
)

var (
	privateKey []byte
	publicKey  []byte
)

type Entry struct {
	Id          bson.ObjectId `bson:"_id"`
	Word        string        `bson:"word"`
	Translation string        `bson:"translation"`
}

func init() {
	privateKey, _ = ioutil.ReadFile("./demo.rsa")
	publicKey, _ = ioutil.ReadFile("./demo.rsa.pub")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/entry/", entriesGet).Methods("GET")
	r.HandleFunc("/api/entry/", entriesPost).Methods("POST")
	r.HandleFunc("/api/entry/{id}", entriesDelete).Methods("DELETE")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func entriesGet(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, ":", r.RequestURI)

	// Generate JSON Web Token and set cookie with a token.
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims["ID"] = "2134asdf43451dscfds32423ASDF"
	token.Claims["exp"] = time.Now().Unix() + 36000

	tokenString, _ := token.SignedString(privateKey)
	expires := time.Now().AddDate(0, 0, 1)
	cookie := http.Cookie{Name: "token", Value: tokenString, Path: "/", Expires: expires, RawExpires: expires.Format(time.UnixDate), HttpOnly: true}
	http.SetCookie(w, &cookie)

	session, err := mgo.Dial(DbServer)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()

	session.SetSafe(&mgo.Safe{})
	collection := session.DB(Database).C(Collection)
	var results []Entry
	err = collection.Find(bson.M{}).All(&results)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(results)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func entriesPost(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, ":", r.RequestURI)

	// Now validate token
	cookie, _ := r.Cookie("token")
	token, _ := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var entry Entry
	if err = json.Unmarshal(body, &entry); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	entry.Id = bson.NewObjectId()

	session, err := mgo.Dial(DbServer)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()

	session.SetSafe(&mgo.Safe{})
	collection := session.DB(Database).C(Collection)
	if err = collection.Insert(&entry); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Created"))
}

func entriesDelete(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, ":", r.RequestURI)
	_id := mux.Vars(r)["id"]
	log.Println(_id)

	session, err := mgo.Dial(DbServer)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()

	session.SetSafe(&mgo.Safe{})
	if err = session.DB(Database).C(Collection).RemoveId(bson.ObjectIdHex(_id)); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Accepted"))
}
