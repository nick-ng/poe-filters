MY_SESSION=$(tmux list-sessions | grep "poefilters")
if [[ ! $MY_SESSION ]]; then
		# create a new session and `-d`etach
		tmux new-session -d -s poefilters
		tmux split-window -h
		tmux send "go run . --watch" Enter
		tmux send "tmux resizep -x 25%" Enter
		tmux select-pane -t 0
fi
tmux attach-session -d -t poefilters
