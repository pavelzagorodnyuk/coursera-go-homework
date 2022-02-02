package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const AccessToken string = "@ItIsToken"

// TestFindUsersConnection checks connection between FindUsers and SearchServer
func TestFindUsersConnection(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(SearchServer))

	var request = SearchRequest{
		Limit:      5, // limit must be <= 25 for correct work the test
		Offset:     0,
		Query:      "A",
		OrderField: "Name",
		OrderBy:    OrderByAsc,
	}

	var params = url.Values{}
	params.Add("limit", strconv.Itoa(request.Limit+1))
	params.Add("offset", strconv.Itoa(request.Offset))
	params.Add("query", request.Query)
	params.Add("order_field", request.OrderField)
	params.Add("order_by", strconv.Itoa(request.OrderBy))
	_, errorDescription := http.Get("?" + params.Encode())

	var tests = []struct {
		client           SearchClient
		responseIsNil    bool
		errorIsNil       bool
		errorDescription string
	}{
		{
			client: SearchClient{
				AccessToken: AccessToken,
				URL:         server.URL,
			},
			responseIsNil: false,
			errorIsNil:    true,
		},
		{
			client: SearchClient{
				AccessToken: "!" + AccessToken,
				URL:         server.URL,
			},
			responseIsNil:    true,
			errorIsNil:       false,
			errorDescription: "Bad AccessToken",
		},
		{
			client: SearchClient{
				AccessToken: AccessToken,
				URL:         "",
			},
			responseIsNil:    true,
			errorIsNil:       false,
			errorDescription: "unknown error " + errorDescription.Error(),
		},
	}

	for index := range tests {
		response, err := tests[index].client.FindUsers(request)

		if tests[index].responseIsNil && response != nil ||
			!tests[index].responseIsNil && response == nil ||
			tests[index].errorIsNil && err != nil ||
			!tests[index].errorIsNil && err == nil {

			t.Errorf("[test #%d] got unexpected result\n\tresponse [expected: %t]: %v\n\terror [expected: %t]: %v",
				index, !tests[index].responseIsNil, response, !tests[index].errorIsNil, err)
		}

		if !tests[index].errorIsNil && err != nil && err.Error() != tests[index].errorDescription {
			t.Errorf("[test #%d] got unexpected error\n\texpected: %s\n\tgot: %s\n\t",
				index, tests[index].errorDescription, err)
		}
	}

}

// RequestsTests are tests for a TestFindUsersRequests
var RequestsTests = []struct {
	request  SearchRequest
	response struct {
		usersNumber int
		nextPage    bool
	}
	errorDescription string
	responseIsNil    bool
	errorIsNil       bool
}{
	{
		request: SearchRequest{
			Limit:   5,
			Query:   "Rose Carney",
			OrderBy: OrderByAsIs,
		},
		response: struct {
			usersNumber int
			nextPage    bool
		}{
			usersNumber: 1,
			nextPage:    false,
		},
		responseIsNil: false,
		errorIsNil:    true,
	},
	{
		request: SearchRequest{
			Limit:      30,
			OrderField: "Id",
			OrderBy:    OrderByAsc,
		},
		response: struct {
			usersNumber int
			nextPage    bool
		}{
			usersNumber: 25,
			nextPage:    true,
		},
		responseIsNil: false,
		errorIsNil:    true,
	},
	{
		request: SearchRequest{
			Limit:      15,
			OrderField: "Id",
			OrderBy:    OrderByAsc,
		},
		response: struct {
			usersNumber int
			nextPage    bool
		}{
			usersNumber: 15,
			nextPage:    true,
		},
		responseIsNil: false,
		errorIsNil:    true,
	},
	{
		request: SearchRequest{
			Limit:      30,
			OrderField: "Id",
			OrderBy:    OrderByAsc,
		},
		response: struct {
			usersNumber int
			nextPage    bool
		}{
			usersNumber: 25,
			nextPage:    true,
		},
		responseIsNil: false,
		errorIsNil:    true,
	},
	{
		request: SearchRequest{
			Limit:   -1,
			OrderBy: OrderByAsIs,
		},
		errorDescription: "limit must be >= 0",
		responseIsNil:    true,
		errorIsNil:       false,
	},
	{
		request: SearchRequest{
			Offset:  -1,
			OrderBy: OrderByAsIs,
		},
		errorDescription: "offset must be >= 0",
		responseIsNil:    true,
		errorIsNil:       false,
	},
}

