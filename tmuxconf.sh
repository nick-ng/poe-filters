MY_SESSION=$(tmux list-sessions | grep "poefilters")
if [[ ! $MY_SESSION ]]; then
		# create a new session and `-d`etach
		tmux new-session -d -s poefilters
		tmux split-window -h
		tmux resizep -t"{right}" -x "33%"
		tmux select-pane -t 0
fi
tmux attach-session -d -t poefilters
