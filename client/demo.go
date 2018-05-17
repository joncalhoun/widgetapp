package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

func main() {
	login := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "jon@calhoun.io",
		Password: "demo",
	}
	pr, pw := io.Pipe()
	enc := json.NewEncoder(pw)
	go func() {
		err := enc.Encode(login)
		if err != nil {
			log.Println("Error encoding login info:", err)
		}
		pw.Close()
	}()
	res, err := http.Post("http://localhost:3000/api/signin", "application/json", pr)
	if err != nil {
		panic(err)
	}
	// io.Copy(os.Stdout, res.Body)
	var t oauth2.Token
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&t)
	if err != nil {
		panic(err)
	}
	res.Body.Close()

	var conf oauth2.Config
	client := conf.Client(context.Background(), &t)
	pr, pw = io.Pipe()
	go func() {
		req := struct {
			Name  string `json:"name"`
			Color string `json:"color"`
			Price int    `json:"price"`
		}{
			Name:  "API widget",
			Color: "APrIcot",
			Price: 90210,
		}
		enc := json.NewEncoder(pw)
		enc.Encode(req)
		pw.Close()
	}()
	res, err = client.Post("http://localhost:3000/api/widgets", "application/json", pr)
	io.Copy(os.Stdout, res.Body)
	res.Body.Close()

	res, err = client.Get("http://localhost:3000/api/widgets")
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, res.Body)
	res.Body.Close()

}
