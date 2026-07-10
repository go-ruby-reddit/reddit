# frozen_string_literal: true
#
# Pure-Ruby usage of the `reddit` gem (github.com/go-ruby-reddit/reddit),
# as bound by the go-embedded-ruby interpreter (rbgo). The Go core is a thin
# adapter over github.com/go-reddit/reddit; every call returns plain Ruby
# Hashes and Arrays.

require "reddit"

# A descriptive User-Agent is required — Reddit rate-limits generic ones.
# For anything server-side, prefer OAuth (anonymous .json is often blocked).
session = Reddit::Session.new(user_agent: "reddit-gem-example/1.0 (by /u/you)")

# A subreddit listing comes back as an Array of Hashes.
posts = session.subreddit("golang", "hot", 5)
posts.each do |p|
  puts format("%6d  %s  (u/%s)", p["score"], p["title"], p["author"])
end

# The logged-out front page:
front = session.frontpage("best", 5)
puts "front page has #{front.length} posts"

# A post plus its comment tree, as a Hash with "Post" and "Comments" keys:
if posts.any?
  tree = session.comments("golang", posts.first["id"])
  puts %(top comment: #{tree["Comments"].first&.fetch("body", "(none)")})
end
