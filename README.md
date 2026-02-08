# gator

`gator` is a simple RSS feed aggregator CLI written in Go.  
It allows users to follow RSS feeds, periodically fetch posts from those feeds, and browse the latest posts directly from the terminal.

This project was built as part of a guided backend learning course and is intended as a learning example rather than a production-ready tool.

---

## Requirements

To run `gator`, you’ll need:

- **Go** (1.20 or newer recommended)
- **PostgreSQL**
- A Unix-like environment (Linux / macOS recommended)

---

## Installation

Because `gator` is a Go program, it compiles to a single static binary.

Install it using:

```bash


Ensure `$GOPATH/bin` is in your PATH, then run:

gator

## Configuration

Create the following file:

~/.gatorconfig.json

With contents similar to:

{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}

Ensure the database exists and migrations have been applied.

## Commands

Register and log in:

gator register <username>
gator login <username>

Add a feed:

gator addfeed "<feed name>" "<feed url>"

Follow or unfollow a feed:

gator follow "<feed url>"
gator unfollow "<feed url>"

Run the feed aggregator (runs continuously until Ctrl+C):

gator agg 30s

Browse recent posts from feeds you follow (default limit is 2):

gator browse
gator browse 5

## Notes

Posts are deduplicated by URL. RSS feeds may contain missing or malformed data and are handled defensively. This project is intended for educational purposes.
