# Gomapit

A CLI tool to generate sitemap.xml for the given URL.

## Installation

Use the go tool to fetch the code:

```
go get github.com/erbesharat/gomapit
```

## Usage

Available flags:

- `-output=`: Output file path (Default: `./sitemap.xml`) **Required**
- `-depth=`: Max depth of url navigation recursion (Default: `1`) **Optional**
- `-parallel`: Number of parallel workers to navigate through site (Default: `1`) **Optional**

Examples:

```
gomapit -depth 2 -parallel 2 -output ./map.xml https://erbesharat.github.io
```

This command will go two levels deep to find the URLs inside the given URL with 2 parallel workers and writes the result to the `map.xml` file.

Or

```
gomapit https://erbesharat.github.io
```

This command only checks the first page and writes the result to the `sitemap.xml` file in the current directory.

## License

MIT
