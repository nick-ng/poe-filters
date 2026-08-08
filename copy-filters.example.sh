#!/bin/bash

# copy filters to secondary Path of Exile directories
cp -v ~/.steam/steam/steamapps/compatdata/238960/pfx/drive_c/users/steamuser/My\ Documents/My\ Games/Path\ of\ Exile/*.filter /path/to/secondary-poe-dir/

# copy tts to secondary Path of Exile directories
cp -v -u ~/.steam/steam/steamapps/compatdata/238960/pfx/drive_c/users/steamuser/My\ Documents/My\ Games/Path\ of\ Exile/tts/* /path/to/secondary-poe-dir/tts/

# copy sounds to secondary Path of Exile directories
cp -v -u ~/.steam/steam/steamapps/compatdata/238960/pfx/drive_c/users/steamuser/My\ Documents/My\ Games/Path\ of\ Exile/sounds/* /path/to/secondary-poe-dir/sounds/
