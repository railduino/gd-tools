#!/bin/bash

_gdt_completion() {
    local words cword
    words=("${COMP_WORDS[@]}")
    cword=$COMP_CWORD

    local completions
    completions=$( gdt "${words[@]:1}" --generate-bash-completion 2>/dev/null )
    COMPREPLY=( $(compgen -W "$completions" -- "${words[cword]}") )
    return 0
}

complete -F _gdt_completion gdt

