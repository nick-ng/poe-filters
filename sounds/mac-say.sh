say -v Daniel -r 360 -o daniel-helm-3-link.aiff "helm three link"
say -v Daniel -r 360 -o daniel-body-3-link.aiff "body three link"
say -v Daniel -r 360 -o daniel-gloves-3-link.aiff "gloves three link"
say -v Daniel -r 360 -o daniel-boots-3-link.aiff "boots three link"

say -v Daniel -r 360 -o daniel-helm-4-link.aiff "helm four link"
say -v Daniel -r 360 -o daniel-body-4-link.aiff "body four link"
say -v Daniel -r 360 -o daniel-gloves-4-link.aiff "gloves four link"
say -v Daniel -r 360 -o daniel-boots-4-link.aiff "boots four link"

# Normalise each generated .aiff to -16 LUFS (matching the rest of the library) and encode to .mp3.
# Integrated LUFS is unstable on sub-second clips, so measure the concatenation of every clip
# (long enough to be stable), derive one gain to bring the whole set to -16, and apply it to each.
# A uniform gain preserves the clips' relative loudness; a true-peak limiter catches the few clips
# whose peaks would exceed -1 dBFS after the boost.
aiffs=(*.aiff)
inputs=(); maps=""; n=0
for aiff in "${aiffs[@]}"; do inputs+=(-i "$aiff"); maps="${maps}[${n}:a]"; n=$((n+1)); done
ffmpeg -y "${inputs[@]}" -filter_complex "${maps}concat=n=${n}:v=0:a=1[out]" -map "[out]" /tmp/mac-say-combined.wav
combined_lufs=$(ffmpeg -i /tmp/mac-say-combined.wav -af ebur128 -f null - 2>&1 | sed -n 's/.*I: *\(-*[0-9.]*\) LUFS/\1/p' | tail -1)
gain=$(awk "BEGIN{print -16 - ($combined_lufs)}")
echo "combined loudness ${combined_lufs} LUFS -> applying ${gain} dB to each clip"
# Apply the gain + peak limit, then append silence.mp3 as a tail (resampled to match) before encoding.
for aiff in "${aiffs[@]}"; do
	ffmpeg -y -i "$aiff" -i silence.mp3 -filter_complex \
		"[0:a]volume=${gain}dB,alimiter=limit=0.891:level=disabled,aresample=22050,aformat=sample_fmts=fltp:channel_layouts=mono[clip];[1:a]aresample=22050,aformat=sample_fmts=fltp:channel_layouts=mono[sil];[clip][sil]concat=n=2:v=0:a=1[out]" \
		-map "[out]" -b:a 192k "${aiff%.aiff}.mp3"
done

echo "Done"
