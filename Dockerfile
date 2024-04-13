# First stage. Building a binary
# -----------------------------------------------------------------------------
FROM golang:1.22-alpine@sha256:cdc86d9f363e8786845bea2040312b4efa321b828acdeb26f393faa864d887b0 AS builder

COPY . /src
WORKDIR /src
RUN go mod tidy && go mod vendor
RUN go build threadfin.go

# Second stage. Creating an image
# -----------------------------------------------------------------------------
FROM ubuntu:23.10@sha256:5cd569b792a8b7b483d90942381cd7e0b03f0a15520d6e23fb7a1464a25a71b1

ARG BUILD_DATE
ARG VCS_REF
ARG THREADFIN_PORT=34400
ARG THREADFIN_VERSION
# http://stackoverflow.com/questions/48162574/ddg#49462622
ARG APT_KEY_DONT_WARN_ON_DANGEROUS_USAGE=DontWarn

LABEL org.label-schema.build-date="{$BUILD_DATE}" \
      org.label-schema.name="Threadfin" \
      org.label-schema.description="Dockerized Threadfin" \
      org.label-schema.url="https://hub.docker.com/r/volschin/threadfin/" \
      org.label-schema.vcs-ref="{$VCS_REF}" \
      org.label-schema.vcs-url="https://github.com/Threadfin/Threadfin" \
      org.label-schema.vendor="Threadfin" \
      org.label-schema.version="{$THREADFIN_VERSION}" \
      org.label-schema.schema-version="1.0"

ENV THREADFIN_BIN=/home/threadfin/bin
ENV THREADFIN_CONF=/home/threadfin/conf
ENV THREADFIN_HOME=/home/threadfin
ENV THREADFIN_TEMP=/tmp/threadfin
ENV THREADFIN_CACHE=/home/threadfin/cache
ENV THREADFIN_UID=31337
ENV THREADFIN_GID=31337
ENV THREADFIN_USER=threadfin
ENV THREADFIN_BRANCH=main
ENV THREADFIN_DEBUG=0
ENV THREADFIN_PORT=34400
ENV THREADFIN_LOG=/var/log/threadfin.log

# Add binary to PATH
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$THREADFIN_BIN
# https://askubuntu.com/questions/972516/debian-frontend-environment-variable
ENV DEBIAN_FRONTEND=noninteractive

# Set working directory
WORKDIR $THREADFIN_HOME

# Install dependencies:
# mesa-va-drivers: needed for AMD VAAPI. Mesa >= 20.1 is required for HEVC transcoding.
# curl: healthcheck
RUN apt-get -qqy update \
 && apt-get -qqy install --no-install-recommends --no-install-suggests \
   ca-certificates \
   curl \
   vlc \
   openssl \
   locales \

# && dpkg -i *.deb \
# && cd .. \
# && rm -rf intel-compute-runtime \
 && apt-get -qqy autoremove \
 && apt-get -qqy clean autoclean \
 && rm -rf /var/lib/apt/lists/* \
 && mkdir -p /cache /config /media \
 && chmod 777 /cache /config /media \
 && sed -i -e 's/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen && locale-gen \
 && sed -i 's/geteuid/getppid/' /usr/bin/vlc \
 && mkdir -p $THREADFIN_BIN

# ENV DOTNET_SYSTEM_GLOBALIZATION_INVARIANT=1
ENV LC_ALL en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en

# Copy built binary from builder image
COPY --chown=${THREADFIN_UID} --from=builder ["/src/threadfin", "${THREADFIN_BIN}/"]
COPY --from=ghcr.io/volschin/ffmpeg-static:main ["/download/ffmpeg", "${THREADFIN_BIN}/"]
#COPY --from=ghcr.io/linuxserver/mods:jellyfin-opencl-intel ["/opencl-intel", "/opencl-intel"]
# Set binary permissions and create working directories for Threadfin
RUN chmod +rx $THREADFIN_BIN/threadfin \
  && mkdir $THREADFIN_HOME/cache \
  && mkdir $THREADFIN_CONF \
  && chmod a+rwX $THREADFIN_CONF \
  && mkdir $THREADFIN_TEMP \
  && chmod a+rwX $THREADFIN_TEMP

# volume mappings
VOLUME [ "$THREADFIN_CONF", "$THREADFIN_TEMP" ]
EXPOSE $THREADFIN_PORT

# Run the Threadfin executable
#ENTRYPOINT [ "${THREADFIN_BIN}/threadfin", "-port=${THREADFIN_PORT}", "-config=${THREADFIN_CONF}", "-debug=${THREADFIN_DEBUG}" ]
ENTRYPOINT ${THREADFIN_BIN}/threadfin -port=${THREADFIN_PORT} -config=${THREADFIN_CONF} -debug=${THREADFIN_DEBUG}
