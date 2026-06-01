#!/usr/bin/env bash
set -eEuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE_BUILD_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${IMAGE_BUILD_DIR}/.." && pwd)"

RUN_LABEL="${1:-baseline}"
shift || true

PROFILE_INTERVAL="${PROFILE_INTERVAL:-1}"
RUN_ID="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="${IMAGE_BUILD_DIR}/perf/${RUN_LABEL}-${RUN_ID}"

if [[ $# -eq 0 ]]; then
  BUILD_CMD=("./build.sh")
else
  BUILD_CMD=("$@")
fi

mkdir -p "${OUT_DIR}"

PROFILER_PIDS=()

cleanup() {
  for pid in "${PROFILER_PIDS[@]:-}"; do
    kill "${pid}" 2>/dev/null || true
  done
}
trap cleanup EXIT

write_context() {
  {
    echo "run_id=${RUN_ID}"
    echo "run_label=${RUN_LABEL}"
    echo "date=$(date -Iseconds)"
    echo "profile_interval_sec=${PROFILE_INTERVAL}"
    echo "build_command=${BUILD_CMD[*]}"
    echo
    echo "git_commit=$(git -C "${REPO_ROOT}" rev-parse HEAD 2>/dev/null || echo unknown)"
    echo "git_branch=$(git -C "${REPO_ROOT}" rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)"
    echo "git_status_short:"
    git -C "${REPO_ROOT}" status --short 2>/dev/null || true
    echo
    echo "kernel:"
    uname -a || true
    echo
    echo "cpu:"
    lscpu || true
    echo
    echo "memory:"
    free -h || true
    echo
    echo "block_devices:"
    lsblk -o NAME,ROTA,TYPE,SIZE,MOUNTPOINT || true
    echo
    echo "filesystem:"
    df -h || true
    echo
    echo "docker_info:"
    docker info || true
  } >"${OUT_DIR}/context.txt" 2>&1
}

start_profiler() {
  local name="$1"
  local cmd="$2"

  if ! command -v "${name}" >/dev/null 2>&1; then
    echo "warning: ${name} not found, skipping" >>"${OUT_DIR}/profilers.log"
    return 0
  fi

  bash -lc "${cmd}" >"${OUT_DIR}/${name}.log" 2>&1 &
  PROFILER_PIDS+=("$!")
}

start_docker_stats() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "warning: docker not found, skipping docker stats" >>"${OUT_DIR}/profilers.log"
    return 0
  fi

  (
    while true; do
      date -Iseconds
      docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.BlockIO}}"
      sleep "${PROFILE_INTERVAL}"
    done
  ) >"${OUT_DIR}/docker-stats.log" 2>&1 &
  PROFILER_PIDS+=("$!")
}

write_context

echo "starting profilers" >"${OUT_DIR}/profilers.log"
start_profiler iostat "iostat -dxm ${PROFILE_INTERVAL}"
start_profiler pidstat "pidstat -dru -h ${PROFILE_INTERVAL}"
start_profiler vmstat "vmstat ${PROFILE_INTERVAL}"
start_docker_stats

printf '%s\n' "${PROFILER_PIDS[@]:-}" >"${OUT_DIR}/profiler-pids.txt"

BUILD_EXIT=0
START_TS="$(date +%s)"

pushd "${IMAGE_BUILD_DIR}" >/dev/null
set +e
/usr/bin/time -v "${BUILD_CMD[@]}" >"${OUT_DIR}/build.stdout.log" 2>"${OUT_DIR}/build.stderr.log"
BUILD_EXIT=$?
set -e
popd >/dev/null

END_TS="$(date +%s)"
echo "${BUILD_EXIT}" >"${OUT_DIR}/build.exit_code"

{
  echo "run_id=${RUN_ID}"
  echo "run_label=${RUN_LABEL}"
  echo "out_dir=${OUT_DIR}"
  echo "exit_code=${BUILD_EXIT}"
  echo "wall_seconds=$((END_TS - START_TS))"
  echo
  echo "time_summary:"
  grep -E "Elapsed|User time|System time|Percent of CPU|Maximum resident set size|File system inputs|File system outputs" "${OUT_DIR}/build.stderr.log" || true
} >"${OUT_DIR}/summary.txt"

echo "benchmark output saved to: ${OUT_DIR}"
exit "${BUILD_EXIT}"
