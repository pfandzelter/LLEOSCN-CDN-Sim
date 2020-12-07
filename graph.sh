#!/bin/sh

./graph/graph "$1"
python3 ./graph/graph.py "$1"