FROM ghcr.io/astral-sh/uv:latest AS uv
FROM lscr.io/linuxserver/ffmpeg:8.0.1

# Install system dependencies
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates

# Install uv from the uv image
COPY --from=uv /uv /uvx /bin/

# Use uv to install yt-dlp as a tool
# Install yt-dlp to a location readable by all users (since we run as a non-root user)
# and unset VIRTUAL_ENV to avoid conflicts with the base image's environment (lsiopy).
ENV PATH="/usr/local/bin:$PATH"
ENV UV_TOOL_DIR=/opt/uv-tools
ENV UV_TOOL_BIN_DIR=/usr/local/bin
ENV UV_PYTHON_INSTALL_DIR=/opt/uv-python
ENV VIRTUAL_ENV=""

RUN uv tool install yt-dlp==2026.02.04 && \
    chmod -R a+rX /opt/uv-tools /opt/uv-python

# Set HOME to a writable location for any user
ENV HOME=/tmp

# Set entrypoint to match standard yt-dlp usage
ENTRYPOINT ["yt-dlp"]
CMD ["--help"]
