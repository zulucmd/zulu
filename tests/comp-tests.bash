#!/usr/bin/env bash

echo "===================================================="
echo "Running completions tests on $(uname) with bash $BASH_VERSION"
echo "===================================================="

# Test logging using $BASH_COMP_DEBUG_FILE
verifyDebug() {
  debugFile=/tmp/comptests.bash.debug
  rm -f $debugFile
  export BASH_COMP_DEBUG_FILE=$debugFile
  _completionTests_verifyCompletion "testprog help comp" "completion" nofile
  if ! test -s $debugFile; then
    # File should not be empty
    printf "%bERROR: No debug logs were printed to %s%b\n" "${RED}" "${debugFile}" "${NC}"
    _completionTests_TEST_FAILED=1
  else
    printf "%bSUCCESS: Debug logs were printed to %s%b\n" "${GREEN}" "${debugFile}" "${NC}"
  fi
  unset BASH_COMP_DEBUG_FILE
}

# Test completion with a redirection
# https://github.com/spf13/cobra/issues/1334
verifyRedirect() {
  rm -f notexist
  _completionTests_verifyCompletion "testprog completion bash > notexist" ""
  if test -f notexist; then
    # File should not exist
    printf "%bERROR: completion mistakenly created the file 'notexist'%b\n" "${RED}" "${NC}"
    _completionTests_TEST_FAILED=1
    rm -f notexist
  else
    printf "%bSUCCESS: No extra file created, as expected%b\n" "${GREEN}" "${NC}"
  fi
}

ROOTDIR="$PWD"
export PATH="$ROOTDIR/testprog/bin:$PATH"

# Source the testing logic
# shellcheck source=/dev/null
source "$ROOTDIR/src/comp-test-lib.bash"

# Setup completion of testprog, disabling descriptions.
# Don't use the new source <() form as it does not work with bash v3.
# Normally, compopt is a builtin, and the script checks that it is a
# builtin to disable it if we are in bash3 (where compopt does not exist).
# We replace 'builtin' with 'function' because we cannot use the native
# compopt since we are explicitely calling the completion code instead
# of from within a real completion environment.
# shellcheck source=/dev/null
source /dev/stdin <<-EOF
   $(testprog completion bash --no-descriptions | sed s/builtin/function/g)
EOF

cd testingdir

_completionTests_verifyCompletion "testprog comp" "completion" nofile
_completionTests_verifyCompletion "testprog completion " "bash fish powershell zsh" nofile
_completionTests_verifyCompletion "testprog help comp" "completion" nofile
_completionTests_verifyCompletion "testprog completion bash " "" nofile

#################################################
# Completions are filtered by prefix by program
#################################################

# Test ShellCompDirectiveDefault => File completion when no other completions
_completionTests_verifyCompletion "testprog prefix default " "bear bearpaw dog unicorn"
_completionTests_verifyCompletion "testprog prefix default u" "unicorn"
_completionTests_verifyCompletion "testprog prefix default f" ""
_completionTests_verifyCompletion "testprog prefix default z" ""

# Test ShellCompDirectiveNoFileComp => No file completion even when there are no other completions
_completionTests_verifyCompletion "testprog prefix nofile " "bear bearpaw dog unicorn" nofile
_completionTests_verifyCompletion "testprog prefix nofile u" "unicorn" nofile
_completionTests_verifyCompletion "testprog prefix nofile f" "" nofile
_completionTests_verifyCompletion "testprog prefix nofile z" "" nofile

# Test ShellCompDirectiveNoSpace => No space even when there is a single completion
_completionTests_verifyCompletion "testprog prefix nospace " "bear bearpaw dog unicorn" nospace
_completionTests_verifyCompletion "testprog prefix nospace b" "bear bearpaw" nospace
_completionTests_verifyCompletion "testprog prefix nospace u" "unicorn" nospace
_completionTests_verifyCompletion "testprog prefix nospace f" "" nospace
_completionTests_verifyCompletion "testprog prefix nospace z" "" nospace
_completionTests_verifyCompletion "testprog prefix nofilenospace " "bear bearpaw dog unicorn" nofile nospace
_completionTests_verifyCompletion "testprog prefix nofilenospace b" "bear bearpaw" nofile nospace
_completionTests_verifyCompletion "testprog prefix nofilenospace u" "unicorn" nofile nospace
_completionTests_verifyCompletion "testprog prefix nofilenospace f" "" nofile nospace
_completionTests_verifyCompletion "testprog prefix nofilenospace z" "" nofile nospace

