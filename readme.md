# Godex

Godex is a Go program designed to download unread manga from the follow feed of a user on Mangadex. It simplifies the process of keeping track of your manga and automatically organizes them into folders with correct chapter numbers in CBZ format.

## Installation

-   Build the Program:

```bash
go build -o godex main.go
```

-   Move the Executable

After building, move the godex executable to a directory on your user's path, such as ~/.local/bin.

```bash
mv godex ~/.local/bin/
```

## Usage

-   Generate API Key:

Generate an API key from your Mangadex user settings page. Visit Mangadex and navigate to your settings to find the API section.

-   Create an Environment File:

Create a file named .env based on the provided .env.sample file. Fill in your Mangadex API key and the full path of the directory where you want to store your manga.

-   Run the Program

```bash
godex
```

The program will authenticate with Mangadex using your API key, fetch the unread manga from your follow feed, and download them into the specified directory in CBZ format.


## Disclaimer

This program is provided as-is and is not affiliated with Mangadex. Use it responsibly and respect the terms of service of Mangadex.

## License

This project is licensed under the MIT License.
