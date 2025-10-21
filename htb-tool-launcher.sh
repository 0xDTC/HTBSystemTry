#!/bin/bash
# HTB Tool Auto-Restart Wrapper

while true; do
    /usr/local/bin/htb-tool-bin

    # If exit code is 0 (normal quit), don't restart
    if [ $? -eq 0 ]; then
        break
    fi

    # If exit code is 42 (refresh requested), restart
    if [ $? -eq 42 ]; then
        sleep 1
        continue
    fi

    # For any other exit code, wait a bit before restarting
    sleep 2
done
