FROM jitsi/web:stable-10888

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
  sed -i 's/css\/all.css?v=9139/css\/all.css?v=9139-fistream-settings2/g' "$index"; \
  sed -i 's/libs\/lib-jitsi-meet.min.js?v=9139/libs\/lib-jitsi-meet.min.js?v=9139-fistream-settings2/g' "$index"; \
  sed -i 's/libs\/app.bundle.min.js?v=9139/libs\/app.bundle.min.js?v=9139-fistream-settings2/g' "$index"; \
  grep -q 'css/all.css?v=9139-fistream-settings2' "$index"; \
  grep -q 'libs/lib-jitsi-meet.min.js?v=9139-fistream-settings2' "$index"; \
  grep -q 'libs/app.bundle.min.js?v=9139-fistream-settings2' "$index"
