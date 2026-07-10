package reddit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const listingJSON = `{"kind":"Listing","data":{"after":"t3_next","children":[
  {"kind":"t3","data":{"id":"a1","name":"t3_a1","title":"Hello","author":"alice","subreddit":"golang","score":42,"num_comments":7}},
  {"kind":"t3","data":{"id":"a2","name":"t3_a2","title":"World","author":"bob","subreddit":"golang","score":10,"num_comments":2}}
]}}`

const commentsJSON = `[
  {"kind":"Listing","data":{"children":[{"kind":"t3","data":{"id":"a1","title":"Root","author":"op"}}]}},
  {"kind":"Listing","data":{"children":[{"kind":"t1","data":{"id":"c1","author":"bob","body":"hi","score":3,"replies":""}}]}}
]`

func newSession(t *testing.T, h http.HandlerFunc) *Session {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewSession(WithBaseURL(srv.URL), WithUserAgent("test/1.0"))
}

func TestSubreddit(t *testing.T) {
	s := newSession(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/r/golang/top.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("limit = %q", r.URL.Query().Get("limit"))
		}
		w.Write([]byte(listingJSON))
	})
	posts, err := s.Subreddit(context.Background(), "golang", "top", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 {
		t.Fatalf("got %d posts", len(posts))
	}
	first := posts[0].(map[string]any)
	if first["title"] != "Hello" || first["author"] != "alice" {
		t.Errorf("first post = %v", first)
	}
	// Score decodes as a JSON number (float64).
	if first["score"].(float64) != 42 {
		t.Errorf("score = %v", first["score"])
	}
}

func TestFrontpage(t *testing.T) {
	s := newSession(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hot.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Write([]byte(listingJSON))
	})
	posts, err := s.Frontpage(context.Background(), "hot", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 {
		t.Fatalf("got %d posts", len(posts))
	}
}

func TestComments(t *testing.T) {
	s := newSession(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/r/golang/comments/a1.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Write([]byte(commentsJSON))
	})
	res, err := s.Comments(context.Background(), "golang", "a1")
	if err != nil {
		t.Fatal(err)
	}
	post := res["Post"].(map[string]any)
	if post["title"] != "Root" {
		t.Errorf("post = %v", post)
	}
	comments := res["Comments"].([]any)
	if len(comments) != 1 || comments[0].(map[string]any)["author"] != "bob" {
		t.Errorf("comments = %v", comments)
	}
}

func TestCall(t *testing.T) {
	s := newSession(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/r/golang/hot.json":
			w.Write([]byte(listingJSON))
		case r.URL.Path == "/best.json":
			w.Write([]byte(listingJSON))
		case r.URL.Path == "/r/golang/comments/a1.json":
			w.Write([]byte(commentsJSON))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	})
	ctx := context.Background()

	// subreddit with default sort (arg omitted) + int64 limit (as rbgo passes)
	got, err := s.Call(ctx, "subreddit", "golang", "", int64(25))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.([]any)) != 2 {
		t.Errorf("subreddit call = %v", got)
	}

	// alias "r"
	if _, err := s.Call(ctx, "R", "golang"); err != nil {
		t.Fatal(err)
	}

	// frontpage alias with explicit sort
	if _, err := s.Call(ctx, "home", "best"); err != nil {
		t.Fatal(err)
	}

	// comments
	tree, err := s.Call(ctx, "comments", "golang", "a1")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := tree.(map[string]any); !ok {
		t.Errorf("comments call type = %T", tree)
	}
}

func TestCallErrors(t *testing.T) {
	s := NewSession()
	ctx := context.Background()
	if _, err := s.Call(ctx, "subreddit"); err == nil {
		t.Error("subreddit without name should error")
	}
	if _, err := s.Call(ctx, "comments", "golang"); err == nil {
		t.Error("comments without id should error")
	}
	if _, err := s.Call(ctx, "nope"); err == nil {
		t.Error("unknown method should error")
	}
}

func TestPropagatesClientErrors(t *testing.T) {
	s := newSession(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	ctx := context.Background()
	if _, err := s.Subreddit(ctx, "golang", "hot", 0); err == nil {
		t.Error("want error on 403 (subreddit)")
	}
	if _, err := s.Frontpage(ctx, "hot", 0); err == nil {
		t.Error("want error on 403 (frontpage)")
	}
	if _, err := s.Comments(ctx, "golang", "a1"); err == nil {
		t.Error("want error on 403 (comments)")
	}
}

func TestClientAccessor(t *testing.T) {
	s := NewSession()
	if s.Client() == nil {
		t.Error("Client() nil")
	}
}

func TestCoercion(t *testing.T) {
	if toString("x") != "x" || toString(nil) != "" || toString(42) != "42" {
		t.Error("toString")
	}
	if toString(stringer{}) != "S" {
		t.Error("toString stringer")
	}
	if toInt(3) != 3 || toInt(int64(4)) != 4 || toInt(5.0) != 5 || toInt(nil) != 0 || toInt("x") != 0 {
		t.Error("toInt")
	}
	if argString(nil, 0, "def") != "def" || argString([]any{""}, 0, "def") != "def" || argString([]any{"a"}, 0, "def") != "a" {
		t.Error("argString")
	}
	if argInt(nil, 0, 9) != 9 || argInt([]any{int64(3)}, 0, 9) != 3 {
		t.Error("argInt")
	}
}

type stringer struct{}

func (stringer) String() string { return "S" }

func TestToArrayHashMarshalError(t *testing.T) {
	// A value that cannot be JSON-marshalled surfaces the error.
	if _, err := toArray(make(chan int)); err == nil {
		t.Error("toArray should error on unmarshalable input")
	}
	if _, err := toHash(make(chan int)); err == nil {
		t.Error("toHash should error on unmarshalable input")
	}
	// A JSON scalar cannot decode into a Hash/Array -> unmarshal error.
	if _, err := toArray(42); err == nil {
		t.Error("toArray(scalar) should error")
	}
	if _, err := toHash(42); err == nil {
		t.Error("toHash(scalar) should error")
	}
}
