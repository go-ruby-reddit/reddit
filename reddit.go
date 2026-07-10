// Package reddit is the pure-Go, Ruby-runtime-independent core of the Ruby
// `reddit` gem: a reader-oriented client for Reddit's public API, shaped so
// that github.com/go-embedded-ruby/ruby can bind it as `require "reddit"`.
//
// It is a thin, reflective adapter over the typed client in
// github.com/go-reddit/reddit. A Session exposes the underlying client's
// operations through a single dynamic entry point, Call, which maps a
// Ruby-style snake_case method name ("subreddit", "frontpage", "comments") to
// the corresponding Go call, coerces the Ruby arguments (strings, integers and
// option Hashes), and normalises the result into Ruby-shaped values — a Hash
// (map[string]any), an Array ([]any) or a scalar. That uniform surface is what
// an rbgo binding drives from method_missing; nothing here depends on the Ruby
// runtime, so it is equally usable as a standalone Go library.
//
//	s := reddit.NewSession(reddit.WithUserAgent("app/1.0 (by /u/you)"))
//	posts, err := s.Call(ctx, "subreddit", "golang", "hot", 25) // Array of Hashes
//	tree,  err := s.Call(ctx, "comments", "golang", "abc123")   // Hash
package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	godreddit "github.com/go-reddit/reddit"
)

// Option configures the underlying client. Re-exported from go-reddit/reddit
// so callers (and the rbgo binding) need not import both packages.
type Option = godreddit.Option

// Re-exported client options.
var (
	WithUserAgent  = godreddit.WithUserAgent
	WithOAuth      = godreddit.WithOAuth
	WithOAuthScript = godreddit.WithOAuthScript
	WithHTTPClient = godreddit.WithHTTPClient
	WithBaseURL    = godreddit.WithBaseURL
)

// Session is a Ruby-facing handle over a go-reddit client.
type Session struct {
	c *godreddit.Client
}

// NewSession builds a Session pointed at the public API by default.
func NewSession(opts ...Option) *Session {
	return &Session{c: godreddit.NewClient(opts...)}
}

// Client exposes the underlying typed client for callers that want it.
func (s *Session) Client() *godreddit.Client { return s.c }

// Subreddit returns a page of posts from r/<name> as a Ruby Array of Hashes.
// sort defaults to "hot"; limit <= 0 uses Reddit's default page size.
func (s *Session) Subreddit(ctx context.Context, name, sort string, limit int) ([]any, error) {
	page, err := s.c.Subreddit(ctx, name, godreddit.Sort(sort), godreddit.ListingOptions{Limit: limit})
	if err != nil {
		return nil, err
	}
	return toArray(page.Posts)
}

// Frontpage returns the logged-out front page as a Ruby Array of Hashes.
func (s *Session) Frontpage(ctx context.Context, sort string, limit int) ([]any, error) {
	page, err := s.c.Frontpage(ctx, godreddit.Sort(sort), godreddit.ListingOptions{Limit: limit})
	if err != nil {
		return nil, err
	}
	return toArray(page.Posts)
}

// Comments returns a post and its comment tree as a Ruby Hash with "post" and
// "comments" keys.
func (s *Session) Comments(ctx context.Context, subreddit, id string) (map[string]any, error) {
	res, err := s.c.Comments(ctx, subreddit, id, godreddit.ListingOptions{Limit: 100})
	if err != nil {
		return nil, err
	}
	return toHash(res)
}

// Call is the dynamic dispatch entry point an rbgo binding uses. It routes a
// snake_case method name to the matching Session operation and coerces the
// Ruby-supplied arguments. Unknown methods return an error naming the method.
func (s *Session) Call(ctx context.Context, method string, args ...any) (any, error) {
	switch strings.ToLower(method) {
	case "subreddit", "r":
		if len(args) < 1 {
			return nil, fmt.Errorf("reddit: %q requires a subreddit name", method)
		}
		return s.Subreddit(ctx, toString(args[0]), argString(args, 1, "hot"), argInt(args, 2, 0))
	case "frontpage", "front", "home":
		return s.Frontpage(ctx, argString(args, 0, "hot"), argInt(args, 1, 0))
	case "comments", "post":
		if len(args) < 2 {
			return nil, fmt.Errorf("reddit: %q requires a subreddit and post id", method)
		}
		return s.Comments(ctx, toString(args[0]), toString(args[1]))
	default:
		return nil, fmt.Errorf("reddit: unknown method %q", method)
	}
}

// --- Ruby value coercion ----------------------------------------------------

// toArray normalises a typed slice into a Ruby Array ([]any) of Hashes via a
// JSON round-trip, matching the field names the client exposes.
func toArray(v any) ([]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out []any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// toHash normalises a typed value into a Ruby Hash (map[string]any).
func toHash(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// toString coerces a Ruby-supplied argument to a string.
func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

// toInt coerces a Ruby-supplied numeric argument (int64/float64/int/string) to
// an int, returning 0 when it cannot be interpreted.
func toInt(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case nil:
		return 0
	default:
		return 0
	}
}

// argString returns the i-th arg as a string, or def when absent/empty.
func argString(args []any, i int, def string) string {
	if i >= len(args) {
		return def
	}
	if s := toString(args[i]); s != "" {
		return s
	}
	return def
}

// argInt returns the i-th arg as an int, or def when absent.
func argInt(args []any, i int, def int) int {
	if i >= len(args) {
		return def
	}
	return toInt(args[i])
}
