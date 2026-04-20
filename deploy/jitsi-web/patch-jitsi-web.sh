#!/bin/sh
set -eu

BUNDLE="/usr/share/jitsi-meet/libs/app.bundle.min.js"
INDEX="/usr/share/jitsi-meet/index.html"
LANG_RU="/usr/share/jitsi-meet/lang/main-ru.json"
BG_MARKER="9139-fistream-backgrounds2"
ASSET_MARKER="9139-fistream-settings7"

require_file() {
  file="$1"
  if [ ! -f "$file" ]; then
    echo "required file not found: $file" >&2
    exit 1
  fi
}

replace_once() {
  file="$1"
  from="$2"
  to="$3"
  if grep -Fq "$to" "$file"; then
    return
  fi
  grep -Fq "$from" "$file"
  sed -i "s|$from|$to|g" "$file"
  grep -Fq "$to" "$file"
}

require_file "$BUNDLE"
require_file "$INDEX"
require_file "$LANG_RU"

# Rename Russian labels for 7 built-in virtual backgrounds.
grep -q '"virtualBackground"' "$LANG_RU"
sed -Ei 's/"image1":[[:space:]]*"[^"]+"/"image1": "\\u0433\\u0435\\u043d\\u0448\\u0438\\u043d"/' "$LANG_RU"
sed -Ei 's/"image2":[[:space:]]*"[^"]+"/"image2": "\\u0413\\u043e\\u0434\\u0436\\u043e \\u0421\\u0430\\u0442\\u043e\\u0440\\u0443"/' "$LANG_RU"
sed -Ei 's/"image3":[[:space:]]*"[^"]+"/"image3": "\\u0437\\u043e\\u043d\\u0430"/' "$LANG_RU"
sed -Ei 's/"image4":[[:space:]]*"[^"]+"/"image4": "\\u041a\\u0438\\u0440\\u0438\\u043b\\u043b"/' "$LANG_RU"
sed -Ei 's/"image5":[[:space:]]*"[^"]+"/"image5": "\\u043a\\u043e\\u043d\\u0438"/' "$LANG_RU"
sed -Ei 's/"image6":[[:space:]]*"[^"]+"/"image6": "\\u0441\\u0442\\u0440\\u043e\\u0439\\u043a\\u0430"/' "$LANG_RU"
sed -Ei 's/"image7":[[:space:]]*"[^"]+"/"image7": "\\u0422\\u042d\\u0426"/' "$LANG_RU"
grep -q '"image1": "\\u0433\\u0435\\u043d\\u0448\\u0438\\u043d"' "$LANG_RU"
grep -q '"image2": "\\u0413\\u043e\\u0434\\u0436\\u043e \\u0421\\u0430\\u0442\\u043e\\u0440\\u0443"' "$LANG_RU"
grep -q '"image3": "\\u0437\\u043e\\u043d\\u0430"' "$LANG_RU"
grep -q '"image4": "\\u041a\\u0438\\u0440\\u0438\\u043b\\u043b"' "$LANG_RU"
grep -q '"image5": "\\u043a\\u043e\\u043d\\u0438"' "$LANG_RU"
grep -q '"image6": "\\u0441\\u0442\\u0440\\u043e\\u0439\\u043a\\u0430"' "$LANG_RU"
grep -q '"image7": "\\u0422\\u042d\\u0426"' "$LANG_RU"

# Hide Shortcuts tab unless explicitly enabled in SETTINGS_SECTIONS.
if grep -Fq 'r.includes("shortcuts")&&!b&&g.push({name:WG,component:J2,labelKey:"settings.shortcuts"' "$BUNDLE"; then
  :
else
  grep -Fq 'name:WG,component:J2,labelKey:"settings.shortcuts"' "$BUNDLE"
  grep -Fq '!b&&g.push({name:WG,component:J2,labelKey:"settings.shortcuts"' "$BUNDLE"
  sed -i \
    's/!b&&g.push({name:WG,component:J2,labelKey:"settings.shortcuts"/r.includes("shortcuts")\&\&!b\&\&g.push({name:WG,component:J2,labelKey:"settings.shortcuts"/g' \
    "$BUNDLE"
  grep -Fq 'r.includes("shortcuts")&&!b&&g.push({name:WG,component:J2,labelKey:"settings.shortcuts"' "$BUNDLE"
fi

# Replace reaction sounds and emoji set.
sed -i 's/reactions-thumbs-up.mp3/puk-v-ekho.mp3/g' "$BUNDLE"
sed -i 's/reactions-applause.mp3/gey-echo.mp3/g' "$BUNDLE"
sed -i 's/reactions-laughter.mp3/mlg-airhorn.mp3/g' "$BUNDLE"
sed -i 's/reactions-surprise.mp3/anime-wow.mp3/g' "$BUNDLE"
sed -i 's/reactions-boo.mp3/vine-boom.mp3/g' "$BUNDLE"
sed -i 's/reactions-crickets.mp3/among-us.mp3/g' "$BUNDLE"
sed -i 's/reactions-love.mp3/pribyl-godzho-satoru.mp3/g' "$BUNDLE"

