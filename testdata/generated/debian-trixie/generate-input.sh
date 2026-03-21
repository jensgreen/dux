#!/bin/bash

docker run --rm debian:trixie@sha256:3615a749858a1cba49b408fb49c37093db813321355a9ab7c1f9f4836341e9db /usr/bin/find -type f -printf '%s\t%P\n'