// TestFindUsersRequests checks behaviour FindUsers for different requests
func TestFindUsersRequests(t *testing.T) {

	var server = httptest.NewServer(http.HandlerFunc(SearchServer))
	var client = SearchClient{
		AccessToken: AccessToken,
		URL:         server.URL,
	}

	for index := range RequestsTests {
		response, err := client.FindUsers(RequestsTests[index].request)

		if RequestsTests[index].responseIsNil && response != nil ||
			!RequestsTests[index].responseIsNil && response == nil ||
			RequestsTests[index].errorIsNil && err != nil ||
			!RequestsTests[index].errorIsNil && err == nil {

			t.Errorf("[test #%d] got unexpected result\n\tresponse [expected: %t]: %v\n\terror [expected: %t]: %v",
				index, !RequestsTests[index].responseIsNil, response, !RequestsTests[index].errorIsNil, err)
			continue
		}

		if !RequestsTests[index].responseIsNil && len(response.Users) != RequestsTests[index].response.usersNumber {
			t.Errorf("[test #%d] the number of users in the received response differs from the expected one\n"+
				"\tExpected: %d\n\tGot: %d", index, RequestsTests[index].response.usersNumber, len(response.Users))
		}

		if !RequestsTests[index].responseIsNil && response.NextPage != RequestsTests[index].response.nextPage {
			t.Errorf("[test #%d] the NextPage parameter differs from the expected one\n"+
				"\tExpected: \"%t\"\n\tGot: \"%t\"", index, RequestsTests[index].response.nextPage, response.NextPage)
		}

		if !RequestsTests[index].errorIsNil && RequestsTests[index].errorDescription != err.Error() {
			t.Errorf("[test #%d] the response error differs from the expected one\n"+
				"\tExpected: %s\n\tGot: %s", index, RequestsTests[index].errorDescription, err.Error())
		}
	}
}

// ServerTests are tests for a TestFindUsersServer
var ServerTests = []struct {
	query         string
	errorPrefix   string
	responseIsNil bool
	errorIsNil    bool
}{
	{
		query:         "__correct",
		responseIsNil: false,
		errorIsNil:    true,
	},
	{
		query:         "__timeout",
		errorPrefix:   "timeout for",
		responseIsNil: true,
		errorIsNil:    false,
	},
	{
		query:         "__internal_server_error",
		errorPrefix:   "SearchServer fatal error",
		responseIsNil: true,
		errorIsNil:    false,
	},
	{
		query:         "__bad_request",
		errorPrefix:   "unknown bad request error:",
		responseIsNil: true,
		errorIsNil:    false,
	},
	{
		query:         "__bad_request_broken_json",
		errorPrefix:   "cant unpack error json:",
		responseIsNil: true,
		errorIsNil:    false,
	},
	{
		query:         "__bad_request_bad_order_field",
		errorPrefix:   "OrderFeld",
		responseIsNil: true,
		errorIsNil:    false,
	},
	{
		query:         "__broken_json",
		errorPrefix:   "cant unpack result json:",
		responseIsNil: true,
		errorIsNil:    false,
	},
}

// TestFindUsersServer checks behaviour FindUsers for case when server give errors
func TestFindUsersServer(t *testing.T) {

	var server = httptest.NewServer(http.HandlerFunc(MistakeSearchServer))
	var client = SearchClient{
		AccessToken: AccessToken,
		URL:         server.URL,
	}

	for index := range ServerTests {
		response, err := client.FindUsers(SearchRequest{
			Query: ServerTests[index].query,
		})

		if ServerTests[index].responseIsNil && response != nil ||
			!ServerTests[index].responseIsNil && response == nil ||
			ServerTests[index].errorIsNil && err != nil ||
			!ServerTests[index].errorIsNil && err == nil {

			t.Errorf("[test #%d] got unexpected result\n\tresponse [expected: %t]: %v\n\terror [expected: %t]: %v",
				index, !ServerTests[index].responseIsNil, response, !ServerTests[index].errorIsNil, err)
			continue
		}

		if !ServerTests[index].errorIsNil && !strings.HasPrefix(err.Error(), ServerTests[index].errorPrefix) {
			t.Errorf("[test #%d] the response error differs from the expected one\n"+
				"\tExpected prefix of error: %s\n\tGot: %s", index, ServerTests[index].errorPrefix, err.Error())
		}
	}
}

