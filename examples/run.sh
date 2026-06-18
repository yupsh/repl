#!/bin/bash
# Run the entire yupsh example suite through the yupsh binary.
#
# Each examples/*.yup file is a self-documenting yupsh session: comment lines
# (starting with #) and blank lines are ignored by the REPL, and every command
# line is fed to the binary on stdin. This runner builds yupsh (unless YUPSH
# points at an existing binary), sets up a deterministic fixture tree so the
# filesystem examples are reproducible, runs every example, prints each cleaned
# transcript, and fails if any example writes to stderr.
#
# Usage:
#   examples/run.sh
# Where:
#   YUPSH is an existing yupsh binary to use instead of building one. Default:
#         a fresh build of ./cmd/yupsh into a temp directory.
#
# Future: once cmd-xargs can exec a command per argument group (today it only
# regroups fields into argument lines), this whole runner collapses into a
# yupsh-native pipeline that uses yupsh's OWN xargs to run the suite — roughly:
#   ls examples | grep '[.]yup$' | xargs -n 1 <feed each to yupsh>
# Until xargs gains exec, the per-example loop below drives the binary from bash.

set -o errexit
set -o nounset
set -o pipefail

exec 3>&1 4>&2

# Constants — anchored to this script's directory so the runner works from any cwd.
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${HERE}/.." && pwd)"
WORK_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/yupsh-examples.XXXXXX")"
FIXTURES="${WORK_ROOT}/work"
HOME_DIR="${WORK_ROOT}/home"
YUPSH=${YUPSH:-${WORK_ROOT}/yupsh}

trap 'rm -rf "${WORK_ROOT}"' EXIT

# build-binary builds yupsh into WORK_ROOT unless YUPSH already names a binary.
build-binary() {
  [[ -x ${YUPSH} ]] && {
    printf >&3 '▸ Using yupsh binary: %s\n' "${YUPSH}"
    return 0
  }
  printf >&3 '▸ Building yupsh from ./cmd/yupsh\n'
  go -C "${REPO_ROOT}" build -o "${YUPSH}" ./cmd/yupsh
}

# setup-fixtures seeds a deterministic work tree and home directory so the
# filesystem examples (08-files-and-globs) are reproducible.
setup-fixtures() {
  mkdir -p "${FIXTURES}/sub" "${HOME_DIR}"
  printf 'banana\napple\ncherry\napple\nbanana\n' >"${FIXTURES}/fruits.txt"
  printf 'roses are red\nviolets are blue\n' >"${FIXTURES}/poem.txt"
  printf 'id,name\n1,alice\n2,bob\n' >"${FIXTURES}/people.csv"
  printf 'nested file\n' >"${FIXTURES}/sub/note.md"
  printf 'welcome to your yup home\n' >"${HOME_DIR}/welcome.txt"
}

# run-example feeds one example file to yupsh and reports whether it ran clean.
# It prints the cleaned transcript (banner and prompts stripped) and returns
# non-zero when the example wrote anything to stderr.
run-example() {
  local file=${1}
  local name=${file##*/}
  local out="${WORK_ROOT}/${name}.out"
  local err="${WORK_ROOT}/${name}.err"

  printf >&3 '\n\033[36m▸ %s\033[0m\n' "${name}"
  ( cd "${FIXTURES}" && HOME="${HOME_DIR}" "${YUPSH}" <"${file}" ) >"${out}" 2>"${err}" || true

  # Strip the 4-line startup banner, every "yup> " prompt (skipped comment and
  # blank lines emit bare prompts that cluster on one line), drop the resulting
  # empty lines, and indent what the commands actually produced.
  tail -n +5 "${out}" \
    | sed -e 's/^\(yup> \)*//' \
    | sed -e '/^$/d' \
    | sed -e 's/^/    /'

  [[ -s ${err} ]] && {
    printf >&4 '  \033[31m✗ stderr:\033[0m\n'
    sed 's/^/    /' "${err}" >&4
    return 1
  }
  return 0
}

# main builds the binary, seeds fixtures, and runs every example, tracking
# failures so the suite is a health check as well as a showcase.
main() {
  build-binary
  setup-fixtures

  local total=0 failed=0
  for file in "${HERE}"/*.yup; do
    # perl examples need a perl interpreter; skip them when none is installed.
    [[ ${file} == *-perl.yup && -z $(command -v perl) ]] && {
      printf >&3 '\n▸ %s — skipped (perl not installed)\n' "${file##*/}"
      continue
    }
    ((total += 1))
    run-example "${file}" || ((failed += 1))
  done

  printf >&3 '\n'
  ((failed == 0)) && {
    printf >&3 '\033[32m✓ all %d examples ran clean\033[0m\n' "${total}"
    return 0
  }
  printf >&4 '\033[31m✗ %d of %d examples wrote to stderr\033[0m\n' "${failed}" "${total}"
  return 1
}

main || exit 1
exit 0
