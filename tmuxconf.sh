MY_SESSION=$(tmux list-sessions | grep "poefilters")
if [[ ! $MY_SESSION ]]; then
		# create a new session and `-d`etach
		tmux new-session -d -s poefilters
		tmux split-window -h
		tmux resizep -t"{right}" -x "25%"
		tmux select-pane -t 0
		tmux send "go run . --watch" Enter
fi
tmux attach-session -d -t poefilters