// SearchServer serves requests of FindUsers
func SearchServer(response http.ResponseWriter, request *http.Request) {

	// checking a HTTP method and an access token
	if request.Method != http.MethodGet {
		response.WriteHeader(http.StatusBadRequest)
		response.Header().Add("Allow", "GET")
		response.Write([]byte("Request method is not valid. Only the GET method is available."))
		return
	}

	clientToken := request.Header.Get("AccessToken")
	if clientToken != AccessToken {
		response.WriteHeader(http.StatusUnauthorized)
		return
	}

	var (
		search SearchRequest
		err    error
		values url.Values = request.URL.Query()
	)

	// getting a SearchRequest from a request URL
	search.Limit, err = strconv.Atoi(values.Get("limit"))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("Value of Limit parameter is not valid"))
		return
	}
	if search.Limit < 0 {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("Value of Limit parameter must be >= 0"))
		return
	}

	search.Offset, err = strconv.Atoi(values.Get("offset"))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("Value of Offset parameter is not valid"))
		return
	}
	if search.Offset < 0 {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("Value of Offset parameter must be >= 0"))
		return
	}

	search.Query = values.Get("query")

	search.OrderField = values.Get("order_field")
	switch search.OrderField {
	case "Id", "Age", "Name":
		break
	case "":
		search.OrderField = "Name"
	default:
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("ErrorBadOrderField"))
		return
	}

	search.OrderBy, err = strconv.Atoi(values.Get("order_by"))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("Value of OrderBy parameter is not valid"))
		return
	}
	if search.OrderBy < -1 || search.OrderBy > 1 {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("Value of OrderBy parameter must be OrderByAsIs, OrderByAsc or OrderByDesc"))
		return
	}

	// executing the SearchRequest
	searchResponse, err := DoRequest(&search)
	if err != nil {
		if err == BadRequest {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte(err.Error()))
			return
		}
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// sending a response
	marshalJSON, err := json.Marshal(searchResponse.Users)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
	response.Header().Add("Content-Type", "application/json; charset=UTF-8")
	response.Header().Add("Content-Length", strconv.Itoa(len(marshalJSON)))
	response.Write(marshalJSON)
}

func (user *User) UnmarshalXML(decoder *xml.Decoder, token xml.StartElement) error {

	var middle struct {
		Id        int    `xml:"id"`
		FirstName string `xml:"first_name"`
		LastName  string `xml:"last_name"`
		Age       int    `xml:"age"`
		About     string `xml:"about"`
		Gender    string `xml:"gender"`
	}

	if err := decoder.DecodeElement(&middle, &token); err != nil {
		return err
	}

	user.Id = middle.Id
	user.Name = strings.Join([]string{middle.FirstName, middle.LastName}, " ")
	user.Age = middle.Age
	user.About = middle.About
	user.Gender = middle.Gender

	return nil
}

type usersSort struct {
	users      []User
	orderField string
	descOrder  bool
}

func (s *usersSort) Len() int {
	return len(s.users)
}

func (s *usersSort) Swap(i, j int) {
	s.users[i], s.users[j] = s.users[j], s.users[i]
}

func (s *usersSort) Less(i, j int) bool {
	switch s.orderField {
	case "Id":
		return !s.descOrder && s.users[i].Id < s.users[j].Id ||
			s.descOrder && s.users[i].Id > s.users[j].Id
	case "Age":
		if s.users[i].Age == s.users[j].Age {
			return !s.descOrder && s.users[i].Id < s.users[j].Id ||
				s.descOrder && s.users[i].Id > s.users[j].Id
		} else {
			return !s.descOrder && s.users[i].Age < s.users[j].Age ||
				s.descOrder && s.users[i].Age > s.users[j].Age
		}
	case "Name":
		if s.users[i].Name == s.users[j].Name {
			return !s.descOrder && s.users[i].Id < s.users[j].Id ||
				s.descOrder && s.users[i].Id > s.users[j].Id
		} else {
			return !s.descOrder && s.users[i].Name < s.users[j].Name ||
				s.descOrder && s.users[i].Name > s.users[j].Name
		}
	}
	return true
}

