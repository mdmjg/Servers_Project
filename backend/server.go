package main

// structs for json data
import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
)

type Resp struct {
	Name            string `json:"name"`
	Servers_changed string `json:"servers_changed"`
	Ssl_grade       string `json:"ssl_grade"`
	Logo            string `json:"logo"`
}

func getDomain(ctx *fasthttp.RequestCtx) {
	fmt.Println("entering")
	var d Domain
	fmt.Printf(string(ctx.PostBody()))
	json.Unmarshal(ctx.PostBody(), &d)

	fmt.Println(d)
	d = populate(d.Name)

	//insert data into database
	insertDomain(d)

	// get updated values for servers
	d.Servers_changed, d.Ssl_grade, d.Previous_ssl_grade = fetchSSL(d.Name)

	// set response headers
	ctx.SetContentType("application/json; charset=UTF-8")
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.SetStatusCode(fasthttp.StatusOK)

	//return the response
	writer := json.NewEncoder(ctx.Response.BodyWriter())
	err := writer.Encode(d)
	if err != nil {
		panic(err)
	}
}

func listDomains(ctx *fasthttp.RequestCtx) {
	fmt.Println("helloo")
	// set response headers
	ctx.SetContentType("application/json; charset=UTF-8")
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.SetStatusCode(fasthttp.StatusOK)
	allDomains := fetchAll()
	fmt.Println(allDomains)

	//return the response
	writer := json.NewEncoder(ctx.Response.BodyWriter())
	err := writer.Encode(allDomains)
	if err != nil {
		panic(err)
	}
}

func main() {
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/listDomains":
			listDomains(ctx)
		case "/getDomain":
			getDomain(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}
	// router.GET("/go/:route", Test)
	// router.GET("/hello/:name", listDomains)
	// router.POST("/getDomain", getDomain)
	fmt.Println("server starting")

	log.Fatal(fasthttp.ListenAndServe(":8000", requestHandler))

}
