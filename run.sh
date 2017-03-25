#!/usr/bin/env bash

go build -o personal-website

echo "Minifying CSS"
minify \
	resources/css/base.css \
    --output resources/min.css

echo "Minifying JS"
minify \
	resources/js/lib/jquery-3.2.0.min.js \
	resources/js/base.js \
    --output resources/min.js

echo "================"
echo "Starting Website"
./personal-website 2>&1 | tee -a ./log/personal-website.log