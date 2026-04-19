FROM jitsi/web:stable-10888

# Replace default virtual background images with fistream custom set.
COPY deploy/jitsi-web/backgrounds/background-1.jpg /usr/share/jitsi-meet/images/virtual-background/background-1.jpg
COPY deploy/jitsi-web/backgrounds/background-2.jpg /usr/share/jitsi-meet/images/virtual-background/background-2.jpg
COPY deploy/jitsi-web/backgrounds/background-3.jpg /usr/share/jitsi-meet/images/virtual-background/background-3.jpg
COPY deploy/jitsi-web/backgrounds/background-4.jpg /usr/share/jitsi-meet/images/virtual-background/background-4.jpg
COPY deploy/jitsi-web/backgrounds/background-5.jpg /usr/share/jitsi-meet/images/virtual-background/background-5.jpg
COPY deploy/jitsi-web/backgrounds/background-6.jpg /usr/share/jitsi-meet/images/virtual-background/background-6.jpg
COPY deploy/jitsi-web/backgrounds/background-7.jpg /usr/share/jitsi-meet/images/virtual-background/background-7.jpg

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

# Bump static asset query marker to force clients to fetch fresh JS/CSS after hotfix.
RUN index="/usr/share/jitsi-meet/index.html"; \
  set -eu; \
  grep -q 'css/all.css?v=9139' "$index"; \
  grep -q 'libs/lib-jitsi-meet.min.js?v=9139' "$index"; \
  grep -q 'libs/app.bundle.min.js?v=9139' "$index"; \
  sed -i 's/css\/all.css?v=9139/css\/all.css?v=9139-fistream-settings3/g' "$index"; \
  sed -i 's/libs\/lib-jitsi-meet.min.js?v=9139/libs\/lib-jitsi-meet.min.js?v=9139-fistream-settings3/g' "$index"; \
  sed -i 's/libs\/app.bundle.min.js?v=9139/libs\/app.bundle.min.js?v=9139-fistream-settings3/g' "$index"; \
  grep -q 'css/all.css?v=9139-fistream-settings3' "$index"; \
  grep -q 'libs/lib-jitsi-meet.min.js?v=9139-fistream-settings3' "$index"; \
  grep -q 'libs/app.bundle.min.js?v=9139-fistream-settings3' "$index"