sed -i 's/like:{message:":thumbs_up:",emoji:"[^"]*"/like:{message:"\\ud83d\\udca9",emoji:"\\ud83d\\udca9"/' "$BUNDLE"
sed -i 's/clap:{message:":clap:",emoji:"[^"]*"/clap:{message:"\\ud83c\\udf46",emoji:"\\ud83c\\udf46"/' "$BUNDLE"
sed -i 's/laugh:{message:":grinning_face:",emoji:"[^"]*"/laugh:{message:"\\ud83e\\udd11",emoji:"\\ud83e\\udd11"/' "$BUNDLE"
sed -i 's/surprised:{message:":face_with_open_mouth:",emoji:"[^"]*"/surprised:{message:"\\ud83c\\udf51",emoji:"\\ud83c\\udf51"/' "$BUNDLE"
sed -i 's/boo:{message:":slightly_frowning_face:",emoji:"[^"]*"/boo:{message:"\\ud83e\\udd21",emoji:"\\ud83e\\udd21"/' "$BUNDLE"
sed -i 's/silence:{message:":face_without_mouth:",emoji:"[^"]*"/silence:{message:"\\ud83d\\udc79",emoji:"\\ud83d\\udc79"/' "$BUNDLE"
sed -i 's/love:{message:":heart:",emoji:"[^"]*"/love:{message:"\\ud83d\\ude0e",emoji:"\\ud83d\\ude0e"/' "$BUNDLE"
sed -Ei 's/shortcutChar:"[TCLOBSH]"/shortcutChar:""/g' "$BUNDLE"
sed -i 's/tooltip:`\${t(`toolbar.\${n}`)} (\${r} + \${YJ\[n\].shortcutChar})`/tooltip:void 0/g' "$BUNDLE"

grep -Fq 'puk-v-ekho.mp3' "$BUNDLE"
grep -Fq 'pribyl-godzho-satoru.mp3' "$BUNDLE"
grep -Fq 'like:{message:"\ud83d\udca9",emoji:"\ud83d\udca9"' "$BUNDLE"
grep -Fq 'love:{message:"\ud83d\ude0e",emoji:"\ud83d\ude0e"' "$BUNDLE"

# Cache-bust background and locale URLs only once.
if grep -Fq "images/virtual-background/background-1.jpg?v=${BG_MARKER}" "$BUNDLE"; then
  :
else
  grep -Fq 'images/virtual-background/background-1.jpg' "$BUNDLE"
  grep -Fq 'images/virtual-background/background-7.jpg' "$BUNDLE"
  replace_once "$BUNDLE" 'images/virtual-background/background-1.jpg' "images/virtual-background/background-1.jpg?v=${BG_MARKER}"
  replace_once "$BUNDLE" 'images/virtual-background/background-2.jpg' "images/virtual-background/background-2.jpg?v=${BG_MARKER}"
  replace_once "$BUNDLE" 'images/virtual-background/background-3.jpg' "images/virtual-background/background-3.jpg?v=${BG_MARKER}"
  replace_once "$BUNDLE" 'images/virtual-background/background-4.jpg' "images/virtual-background/background-4.jpg?v=${BG_MARKER}"
  replace_once "$BUNDLE" 'images/virtual-background/background-5.jpg' "images/virtual-background/background-5.jpg?v=${BG_MARKER}"
  replace_once "$BUNDLE" 'images/virtual-background/background-6.jpg' "images/virtual-background/background-6.jpg?v=${BG_MARKER}"
  replace_once "$BUNDLE" 'images/virtual-background/background-7.jpg' "images/virtual-background/background-7.jpg?v=${BG_MARKER}"
fi

if grep -Fq "lang/{{ns}}-{{lng}}.json?v=${BG_MARKER}" "$BUNDLE"; then
  :
else
  replace_once "$BUNDLE" 'lang/{{ns}}-{{lng}}.json' "lang/{{ns}}-{{lng}}.json?v=${BG_MARKER}"
  replace_once "$BUNDLE" 'lang/{{ns}}.json' "lang/{{ns}}.json?v=${BG_MARKER}"
fi

# Force loading updated static assets after hotfixes.
if grep -Fq "libs/app.bundle.min.js?v=${ASSET_MARKER}" "$INDEX"; then
  :
else
  grep -Fq 'css/all.css?v=9139' "$INDEX"
  grep -Fq 'libs/lib-jitsi-meet.min.js?v=9139' "$INDEX"
  grep -Fq 'libs/app.bundle.min.js?v=9139' "$INDEX"
  sed -i "s|css/all.css?v=9139|css/all.css?v=${ASSET_MARKER}|g" "$INDEX"
  sed -i "s|libs/lib-jitsi-meet.min.js?v=9139|libs/lib-jitsi-meet.min.js?v=${ASSET_MARKER}|g" "$INDEX"
  sed -i "s|libs/app.bundle.min.js?v=9139|libs/app.bundle.min.js?v=${ASSET_MARKER}|g" "$INDEX"
fi

grep -Fq "css/all.css?v=${ASSET_MARKER}" "$INDEX"
grep -Fq "libs/lib-jitsi-meet.min.js?v=${ASSET_MARKER}" "$INDEX"
grep -Fq "libs/app.bundle.min.js?v=${ASSET_MARKER}" "$INDEX"

grep -Fq "images/virtual-background/background-1.jpg?v=${BG_MARKER}" "$BUNDLE"
grep -Fq "images/virtual-background/background-7.jpg?v=${BG_MARKER}" "$BUNDLE"
grep -Fq "lang/{{ns}}-{{lng}}.json?v=${BG_MARKER}" "$BUNDLE"
grep -Fq "lang/{{ns}}.json?v=${BG_MARKER}" "$BUNDLE"
