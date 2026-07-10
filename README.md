# go-ruby-reddit/reddit

The pure-Go, Ruby-runtime-independent core of the Ruby `reddit` gem — a
reader-oriented client for Reddit's public API, shaped so that
[go-embedded-ruby](https://github.com/go-embedded-ruby) (rbgo) can bind it as
`require "reddit"`.

It is a thin, reflective adapter over [go-reddit/reddit](https://github.com/go-reddit/reddit):
a `Session` exposes the client's operations through a single dynamic entry
point, `Call`, which maps a snake_case method name to a Go call, coerces the
Ruby arguments, and normalises the result into Ruby-shaped values (Hashes,
Arrays, scalars). Nothing here depends on the Ruby runtime, so it is equally
usable as a standalone Go library. **CGO=0, 100% test coverage.**

## Ruby (under rbgo)

```ruby
require "reddit"

session = Reddit::Session.new(user_agent: "app/1.0 (by /u/you)")
posts = session.subreddit("golang", "hot", 25)   # Array of Hashes
posts.each { |p| puts "#{p['score']}  #{p['title']}" }

tree = session.comments("golang", posts.first["id"])  # Hash
```

See [examples/reddit_usage.rb](examples/reddit_usage.rb).

## Go

```go
import "github.com/go-ruby-reddit/reddit"

s := reddit.NewSession(reddit.WithUserAgent("app/1.0 (by /u/you)"))

posts, _ := s.Subreddit(ctx, "golang", "hot", 25) // []any of map[string]any
front, _ := s.Frontpage(ctx, "best", 25)
tree,  _ := s.Comments(ctx, "golang", "abc123")   // map[string]any

// Dynamic dispatch (what the rbgo binding drives):
got, _ := s.Call(ctx, "subreddit", "golang", "hot", 25)
```

Re-exported options: `WithUserAgent`, `WithOAuth`, `WithOAuthScript`,
`WithHTTPClient`, `WithBaseURL`.

Dispatch names accepted by `Call`: `subreddit`/`r`, `frontpage`/`front`/`home`,
`comments`/`post`.

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright the go-ruby-reddit/reddit authors.
