#!/bin/bash

COVERAGE=$(cat cover.html | grep "%)"  | sed 's/[][()><%]/ /g' | awk '{ print $4 }' | awk '{s+=$1}END{print s/NR}')

echo "-------------------------------------------------------------------------"
echo "COVERAGE IS ${COVERAGE}%"
echo "-------------------------------------------------------------------------"
