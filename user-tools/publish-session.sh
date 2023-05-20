#!/bin/bash
# This script is used to publish a session to the web. probably just a placeholder $1 is session number

set -e

if [ -z "$1" ]; then
	echo "Usage: $0 <session>"
	exit 1
fi

google-chrome --profile-directory="Profile 3" --new-window \
	https://studio.youtube.com/channel/UCAvwN8vS3f0FFNxWoh3jqyQ/videos/upload?d=ud \
	https://www.tiktok.com/upload/ \
	https://www.instagram.com/astudyoflight_/ \
	https://twitter.com/compose/tweet \
	https://business.facebook.com/latest/posts/published_posts?asset_id=116836424739888&nav_ref=bm_home_redirect \

cd ~/asol/software/asol-backend

urxvt -e bash -ic "python3 ./user-tools/upload_helper.py -n $1 -d /home/greg/Downloads" #-d /mnt/md0/light-stores"