#################################################
# Completions are NOT filtered by prefix by the program
#################################################

# Test ShellCompDirectiveDefault => File completion when no other completions
_completionTests_verifyCompletion "testprog noprefix default u" "unicorn"
_completionTests_verifyCompletion "testprog noprefix default f" ""
_completionTests_verifyCompletion "testprog noprefix default z" ""

# Test ShellCompDirectiveNoFileComp => No file completion even when there are no other completions
_completionTests_verifyCompletion "testprog noprefix nofile u" "unicorn" nofile
_completionTests_verifyCompletion "testprog noprefix nofile f" "" nofile
_completionTests_verifyCompletion "testprog noprefix nofile z" "" nofile

# Test ShellCompDirectiveNoSpace => No space even when there is a single completion
_completionTests_verifyCompletion "testprog noprefix nospace b" "bear bearpaw" nospace
_completionTests_verifyCompletion "testprog noprefix nospace u" "unicorn" nospace
_completionTests_verifyCompletion "testprog noprefix nospace f" "" nospace
_completionTests_verifyCompletion "testprog noprefix nospace z" "" nospace
_completionTests_verifyCompletion "testprog noprefix nofilenospace b" "bear bearpaw" nofile nospace
_completionTests_verifyCompletion "testprog noprefix nofilenospace u" "unicorn" nofile nospace
_completionTests_verifyCompletion "testprog noprefix nofilenospace f" "" nofile nospace
_completionTests_verifyCompletion "testprog noprefix nofilenospace z" "" nofile nospace

#################################################
# Other directives
#################################################
# Test ShellCompDirectiveFilterFileExt
_completionTests_verifyCompletion "testprog fileext setup" "setup.json setup.yaml"

# Test ShellCompDirectiveFilterDirs
# TODO these are broken, needs to be fixed.
_completionTests_verifyCompletion "testprog dir di" "dir dir2"
_completionTests_verifyCompletion "testprog subdir " "jsondir txtdir yamldir"
_completionTests_verifyCompletion "testprog subdir j" "jsondir"
_completionTests_verifyCompletion "testprog --theme " "jsondir txtdir yamldir"
_completionTests_verifyCompletion "testprog --theme t" "txtdir"
_completionTests_verifyCompletion "testprog --theme=" "jsondir txtdir yamldir"
_completionTests_verifyCompletion "testprog --theme=t" "txtdir"

# Test ShellCompDirectiveError => File completion only
_completionTests_verifyCompletion "testprog error u" ""

#################################################
# Flags
#################################################
_completionTests_verifyCompletion "testprog --custom" "--customComp" nofile
_completionTests_verifyCompletion "testprog --customComp " "firstComp secondComp forthComp" nofile
_completionTests_verifyCompletion "testprog --customComp f" "firstComp forthComp" nofile
_completionTests_verifyCompletion "testprog --customComp=" "firstComp secondComp forthComp" nofile
_completionTests_verifyCompletion "testprog --customComp=f" "firstComp forthComp" nofile

#################################################
# Special cases
#################################################
# Test when there is a space before the binary name
# https://github.com/spf13/cobra/issues/1303
_completionTests_verifyCompletion " testprog prefix default u" "unicorn"

# Test using env variable and ~
# https://github.com/spf13/cobra/issues/1306
OLD_HOME=$HOME
HOME="$(mktemp -d)"
cp "$ROOTDIR/testprog/bin/testprog" "$HOME/"
# Must use single quotes to keep the environment variable
_completionTests_verifyCompletion "\$HOME/testprog prefix default u" "unicorn"
_completionTests_verifyCompletion "~/testprog prefix default u" "unicorn"
rm "$HOME/testprog"
rmdir "$HOME"
HOME=$OLD_HOME

# An argument starting with dashes
_completionTests_verifyCompletion "testprog dasharg " "--arg"
# Needs bash completion v2
_completionTests_verifyCompletion "testprog dasharg -- --" "--arg"

# Test debug printouts
verifyDebug

# Test completion with a redirection
# https://github.com/spf13/cobra/issues/1334
if (("${BASH_VERSINFO[0]}" != 3)); then
  # We know and accept that this fails with bash 3
  # https://github.com/spf13/cobra/issues/1334
  verifyRedirect
fi

# Measure speed of execution without descriptions (for both v1 and v2)
_completionTests_timing "testprog manycomps " 0.2 "no descriptions"

