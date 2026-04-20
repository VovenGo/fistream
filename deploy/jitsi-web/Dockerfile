FROM jitsi/web:stable-10888

# Replace default virtual backgrounds with fistream custom set.
COPY deploy/jitsi-web/backgrounds/background-1.jpg /usr/share/jitsi-meet/images/virtual-background/background-1.jpg
COPY deploy/jitsi-web/backgrounds/background-2.jpg /usr/share/jitsi-meet/images/virtual-background/background-2.jpg
COPY deploy/jitsi-web/backgrounds/background-3.jpg /usr/share/jitsi-meet/images/virtual-background/background-3.jpg
COPY deploy/jitsi-web/backgrounds/background-4.jpg /usr/share/jitsi-meet/images/virtual-background/background-4.jpg
COPY deploy/jitsi-web/backgrounds/background-5.jpg /usr/share/jitsi-meet/images/virtual-background/background-5.jpg
COPY deploy/jitsi-web/backgrounds/background-6.jpg /usr/share/jitsi-meet/images/virtual-background/background-6.jpg
COPY deploy/jitsi-web/backgrounds/background-7.jpg /usr/share/jitsi-meet/images/virtual-background/background-7.jpg

# Replace reaction sounds with fistream custom set.
COPY deploy/jitsi-web/reactions/puk-v-ekho.mp3 /usr/share/jitsi-meet/sounds/puk-v-ekho.mp3
COPY deploy/jitsi-web/reactions/gey-echo.mp3 /usr/share/jitsi-meet/sounds/gey-echo.mp3
COPY deploy/jitsi-web/reactions/mlg-airhorn.mp3 /usr/share/jitsi-meet/sounds/mlg-airhorn.mp3
COPY deploy/jitsi-web/reactions/anime-wow.mp3 /usr/share/jitsi-meet/sounds/anime-wow.mp3
COPY deploy/jitsi-web/reactions/vine-boom.mp3 /usr/share/jitsi-meet/sounds/vine-boom.mp3
COPY deploy/jitsi-web/reactions/among-us.mp3 /usr/share/jitsi-meet/sounds/among-us.mp3
COPY deploy/jitsi-web/reactions/pribyl-godzho-satoru.mp3 /usr/share/jitsi-meet/sounds/pribyl-godzho-satoru.mp3

# Apply idempotent bundle/index/lang patches required by fistream UI contract.
COPY deploy/jitsi-web/patch-jitsi-web.sh /usr/local/bin/patch-jitsi-web.sh
RUN sh /usr/local/bin/patch-jitsi-web.sh && rm -f /usr/local/bin/patch-jitsi-web.sh
