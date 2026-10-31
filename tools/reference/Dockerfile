# syntax=docker/dockerfile:1.7
#
# Reference image for the conformance suite: Graphviz 16.1.0 built from its
# release sources without Pango, so that text width comes from Graphviz's
# built-in font metrics and does not depend on installed fonts, and a pinned
# resvg that rasterizes SVG for the similarity comparison. Every download and
# base image is pinned by checksum or digest.
#
# Build for linux/amd64 only, so that references do not depend on the host:
#
#   docker build --platform linux/amd64 -t gogv-reference tools/reference

ARG BASE_IMAGE=debian:trixie-slim@sha256:f6b08b09a05725f9431502d3b5714f5c4576e82c1a84fa1debe0c24a7e0c3ec4

FROM ${BASE_IMAGE} AS build
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential pkg-config bison flex libexpat1-dev zlib1g-dev libltdl-dev \
 && rm -rf /var/lib/apt/lists/*
ADD --checksum=sha256:0f661718f3e5268dc3bad0fb53ab646605d26db70f4080799aff1dc2f61783c3 \
    https://gitlab.com/api/v4/projects/4207231/packages/generic/graphviz-releases/16.1.0/graphviz-16.1.0.tar.xz \
    /tmp/graphviz.tar.xz
# No Pango, Cairo, GD or X11 development files are installed, and Pango is
# disabled explicitly: the only text layout is Graphviz's built-in estimate.
RUN tar -xJf /tmp/graphviz.tar.xz -C /tmp \
 && cd /tmp/graphviz-16.1.0 \
 && ./configure --prefix=/opt/graphviz --without-pangocairo --without-x --disable-static \
 && make -j"$(nproc)" \
 && make install

FROM ${BASE_IMAGE}
ENV LANG=C.UTF-8
RUN apt-get update \
 && apt-get install -y --no-install-recommends libexpat1 zlib1g libltdl7 \
 && rm -rf /var/lib/apt/lists/*
COPY --from=build /opt/graphviz /opt/graphviz
ENV PATH=/opt/graphviz/bin:$PATH
# Registers the plugins that were built (core renderers and the dot layout).
RUN dot -c

ADD --checksum=sha256:fa8c26495a187e592c501db15bf9e8a9fdc051d4b2b336b39703d5b59f912b9d \
    https://github.com/linebender/resvg/releases/download/v0.48.1/resvg-linux-x86_64.tar.gz \
    /tmp/resvg.tar.gz
RUN tar -xzf /tmp/resvg.tar.gz -C /usr/local/bin resvg && rm /tmp/resvg.tar.gz
