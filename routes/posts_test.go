package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type PostWithoutId struct {
	UserId int
	Title  string
	Body   string
}

type Post struct {
	Id     int
	UserId int
	Title  string
	Body   string
}

type JsonPlaceholderMock struct{}

// mock function creates some dummy data and encodes to JSON via json.Marshal
func (*JsonPlaceholderMock) GetPosts() (*http.Response, error) {
	mockedPosts := []Post{{
		Id:     1,
		UserId: 2,
		Title:  "Hello World",
		Body:   "Foo Bar",
	}}

	respBody, err := json.Marshal(mockedPosts)
	if err != nil {
		log.Panicf("Error reading mocked response data: %v", err)
	}

	// then returns a minimal HTTP response with status code and body
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
	}, nil
}

func TestGetPostsHandler(t *testing.T) {
	// set GetPosts package-scoped variable to the mock function
	GetPosts = (&JsonPlaceholderMock{}).GetPosts

	// create a new GET request to send to /posts
	req, err := http.NewRequest("GET", "/posts", nil)
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	// NewRecorder records the ResponseWriter's mutations
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PostsResource{}.List)
	// call the handler with the response recorder rr and created request req
	handler.ServeHTTP(rr, req)
	// if any error encountered, fail the test
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusOK, status)
	}

	// decode body and store result in variable
	var posts []Post

	if err := json.NewDecoder(rr.Body).Decode(&posts); err != nil {
		t.Errorf("Error decoding response body: %v", err)
	}

	resultTotal := len(posts)
	expectedTotal := 1

	if resultTotal != expectedTotal {
		t.Errorf("Expected: %d. Got: %d.", expectedTotal, resultTotal)
	}
}

// mock the CreatePost function to avoid sending network request
func (*JsonPlaceholderMock) CreatePost(body io.ReadCloser) (*http.Response, error) {
	// body needs to contain ID, title, body of text
	// since request body must be passed as type io.ReadCloser to CreatePost, it must be read into a buffer, converted into a byte slice and then decoded into a Go struct so it can be accessed normally
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(body)

	var reqPost PostWithoutId
	if err := json.Unmarshal(buffer.Bytes(), &reqPost); err != nil {
		log.Panicf("Error decoding request body: %v", err)
	}

	// when POST /posts request sent to JSONPlaceholder API, it returns a response that contains the newly created post

	newPost := Post{
		Id:     101,
		UserId: reqPost.UserId,
		Title:  reqPost.Title,
		Body:   reqPost.Body,
	}

	// encode newPost to JSON via json.Marshal
	respBody, err := json.Marshal(newPost)
	if err != nil {
		log.Panicf("Error reading mocked response data: %v", err)
	}

	// HTTP response should return with a 200 status to indicate success
	// nopCloser returns a ReadCloser that wraps the Reader (in this case, the bytes.NewBuffer(respBody), which prepares a buffer to read respBody) with a no-op Close method, which allows the Reader to adhere to the ReadCloser interface
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
	}, nil
}

// write test for CreatePostHandler, but with adjustments for POST
func TestCreatePost(t *testing.T) {
	// set CreatePost package-scoped variable to mock function above
	CreatePost = (&JsonPlaceholderMock{}).CreatePost

	// init postWithoutId to a PostWithoutId struct
	postWithoutId := PostWithoutId{
		UserId: 1,
		Title:  "Hello World",
		Body:   "Foo Bar",
	}

	// encode postWithoutId to JSON via json.Marshal
	reqBody, err := json.Marshal(postWithoutId)
	if err != nil {
		log.Panicf("Error reading mocked request data: %v", err)
	}

	// create a new POST request via http.NewRequest
	// because NewRequest accepts a body of type io.Reader, we must convert reqBody which is currently a byte slice to a type compatible with Reader interface, which implements a single method Read
	// byte.NewBuffer creates and initalises a new Buffer using byte slice argument as its initial contents.
	// Buffer type also has Read and Write methods, which matches io.Reader interface
	// therefore, to pass reqBody to http.NewRequest as the POST request body, we must first pass it to bytes.NewBuffer and then pass the returned buffer as the request's body.
	// request will finally then be passed to the route handler
	req, err := http.NewRequest("POST", "/posts", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Errorf("Error creating a new request: %v", err)
	}

	// once the request is successfully created, set its header Content-Type to application/json, which specifies the body as JSON
	req.Header.Set("Content-Type", "application/json")

	// like TestGetPostHandler, setup response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PostsResource{}.Create)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusOK, status)
	}

	// decode body of response and store at memory address of post variable
	var post Post

	if err := json.NewDecoder(rr.Body).Decode(&post); err != nil {
		t.Errorf("Error decoding response body: %v", err)
	}

	resultId := post.Id
	expectedId := 101

	if resultId != expectedId {
		t.Errorf("Expected: %d. Got: %d.", expectedId, resultId)
	}
}