# Test other bash completion types with descriptions disabled.
# There should be no change in behaviour when there are no descriptions.
# The types are: menu-complete/menu-complete-backward (COMP_TYPE == 37)
# and insert-completions (COMP_TYPE == 42)
COMP_TYPE=37
_completionTests_verifyCompletion "testprog prefix nospace b" "bear bearpaw" nospace
_completionTests_verifyCompletion "testprog prefix nofile b" "bear bearpaw" nofile

# Measure speed of execution with menu-complete without descriptions
_completionTests_timing "testprog manycomps " 0.2 "menu-complete no descs"

COMP_TYPE=42
_completionTests_verifyCompletion "testprog prefix nospace b" "bear bearpaw" nospace
_completionTests_verifyCompletion "testprog prefix nofile b" "bear bearpaw" nofile

# Measure speed of execution with insert-completions without descriptions (for both v1 and v2)
_completionTests_timing "testprog manycomps " 0.2 "insert-completions no descs"

unset COMP_TYPE

# Setup completion of testprog, enabling descriptions for v2.
# Don't use the new source <() form as it does not work with bash v3.
# Normally, compopt is a builtin, and the script checks that it is a
# builtin to disable it if we are in bash3 (where compopt does not exist).
# We replace 'builtin' with 'function' because we cannot use the native
# compopt since we are explicitly calling the completion code instead
# of from within a real completion environment.
# shellcheck source=/dev/null
source /dev/stdin <<-EOF
   $(testprog completion bash --descriptions | sed s/builtin/function/g)
EOF

# Check disabled because it's used in comp-test-lib.bash
# shellcheck disable=SC2034
# Disable sorting of output because it would mix up the descriptions.
BASH_COMP_NO_SORT=1

# When running docker without the --tty/-t flag, the COLUMNS variable is not set.
# bash completion needs it to handle descriptions, so we set it here if it is unset.
COLUMNS=${COLUMNS-100}

# Test descriptions with ShellCompDirectiveDefault
_completionTests_verifyCompletion "testprog prefix default " "bear     (an animal) bearpaw  (a dessert) dog unicorn  (mythical)"
_completionTests_verifyCompletion "testprog prefix default b" "bear     (an animal) bearpaw  (a dessert)"
_completionTests_verifyCompletion "testprog prefix default bearp" "bearpaw"

# Test descriptions with ShellCompDirectiveNoFileComp
_completionTests_verifyCompletion "testprog prefix nofile " "bear     (an animal) bearpaw  (a dessert) dog unicorn  (mythical)" nofile
_completionTests_verifyCompletion "testprog prefix nofile b" "bear     (an animal) bearpaw  (a dessert)" nofile
_completionTests_verifyCompletion "testprog prefix nofile bearp" "bearpaw" nofile

# Test descriptions with ShellCompDirectiveNoSpace
_completionTests_verifyCompletion "testprog prefix nospace " "bear     (an animal) bearpaw  (a dessert) dog unicorn  (mythical)" nospace
_completionTests_verifyCompletion "testprog prefix nospace b" "bear     (an animal) bearpaw  (a dessert)" nospace
_completionTests_verifyCompletion "testprog prefix nospace bearp" "bearpaw" nospace

# Test descriptions with completion of flag values
_completionTests_verifyCompletion "testprog --customComp " "firstComp   (the first value) secondComp  (the second value) forthComp" nofile
_completionTests_verifyCompletion "testprog --customComp f" "firstComp  (the first value) forthComp" nofile
_completionTests_verifyCompletion "testprog --customComp fi" "firstComp" nofile

# Measure speed of execution with descriptions
_completionTests_timing "testprog manycomps " 0.5 "with descriptions"

# Test descriptions are properly removed when using other bash completion types
# The types are: menu-complete/menu-complete-backward (COMP_TYPE == 37)
# and insert-completions (COMP_TYPE == 42)
COMP_TYPE=37
_completionTests_verifyCompletion "testprog prefix nospace b" "bear bearpaw" nospace
_completionTests_verifyCompletion "testprog prefix nofile b" "bear bearpaw" nofile

# Measure speed of execution with menu-complete with descriptions
_completionTests_timing "testprog manycomps " 0.2 "menu-complete with descs"

COMP_TYPE=42
_completionTests_verifyCompletion "testprog prefix nospace b" "bear bearpaw" nospace
_completionTests_verifyCompletion "testprog prefix nofile b" "bear bearpaw" nofile

# Measure speed of execution with insert-completions with descriptions
_completionTests_timing "testprog manycomps " 0.2 "insert-completions no descs"

unset COMP_TYPE

# This must be the last call.  It allows to exit with an exit code
# that reflects the final status of all the tests.
_completionTests_exit
