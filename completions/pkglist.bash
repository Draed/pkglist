# File: completions/pkglist.bash
# Description: Bash completion for pkglist. Completes -p/--provider with
#   the package providers actually detected as installed on this system
#   (via `pkglist --providers`), including comma-separated lists.
# Date: 2026-10-17

_pkglist() {
    local cur prev providers head tail opt matches

    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    ## detected providers, one per line, queried fresh each time so a
    ## newly installed provider shows up without reloading the shell
    providers="$(command pkglist --providers 2>/dev/null)"

    case "$prev" in
        -p|--provider)
            COMPREPLY=( $(compgen -W "${providers}" -- "${cur}") )
            return 0
            ;;
    esac

    case "${cur}" in
        --provider=*)
            tail="${cur#--provider=}"
            head=""
            if [[ "${tail}" == *,* ]]; then
                head="${tail%,*},"
                tail="${tail##*,}"
            fi
            matches=()
            for opt in ${providers}; do
                ## skip providers already picked earlier in the comma list
                [[ ",${head}" == *",${opt},"* ]] && continue
                matches+=( "--provider=${head}${opt}" )
            done
            COMPREPLY=( $(compgen -W "${matches[*]}" -- "${cur}") )
            return 0
            ;;
        -*)
            COMPREPLY=( $(compgen -W "--verbose -v --provider -p --list-providers --providers --help -h" -- "${cur}") )
            return 0
            ;;
    esac

    return 0
}

complete -F _pkglist pkglist
