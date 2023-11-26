# poe-filters
Make my item filters for Path of Exile

## Usage
1. See [example.filter](https://github.com/nick-ng/poe-filters/blob/main/my-filters/example.filter)
2. Run executable once to create filter directories
3. Write your .filter files in the `my-filters` directory
4. Run executable again
5. Copy files in `output-filters` to `C:\Users\<your username>\Documents\My Games\Path of Exile`

## Development - VS Code
1. `cp ./.vscode/launch.example.json ./.vscode/launch.json`
2. Change/copy "With Args" block so it has the args you want
3. Run with VS Code's debugger

## Notes

* `CustomAlertSound` can handle absolute paths i.e. `D:\etc\sound.mp3`

## ToDos

- make common base filters
  - 2h weapon progression
- add rare hiding to bases(?) filter
  - use sov general and hide rares on non-t0 bases in maps?

### ToDo Comments

- main.go:89: @todo(nick-ng): move some functions to separate files
