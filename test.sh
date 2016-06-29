#!/bin/sh

echo "./exslack commands1.txt"
./exslack commands1.txt
echo "./exslack -log joblog.txt commands1.txt"
./exslack -log joblog.txt commands1.txt
echo "./exslack -c -log joblog.txt commands1.txt"
./exslack -c -log joblog.txt commands1.txt
echo "./exslack -cpu 4 -log joblog.txt commands1.txt"
./exslack -cpu 4 -log joblog.txt commands1.txt
echo "./exslack -c -cpu 4 -log joblog.txt commands1.txt"
./exslack -c -cpu 4 -log joblog.txt commands1.txt
echo "./exslack -c -cpu 4 -log joblog.txt commands2.txt"
./exslack -c -cpu 4 -log joblog.txt commands2.txt
echo "./exslack -c -cpu 4 -log joblog.txt commands3.txt"
./exslack -c -cpu 4 -log joblog.txt commands3.txt
