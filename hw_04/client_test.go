package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const testToken = "112233"

type TestCase struct {
	Id      string
	Params  SearchRequest
	Result  []User
	IsError bool
}

func TestFindUsersParams(t *testing.T) {
	testCases := []TestCase{
		// [order_by] - success
		{Id: "Valid order_by (=1)", Params: SearchRequest{OrderBy: 1}, IsError: false},
		{Id: "Valid order_by (=0)", Params: SearchRequest{OrderBy: 0}, IsError: false},
		{Id: "Valid order_by (=-1)", Params: SearchRequest{OrderBy: -1}, IsError: false},
		// [order_by] - error
		{Id: "Invalid order_by (>1)", Params: SearchRequest{OrderBy: 2}, IsError: true},
		{Id: "Invalid order_by (<-1)", Params: SearchRequest{OrderBy: -2}, IsError: true},

		// [order_field] - success
		{Id: "Valid order_field (=Id)", Params: SearchRequest{OrderField: "Id", OrderBy: OrderByAsc}, IsError: false},
		{Id: "Valid order_field (=Age)", Params: SearchRequest{OrderField: "Age", OrderBy: OrderByAsc}, IsError: false},
		{Id: "Valid order_field (=Name)", Params: SearchRequest{OrderField: "Name", OrderBy: OrderByAsc}, IsError: false},
		{Id: "Valid order_field (='empty')", Params: SearchRequest{OrderField: "", OrderBy: OrderByAsc}, IsError: false},
		// [order_field] - error
		{Id: "Invalid order_field (=InvalidTestField)", Params: SearchRequest{OrderField: "InvalidTestField", OrderBy: OrderByAsc}, IsError: true},

		// [order_by] - error
		{Id: "Invalid order_by (=InvalidOrderBy)", Params: SearchRequest{OrderField: "InvalidOrderBy", OrderBy: OrderByAsc}, IsError: true},
	}
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	client := &SearchClient{
		URL:         server.URL,
		AccessToken: testToken,
	}
	defer server.Close()

	for _, tc := range testCases {
		_, err := client.FindUsers(tc.Params)

		if tc.IsError && err == nil {
			t.Errorf("[%s] expected error, got nil", tc.Id)
		}

		if !tc.IsError && err != nil {
			t.Errorf("[%s] unexpected error: %#v", tc.Id, err)
		}
	}
}
