/#!/bin/bash

MY_SESSION=$(tmux list-sessions | grep "poefilters")
if [[ ! $MY_SESSION ]]; then
		# create a new session and `-d`etach
		tmux new-session -d -s poefilters
		tmux send "nvim ." Enter
		tmux new-window
		# tmux split-window -h
fi
