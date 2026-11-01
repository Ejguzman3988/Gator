# Gator RSS Feed

Gator is a CLI RSS feed for users to login locally and gather rss posts from feeds.

## Installation

You will need postgres and go installed

## Usage

```bash
go install
```

You will need to set up a config and database schema at ~/.gatorconfig.json

This is where you will include your database url and the current users will be saved.

```bash
go run . register user
go run . login user
go run . addfeed "https://examplefeed.com"
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)
