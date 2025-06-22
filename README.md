# poe-filters
Make my item filters for Path of Exile

## Requirements

- [Go](https://go.dev/)
- [ffmpeg](https://ffmpeg.org/)

## Usage

When you run it, it will try to copy filters to the Path of Exile directory

1. See [example.filter](https://github.com/nick-ng/poe-filters/blob/main/my-filters/example.filter)
2. Run executable once to create filter directories
3. Write your .filter files in the `my-filters` directory
4. Run executable again

## Development - VS Code
1. `cp ./.vscode/launch.example.json ./.vscode/launch.json`
2. Change/copy "With Args" block so it has the args you want
3. Run with VS Code's debugger

## Notes

- `CustomAlertSound` can handle absolute paths i.e. `D:\etc\sound.mp3`

## Attributions

- `utils\tailwindcss-colours.json` taken from https://tailwindcss.com

## ToDos

- make `colour-tokens.json` only have rgb values and automatically make text, background and border versions when replacing

### ToDo Comments

- main.go:149: @todo(nick-ng): move some functions to separate files
- main.go:401: @todo(nick-ng): move these to a method so we can process all commands (multi-line or otherwise) in a single loop?
- main.go:527: // @todo(nick-ng): move this to its own loop
- utils\items.go:66: @todo(nick-ng): since you have custom styles, the item class can also be part of the custom styles...
- utils\misc.go:160: @todo(nick-ng): this has weird behaviour if you "open" and "close" quotes multiple times
