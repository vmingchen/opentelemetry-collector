#!/bin/bash

set -e

SED="sed"

PASS_COLOR=$(printf "\033[32mPASS\033[0m")
FAIL_COLOR=$(printf "\033[31mFAIL\033[0m")
TEST_COLORIZE="${SED} 's/PASS/${PASS_COLOR}/' | ${SED} 's/FAIL/${FAIL_COLOR}/'"
echo ${TEST_ARGS}
mkdir -p results
TESTBED_CONFIG=inprocess.yaml go test -v ${TEST_ARGS} 2>&1 | tee results/testoutput.log | bash -c "${TEST_COLORIZE}"

testStatus=${PIPESTATUS[0]}

mkdir -p results/junit
go-junit-report < results/testoutput.log > results/junit/results.xml

bash -c "cat results/CORRECTNESSRESULTS.md | ${TEST_COLORIZE}"

exit ${testStatus}
