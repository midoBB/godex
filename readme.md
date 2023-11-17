# Godex

Godex is a command-line tool for downloading manga from Mangadex. It allows you to download all unread manga from your follow feed, organize them into folders with correct chapter numbers, and store them in CBZ format.

## Installation

1. **Build the Program:**

    ```bash
    go build -o godex main.go
    ```

2. **Move the Executable:**

    After building, move the `godex` executable to a directory on your user's path, such as `~/.local/bin`.

    ```bash
    mv godex ~/.local/bin/
    ```

## Usage

-   Generate API Key:

Generate an API key from your Mangadex user settings page. Visit Mangadex and navigate to your settings to find the API section.

-   Create an Environment File:

Create a file named .env based on the provided .env.sample file. Fill in your Mangadex API key and the full path of the directory where you want to store your manga.

-   Run the Program

### Download Unread Manga:

```bash
godex
```

The program will authenticate with Mangadex using your API key, fetch the unread manga from your follow feed, and download them into the specified directory in CBZ format.

### Download All Chapters Based on Manga URL:

```bash
godex full --url <manga_url>
```

Downloads all available chapters of a manga based on the provided MangaDex URL.

### Load Environment Variables from a File:

```bash
godex load --env <path_to_env_file>
```

Load environment variables from a specified file. This is useful for managing your Mangadex API key and manga directory.

### Prompt for Configuration:

```bash
godex prompt
```

Fill out the configuration interactively through a series of prompts.

## Additional Commands

- `godex completion`: Generate the autocompletion script for the specified shell.
- `godex help [command]`: Get help about any command.

## Flags

- Global Flags:
  - `-h, --help`: Display help for the main godex command.

- Load Environment Variables Command Flags:
  - `-e, --env <path_to_env_file>`: Path to the environment file.

- Download All Chapters Command Flags:
  - `-h, --help`: Display help for the `full` command.
  - `-u, --url <manga_url>`: URL of the manga to download.

- Prompt for Configuration Command Flags:
  - `-h, --help`: Display help for the `prompt` command.

## Disclaimer

This program is provided as-is and is not affiliated with Mangadex. Use it responsibly and respect the terms of service of Mangadex.

## License

This project is licensed under the [MIT License](LICENSE).
