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