var BadRequest = errors.New("SearchRequest is not valid")

func DoRequest(request *SearchRequest) (*SearchResponse, error) {

	// checking a request
	if request == nil || request.Limit < 0 || request.Offset < 0 ||
		request.OrderBy != OrderByAsIs && request.OrderBy != OrderByAsc && request.OrderBy != OrderByDesc {
		return nil, BadRequest
	}

	dataset, err := os.Open("./dataset.xml")
	if err != nil {
		return nil, err
	}
	defer dataset.Close()

	// reading user data
	decoder := xml.NewDecoder(dataset)
	users := make([]User, 0)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "row" {
				var user User
				err = decoder.DecodeElement(&user, &element)
				if err != nil {
					return nil, err
				}

				if len(request.Query) == 0 {
					users = append(users, user)
				} else if strings.Contains(user.Name, request.Query) || strings.Contains(user.About, request.Query) {
					users = append(users, user)
				}
			}
		}
	}

	// sorting records of users
	if request.OrderBy != OrderByAsIs {

		var srt = usersSort{users: users}

		switch request.OrderField {
		case "Id", "Name", "Age":
			srt.orderField = request.OrderField
		case "":
			srt.orderField = "Name"
		default:
			return nil, BadRequest
		}

		if request.OrderBy == OrderByDesc {
			srt.descOrder = true
		}

		sort.Sort(&srt)
	}

	// making a response
	if request.Offset >= len(users) {
		return &SearchResponse{Users: []User{}, NextPage: false}, nil
	} else if request.Offset+request.Limit >= len(users) {
		return &SearchResponse{Users: users[request.Offset:], NextPage: false}, nil
	} else {
		return &SearchResponse{Users: users[request.Offset : request.Offset+request.Limit], NextPage: true}, nil
	}
}

// MistakeSearchServer is an alternative to SearchServer for TestFindUsersServer.
// MistakeSearchServer gives errors to test the FindUsers.
func MistakeSearchServer(response http.ResponseWriter, request *http.Request) {

	values := request.URL.Query()
	query := values.Get("query")

	switch query {
	case "__correct":
		response.WriteHeader(http.StatusOK)
		body, _ := json.Marshal([]User{{
			Id:     0,
			Name:   "Leann Travis",
			Age:    34,
			About:  "Lorem magna dolore et velit ut officia. Cupidatat deserunt elit mollit amet nulla voluptate sit. Quis aute aliquip officia deserunt sint sint nisi. Laboris sit et ea dolore consequat laboris non. Consequat do enim excepteur qui mollit consectetur eiusmod laborum ut duis mollit dolor est. Excepteur amet duis enim laborum aliqua nulla ea minim.",
			Gender: "female",
		}})
		response.Write(body)

	case "__timeout":
		time.Sleep(5 * time.Second)

	case "__internal_server_error":
		response.WriteHeader(http.StatusInternalServerError)

	case "__bad_request":
		var jsonError = `{"error": "ErrorBadRequest"}`
		response.WriteHeader(http.StatusBadRequest)
		io.WriteString(response, jsonError)

	case "__bad_request_broken_json":
		var jsonError = `{error: ErrorBadRequest}` // broken json
		response.WriteHeader(http.StatusBadRequest)
		io.WriteString(response, jsonError)

	case "__bad_request_bad_order_field":
		var jsonError = `{"error": "ErrorBadOrderField"}`
		response.WriteHeader(http.StatusBadRequest)
		io.WriteString(response, jsonError)

	case "__broken_json":
		io.WriteString(response, "`[{limit: 1}]`") // broken json

	default:
		response.WriteHeader(http.StatusBadRequest)
	}
}
