# First stage. Building a binary
# -----------------------------------------------------------------------------
FROM golang:1.20 AS builder

# Download the source code
RUN apt-get update && apt-get install -y git
RUN git clone https://github.com/Threadfin/Threadfin.git /src

WORKDIR /src

RUN git checkout main
RUN git pull
RUN go mod tidy && go mod vendor
RUN go build threadfin.go

# Second stage. Creating an image
# -----------------------------------------------------------------------------
FROM ubuntu:mantic-20240122

ARG BUILD_DATE
ARG VCS_REF
ARG THREADFIN_PORT=34400
ARG THREADFIN_VERSION
# https://askubuntu.com/questions/972516/debian-frontend-environment-variable
ARG DEBIAN_FRONTEND="noninteractive"
# http://stackoverflow.com/questions/48162574/ddg#49462622
ARG APT_KEY_DONT_WARN_ON_DANGEROUS_USAGE=DontWarn
# https://github.com/NVIDIA/nvidia-docker/wiki/Installation-(Native-GPU-Support)
ENV NVIDIA_DRIVER_CAPABILITIES="compute,video,utility"

# https://github.com/intel/compute-runtime/releases
ARG GMMLIB_VERSION=22.3.11.ci17747749
ARG IGC_VERSION=1.0.15136.4
ARG NEO_VERSION=23.35.27191.9
ARG LEVEL_ZERO_VERSION=1.3.27191.9

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

# Set working directory
WORKDIR $THREADFIN_HOME

# Install dependencies:
# mesa-va-drivers: needed for AMD VAAPI. Mesa >= 20.1 is required for HEVC transcoding.
# curl: healthcheck
RUN apt-get update \
 && apt-get install --no-install-recommends --no-install-suggests -y ca-certificates gnupg wget curl vlc \
 && wget -O - https://repo.jellyfin.org/jellyfin_team.gpg.key | apt-key add - \
# Pinned to mantic until noble is supported
 && echo "deb [arch=$( dpkg --print-architecture )] https://repo.jellyfin.org/$( awk -F'=' '/^ID=/{ print $NF }' /etc/os-release ) mantic main" | tee /etc/apt/sources.list.d/jellyfin.list \
 && apt-get update \
 && apt-get install --no-install-recommends --no-install-suggests -y \
   mesa-va-drivers \
   jellyfin-ffmpeg6 \
   openssl \
   locales \
# Intel VAAPI Tone mapping dependencies:
# Prefer NEO to Beignet since the latter one doesn't support Comet Lake or newer for now.
# Do not use the intel-opencl-icd package from repo since they will not build with RELEASE_WITH_REGKEYS enabled.
 && mkdir intel-compute-runtime \
 && cd intel-compute-runtime \
 && wget https://github.com/intel/compute-runtime/releases/download/${NEO_VERSION}/libigdgmm12_${GMMLIB_VERSION}_amd64.deb \
 && wget https://github.com/intel/intel-graphics-compiler/releases/download/igc-${IGC_VERSION}/intel-igc-core_${IGC_VERSION}_amd64.deb \
 && wget https://github.com/intel/intel-graphics-compiler/releases/download/igc-${IGC_VERSION}/intel-igc-opencl_${IGC_VERSION}_amd64.deb \
 && wget https://github.com/intel/compute-runtime/releases/download/${NEO_VERSION}/intel-opencl-icd_${NEO_VERSION}_amd64.deb \
 && wget https://github.com/intel/compute-runtime/releases/download/${NEO_VERSION}/intel-level-zero-gpu_${LEVEL_ZERO_VERSION}_amd64.deb \
# && wget https://github.com/intel/compute-runtime/releases/download/${NEO_VERSION}/ww35.sum && sha256sum -c ww35.sum \
 && dpkg -i *.deb \
 && cd .. \
 && rm -rf intel-compute-runtime \
 && apt-get remove gnupg wget -y \
 && apt-get clean autoclean -y \
 && apt-get autoremove -y \
 && rm -rf /var/lib/apt/lists/* \
 && mkdir -p /cache /config /media \
 && chmod 777 /cache /config /media \
 && sed -i -e 's/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen && locale-gen \
 && sed -i 's/geteuid/getppid/' /usr/bin/vlc

# ENV DOTNET_SYSTEM_GLOBALIZATION_INVARIANT=1
ENV LC_ALL en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en

RUN mkdir -p $THREADFIN_BIN
# Copy built binary from builder image
COPY --chown=${THREADFIN_UID} --from=builder [ "/src/threadfin", "${THREADFIN_BIN}/" ]

# Set binary permissions and create working directories for Threadfin
RUN chmod +rx $THREADFIN_BIN/threadfin \
  && mkdir $THREADFIN_HOME/cache \
  && mkdir $THREADFIN_CONF \
  && chmod a+rwX $THREADFIN_CONF \
  && mkdir $THREADFIN_TEMP \
  && chmod a+rwX $THREADFIN_TEMP

# volume mappings
VOLUME $THREADFIN_CONF $THREADFIN_TEMP

EXPOSE $THREADFIN_PORT

# Run the Threadfin executable
ENTRYPOINT ${THREADFIN_BIN}/threadfin -port=${THREADFIN_PORT} -config=${THREADFIN_CONF} -debug=${THREADFIN_DEBUG}
