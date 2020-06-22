package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/likexian/whois-go"
	"github.com/valyala/fasthttp"
)

type endpointStruct struct {
	// Name      string
	IpAddress string
	Grade     string
	Country   string
	Owner     string
}

type Domain struct {
	Name               string `json:"domain"`
	Servers_changed    string `json:"servers_changed"`
	Ssl_grade          string `json:"ssl_grade"`
	Previous_ssl_grade string `json:"previous_ssl_grade"`
	Logo               string `json: "logo"`
	Title              string `json:"title"`
	Is_down            string `json: "is_down"`
	Endpoints          []endpointStruct
}
type data interface {
	getValues() []string
}

func (domain *Domain) getValues() []string {
	values := []string{domain.Name, "0", domain.Ssl_grade, "0", domain.Logo, domain.Title, "0"}
	return values
}

func (endpoint *endpointStruct) getValues() []string {
	values := []string{endpoint.IpAddress, endpoint.Grade, endpoint.Country, endpoint.Owner}
	return values
}

func values(d data) []string {
	return d.getValues()
}

//needs some work
func (domain *Domain) getLogo() string {
	name := (*domain).Name
	_, body, _ := fasthttp.Get(nil, "http://www."+name)
	// re := regexp.MustCompile(`<link.* (rel="shortcut icon" [^<]*)`)
	re := regexp.MustCompile(`<link[^>]*(rel="shortcut icon" [^<]*)`)
	result := re.FindString(string(body)) // finds the substring
	if len(result) == 0 {
		return "fakeicon.com"
	}
	re = regexp.MustCompile(`href="([^"]*)`)
	href := re.FindStringSubmatch(result)
	return href[1]
}

// gets the Title of the website using regex on the html
func (domain *Domain) getTitle() string { // domain is truora.com
	name := (*domain).Name
	_, body, err := fasthttp.Get(nil, "http://www."+name)
	if err != nil {
		fmt.Println(err)
		return "Faketitle"
	}
	re := regexp.MustCompile(`<title.*>[\s\S]*<\/title>`)
	result := re.FindString(string(body)) // finds the substring
	i := strings.Index(result, ">")
	if i < 0 {
		return "N/A"
	}
	title := result[i+1 : len(result)-8]
	return title

}

//gets Owner of IPaddress using regex
func getOwner(text string) string {
	// get owner
	// re := regexp.MustCompile(`OrgName:.*OrgId`)
	re := regexp.MustCompile(`[Oo]rg[-]?[Nn]ame:([^\:]*)`) // split the sentence and remove the last word
	result := re.FindString(text)
	orgName := strings.Split(result, " ")
	if len(orgName) <= 1 {
		return "N/A"
	}
	orgName = orgName[1 : len(orgName)-1]
	result = strings.Join(orgName, " ")
	return result
}

//gets Country of IPaddress
func getCountry(text string) string {
	// get country
	re := regexp.MustCompile(`[cC]ountry: [A-Za-z]+`)
	country_name := re.FindString(text)
	if country_name == "" {
		return "N/A"
	}
	country_name = country_name[9:] // trial and error
	return country_name
}

//Calculates lowest SSL grade from the endpoints
func (domain *Domain) getSSLGrade() string {
	lowest := "A+"
	for _, endpoint := range domain.Endpoints {
		if lowest == "A+" && endpoint.Grade == "A" { // asci values call for this exception
			lowest = "A"
			continue
		}
		if endpoint.Grade > lowest { // B > A, lowers SSL but higher ASCII
			lowest = endpoint.Grade
		}
	}
	return lowest
}

// returns the address and grade of each server
func (domain *Domain) getEndpoints() []endpointStruct {
	name := (*domain).Name
	_, body, err := fasthttp.Get(nil, "https://api.ssllabs.com/api/v3/analyze?host="+name)
	if err != nil {
		log.Fatal(err)
	}

	s := string(body)
	i := strings.Index(s, "\"endpoints\"")
	newS := s[i+13 : len(s)-3] // remove trailing brackets
	indEndpoint := strings.Split(newS, "},")

	//initialize slice of endpoints
	endpoints := []endpointStruct{}

	// parse string and convert to endpoint struct
	for i := range indEndpoint {
		indEndpoint[i] = indEndpoint[i] + "}"
		data := endpointStruct{}
		json.Unmarshal([]byte(indEndpoint[i]), &data)
		// data.Name = domain.Name
		endpoints = append(endpoints, data)
	}

	// get country and owner of each endpoint server
	for i, endpoint := range endpoints {
		a := endpoint.IpAddress
		result, err := whois.Whois(a)
		if err == nil {
			// clean result whitespace
			space := regexp.MustCompile(`\s+`)
			result_nospace := space.ReplaceAllString(result, " ")

			endpoints[i].Owner = getOwner(result_nospace)
			endpoints[i].Country = getCountry(result_nospace)
		}

	}
	return endpoints

}

// to get an icon, look in the head for /favicon.ico or shortcut icon and get the url. golangs function .index can help me find the location of a given string

func populate(url string) Domain {
	domain := Domain{Name: url}
	domain.Logo = domain.getLogo()
	domain.Title = domain.getTitle()
	domain.Endpoints = domain.getEndpoints()
	domain.Ssl_grade = domain.getSSLGrade()
	fmt.Println(domain)
	return domain
}
