# Examples

`reddit_usage.rb` shows the intended Ruby surface of the `reddit` gem once the
[go-embedded-ruby](https://github.com/go-embedded-ruby) interpreter (rbgo)
binds it as `require "reddit"`.

Run it under rbgo:

```sh
rbgo examples/reddit_usage.rb
```

The Go core in this repo is runtime-independent, so you can exercise the exact
same operations from plain Go without Ruby:

```go
s := reddit.NewSession(reddit.WithUserAgent("app/1.0 (by /u/you)"))
posts, _ := s.Subreddit(ctx, "golang", "hot", 5) // []any of Hashes
```

> Note: the rbgo `reddit` binding is provided by the go-embedded-ruby project.
> Until it lands, use the Go API above (fully tested here) or drive
> `Session.Call` directly.
