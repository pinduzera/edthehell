#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
manual="$script_dir/MANUAL.md"

die() {
	printf 'manual: %s\n' "$*" >&2
	exit 1
}

command -v gum >/dev/null 2>&1 ||
	die 'gum not found; install it with "brew install gum" or "go install github.com/charmbracelet/gum@latest"'
[[ -f $manual ]] || die "MANUAL.md not found next to $0"

sections() {
	sed -n 's/^## //p' "$manual"
}

extract_section() {
	local want=$1
	awk -v want="$want" '
		want == "Overview" { if (/^## /) exit; print; next }
		/^## / {
			if (printing) exit
			if (substr($0, 4) == want) printing = 1
		}
		printing { print }
	' "$manual"
}

usage() {
	gum style --border rounded --padding "0 2" --margin "1 0" --bold --foreground 212 "post manual"
	cat <<'EOF'
usage: manual.sh [section]

Without arguments, shows an interactive menu. With a section name
(case-insensitive), renders just that section.

  manual.sh --list     list the available sections
  manual.sh --all      render the whole manual
  manual.sh Flags      render one section
EOF
}

render() {
	local title=$1 body=$2
	if [[ -t 1 ]]; then
		gum style --border rounded --padding "0 2" --margin "1 0" --bold --foreground 212 "$title"
		printf '%s\n' "$body" |
			gum format -t markdown --theme dark |
			gum pager --border rounded --border-foreground 240 --height 30
	else
		printf '%s\n' "$body" | gum format -t markdown --theme notty
	fi
}

show_section() {
	local name=$1
	render "$name" "$(extract_section "$name")"
}

find_section() {
	local want=${1,,} name
	while IFS= read -r name; do
		if [[ ${name,,} == "$want" ]]; then
			printf '%s\n' "$name"
			return 0
		fi
	done < <(sections)
	return 1
}

case ${1:-} in
-h | --help)
	usage
	exit 0
	;;
--list)
	sections
	exit 0
	;;
--all)
	render "post manual" "$(cat "$manual")"
	exit 0
	;;
"")
	if [[ ! -t 0 || ! -t 1 ]]; then
		render "post manual" "$(cat "$manual")"
		exit 0
	fi
	;;
*)
	name=$(find_section "$1") || die "unknown section \"$1\"; try --list"
	show_section "$name"
	exit 0
	;;
esac

while true; do
	choice=$(
		{
			sections
			printf '%s\n' "Full manual" "Quit"
		} | gum choose --header "post manual" --cursor "> " --height 20
	)
	case $choice in
	"" | Quit)
		exit 0
		;;
	"Full manual")
		render "post manual" "$(cat "$manual")"
		;;
	*)
		show_section "$choice"
		;;
	esac
done
