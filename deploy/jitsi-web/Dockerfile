FROM jitsi/web:stable-10888

# Replace default virtual background images with fistream custom set.
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

# Rename Russian virtual background labels to match fistream custom set.
RUN lang="/usr/share/jitsi-meet/lang/main-ru.json"; \
  set -eu; \
  grep -q '"virtualBackground"' "$lang"; \
  sed -Ei 's/"image1":[[:space:]]*"[^"]+"/"image1": "\\u0433\\u0435\\u043d\\u0448\\u0438\\u043d"/' "$lang"; \
  sed -Ei 's/"image2":[[:space:]]*"[^"]+"/"image2": "\\u0413\\u043e\\u0434\\u0436\\u043e \\u0421\\u0430\\u0442\\u043e\\u0440\\u0443"/' "$lang"; \
  sed -Ei 's/"image3":[[:space:]]*"[^"]+"/"image3": "\\u0437\\u043e\\u043d\\u0430"/' "$lang"; \
  sed -Ei 's/"image4":[[:space:]]*"[^"]+"/"image4": "\\u041a\\u0438\\u0440\\u0438\\u043b\\u043b"/' "$lang"; \
  sed -Ei 's/"image5":[[:space:]]*"[^"]+"/"image5": "\\u043a\\u043e\\u043d\\u0438"/' "$lang"; \
  sed -Ei 's/"image6":[[:space:]]*"[^"]+"/"image6": "\\u0441\\u0442\\u0440\\u043e\\u0439\\u043a\\u0430"/' "$lang"; \
  sed -Ei 's/"image7":[[:space:]]*"[^"]+"/"image7": "\\u0422\\u042d\\u0426"/' "$lang"; \
  grep -q '"image1": "\\u0433\\u0435\\u043d\\u0448\\u0438\\u043d"' "$lang"; \
  grep -q '"image2": "\\u0413\\u043e\\u0434\\u0436\\u043e \\u0421\\u0430\\u0442\\u043e\\u0440\\u0443"' "$lang"; \
  grep -q '"image3": "\\u0437\\u043e\\u043d\\u0430"' "$lang"; \
  grep -q '"image4": "\\u041a\\u0438\\u0440\\u0438\\u043b\\u043b"' "$lang"; \
  grep -q '"image5": "\\u043a\\u043e\\u043d\\u0438"' "$lang"; \
  grep -q '"image6": "\\u0441\\u0442\\u0440\\u043e\\u0439\\u043a\\u0430"' "$lang"; \
  grep -q '"image7": "\\u0422\\u042d\\u0426"' "$lang"

# Hide "Shortcuts" settings tab unless it is explicitly enabled in SETTINGS_SECTIONS.
RUN bundle="/usr/share/jitsi-meet/libs/app.bundle.min.js"; \
  set -eu; \
  grep -q 'name:WG,component:J2,labelKey:"settings.shortcuts"' "$bundle"; \
  grep -q '!b&&g.push({name:WG,component:J2,labelKey:"settings.shortcuts"' "$bundle"; \
  sed -i \
    's/!b&&g.push({name:WG,component:J2,labelKey:"settings.shortcuts"/r.includes("shortcuts")\&\&!b\&\&g.push({name:WG,component:J2,labelKey:"settings.shortcuts"/g' \
    "$bundle"; \
  grep -q 'r.includes("shortcuts")&&!b&&g.push({name:WG,component:J2,labelKey:"settings.shortcuts"' "$bundle"

