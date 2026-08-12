#!/bin/bash

if [ -f ./copy-filters.custom.sh ]; then
	echo "Copying filters and sounds"

	./copy-filters.custom.sh

	echo "Finished copying filters and sounds"
fi
