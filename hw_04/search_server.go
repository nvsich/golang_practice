package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	datasource = "dataset.xml"
)

type XmlData struct {
	XMLName xml.Name `xml:"root"`
	Rows    []Row    `xml:"row"`
}

type Row struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// Reading xml file
	xmlFile, err := os.Open(datasource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(xmlFile *os.File) {
		err := xmlFile.Close()
		if err != nil {
			panic("failed to close xml file")
		}
	}(xmlFile)

	// Parsing xml data
	var xmlData XmlData
	all, err := io.ReadAll(xmlFile)
	err = xml.Unmarshal(all, &xmlData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filtering data by query
	var users []User
	q := r.URL.Query()
	query := q.Get("query")

	for _, row := range xmlData.Rows {
		isMatching := strings.Contains(row.FirstName, query) ||
			strings.Contains(row.LastName, query) ||
			strings.Contains(row.About, query)

		if isMatching || query == "" {
			users = append(users, User{
				Id:     row.Id,
				Name:   row.FirstName + " " + row.LastName,
				Age:    row.Age,
				Gender: row.Gender,
				About:  row.About,
			})
		}
	}

	// Sorting
	orderBy, err := strconv.Atoi(q.Get("order_by"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !(orderBy == OrderByAsIs || orderBy == OrderByAsc || orderBy == OrderByDesc) {
		http.Error(w, ErrorBadOrderField, http.StatusBadRequest)
		return
	}

	if orderBy != OrderByAsIs {
		orderField := q.Get("order_field")

		var orderFunc func(user1, user2 User) bool
		switch orderField {
		case "Id":
			orderFunc = func(user1, user2 User) bool {
				return user1.Id <= user2.Id
			}
		case "Age":
			orderFunc = func(user1, user2 User) bool {
				return user1.Age <= user2.Age
			}
		case "Name":
			orderFunc = func(user1, user2 User) bool {
				return user1.Name <= user2.Name
			}
		case "":
			orderFunc = func(user1, user2 User) bool {
				return user1.Id <= user2.Id
			}
		default:
			{
				http.Error(w, ErrorBadOrderField, http.StatusBadRequest)
				return
			}
		}

		sort.Slice(users, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return orderFunc(users[i], users[j])
			} else {
				return !orderFunc(users[j], users[i])
			}
		})
	}

	// Limit
	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if (limit > 0) && (limit <= len(users)) {
		users = users[:limit]
	}

	// Offset
	offset, err := strconv.Atoi(q.Get("offset"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if (offset > 0) && (offset <= len(users)) {
		users = users[offset:]
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