# Replace reactions menu emojis/sounds and drop the 7th "love" reaction.
RUN bundle="/usr/share/jitsi-meet/libs/app.bundle.min.js"; \
  set -eu; \
  grep -q 'reactions-thumbs-up.mp3' "$bundle"; \
  grep -q 'reactions-applause.mp3' "$bundle"; \
  grep -q 'reactions-laughter.mp3' "$bundle"; \
  grep -q 'reactions-surprise.mp3' "$bundle"; \
  grep -q 'reactions-boo.mp3' "$bundle"; \
  grep -q 'reactions-crickets.mp3' "$bundle"; \
  grep -q 'reactions-love.mp3' "$bundle"; \
  sed -i 's/reactions-thumbs-up.mp3/puk-v-ekho.mp3/g' "$bundle"; \
  sed -i 's/reactions-applause.mp3/gey-echo.mp3/g' "$bundle"; \
  sed -i 's/reactions-laughter.mp3/mlg-airhorn.mp3/g' "$bundle"; \
  sed -i 's/reactions-surprise.mp3/anime-wow.mp3/g' "$bundle"; \
  sed -i 's/reactions-boo.mp3/vine-boom.mp3/g' "$bundle"; \
  sed -i 's/reactions-crickets.mp3/among-us.mp3/g' "$bundle"; \
  sed -i 's/reactions-love.mp3/among-us.mp3/g' "$bundle"; \
  sed -i 's/like:{message:":thumbs_up:",emoji:"[^"]*"/like:{message:"\\ud83d\\udca9",emoji:"\\ud83d\\udca9"/' "$bundle"; \
  sed -i 's/clap:{message:":clap:",emoji:"[^"]*"/clap:{message:"\\ud83c\\udf46",emoji:"\\ud83c\\udf46"/' "$bundle"; \
  sed -i 's/laugh:{message:":grinning_face:",emoji:"[^"]*"/laugh:{message:"\\ud83e\\udd11",emoji:"\\ud83e\\udd11"/' "$bundle"; \
  sed -i 's/surprised:{message:":face_with_open_mouth:",emoji:"[^"]*"/surprised:{message:"\\ud83c\\udf51",emoji:"\\ud83c\\udf51"/' "$bundle"; \
  sed -i 's/boo:{message:":slightly_frowning_face:",emoji:"[^"]*"/boo:{message:"\\ud83e\\udd21",emoji:"\\ud83e\\udd21"/' "$bundle"; \
  sed -i 's/silence:{message:":face_without_mouth:",emoji:"[^"]*"/silence:{message:"\\ud83d\\udc79",emoji:"\\ud83d\\udc79"/' "$bundle"; \
  sed -i 's/,love:{message:":heart:",emoji:"[^"]*",shortcutChar:"H",soundId:`${HJ}_HEART_`,soundFiles:FJ}//' "$bundle"; \
  grep -q 'puk-v-ekho.mp3' "$bundle"; \
  grep -q 'gey-echo.mp3' "$bundle"; \
  grep -q 'mlg-airhorn.mp3' "$bundle"; \
  grep -q 'anime-wow.mp3' "$bundle"; \
  grep -q 'vine-boom.mp3' "$bundle"; \
  grep -q 'among-us.mp3' "$bundle"; \
  grep -q 'like:{message:"\\ud83d\\udca9",emoji:"\\ud83d\\udca9"' "$bundle"; \
  grep -q 'clap:{message:"\\ud83c\\udf46",emoji:"\\ud83c\\udf46"' "$bundle"; \
  grep -q 'laugh:{message:"\\ud83e\\udd11",emoji:"\\ud83e\\udd11"' "$bundle"; \
  grep -q 'surprised:{message:"\\ud83c\\udf51",emoji:"\\ud83c\\udf51"' "$bundle"; \
  grep -q 'boo:{message:"\\ud83e\\udd21",emoji:"\\ud83e\\udd21"' "$bundle"; \
  grep -q 'silence:{message:"\\ud83d\\udc79",emoji:"\\ud83d\\udc79"' "$bundle"; \
  ! grep -q 'love:{message:":heart:"' "$bundle"

# Cache-bust virtual background and locale resource URLs referenced from app bundle.
RUN bundle="/usr/share/jitsi-meet/libs/app.bundle.min.js"; \
  marker="9139-fistream-backgrounds2"; \
  set -eu; \
  grep -q 'images/virtual-background/background-1.jpg' "$bundle"; \
  grep -q 'images/virtual-background/background-7.jpg' "$bundle"; \
  grep -q 'lang/{{ns}}-{{lng}}.json' "$bundle"; \
  grep -q 'lang/{{ns}}.json' "$bundle"; \
  sed -i "s#images/virtual-background/background-1.jpg#images/virtual-background/background-1.jpg?v=${marker}#g" "$bundle"; \
  sed -i "s#images/virtual-background/background-2.jpg#images/virtual-background/background-2.jpg?v=${marker}#g" "$bundle"; \
  sed -i "s#images/virtual-background/background-3.jpg#images/virtual-background/background-3.jpg?v=${marker}#g" "$bundle"; \
  sed -i "s#images/virtual-background/background-4.jpg#images/virtual-background/background-4.jpg?v=${marker}#g" "$bundle"; \
  sed -i "s#images/virtual-background/background-5.jpg#images/virtual-background/background-5.jpg?v=${marker}#g" "$bundle"; \
  sed -i "s#images/virtual-background/background-6.jpg#images/virtual-background/background-6.jpg?v=${marker}#g" "$bundle"; \
  sed -i "s#images/virtual-background/background-7.jpg#images/virtual-background/background-7.jpg?v=${marker}#g" "$bundle"; \
  sed -i "s#lang/{{ns}}-{{lng}}.json#lang/{{ns}}-{{lng}}.json?v=${marker}#g" "$bundle"; \
  sed -i "s#lang/{{ns}}.json#lang/{{ns}}.json?v=${marker}#g" "$bundle"; \
  grep -q "images/virtual-background/background-1.jpg?v=${marker}" "$bundle"; \
  grep -q "images/virtual-background/background-7.jpg?v=${marker}" "$bundle"; \
  grep -q "lang/{{ns}}-{{lng}}.json?v=${marker}" "$bundle"; \
  grep -q "lang/{{ns}}.json?v=${marker}" "$bundle"

# Bump static asset query marker to force clients to fetch fresh JS/CSS after hotfix.
RUN index="/usr/share/jitsi-meet/index.html"; \
  set -eu; \
  grep -q 'css/all.css?v=9139' "$index"; \
  grep -q 'libs/lib-jitsi-meet.min.js?v=9139' "$index"; \
  grep -q 'libs/app.bundle.min.js?v=9139' "$index"; \
  sed -i 's/css\/all.css?v=9139/css\/all.css?v=9139-fistream-settings5/g' "$index"; \
  sed -i 's/libs\/lib-jitsi-meet.min.js?v=9139/libs\/lib-jitsi-meet.min.js?v=9139-fistream-settings5/g' "$index"; \
  sed -i 's/libs\/app.bundle.min.js?v=9139/libs\/app.bundle.min.js?v=9139-fistream-settings5/g' "$index"; \
  grep -q 'css/all.css?v=9139-fistream-settings5' "$index"; \
  grep -q 'libs/lib-jitsi-meet.min.js?v=9139-fistream-settings5' "$index"; \
  grep -q 'libs/app.bundle.min.js?v=9139-fistream-settings5' "$index"
