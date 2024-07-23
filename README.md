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

## ToDos

### ToDo Comments

- main.go:107: @todo(nick-ng): move some functions to separate files
