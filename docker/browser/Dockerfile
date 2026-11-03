# The sandboxed browser (ADR 0002): agent-browser driving Chrome for Testing,
# headless, as a non-root user. Every version is exact; never use latest.
FROM node:24.22.0-bookworm-slim@sha256:91acba91b842bec0bf7ea6f5e09c644c7ccad72c4996ed9616d861d1c7cd60d4

ARG AGENT_BROWSER_VERSION=0.39.2
# Chrome for Testing build and the sha256 of its linux64 zip. Bump both
# together: `agent-browser install` would fetch whatever is newest instead.
ARG CHROME_VERSION=156.0.8078.80
ARG CHROME_SHA256=b711c720f774be355197f32edf37e0651418ba036d92446ee2da19ac017540a7

# Chrome's shared libraries (the list `agent-browser install --with-deps`
# uses), plus curl for `caddie doctor` and unzip for the Chrome download.
RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    ca-certificates curl unzip fonts-freefont-ttf \
    libasound2 libatk-bridge2.0-0 libatk1.0-0 libatspi2.0-0 libcairo-gobject2 \
    libcairo2 libcups2 libdbus-1-3 libdrm2 libfontconfig1 libfreetype6 libgbm1 \
    libgdk-pixbuf-2.0-0 libgtk-3-0 libnspr4 libnss3 libnss3-tools libpango-1.0-0 \
    libpangocairo-1.0-0 libx11-6 libx11-xcb1 libxcb-shm0 libxcb1 libxcomposite1 \
    libxcursor1 libxdamage1 libxext6 libxfixes3 libxi6 libxkbcommon0 libxrandr2 \
    libxrender1 libxshmfence1 \
  && rm -rf /var/lib/apt/lists/*

RUN curl -fsSLo /tmp/chrome.zip \
    "https://storage.googleapis.com/chrome-for-testing-public/${CHROME_VERSION}/linux64/chrome-linux64.zip" \
  && echo "${CHROME_SHA256}  /tmp/chrome.zip" | sha256sum -c - \
  && unzip -q /tmp/chrome.zip -d /opt \
  && mv /opt/chrome-linux64 /opt/chrome \
  && rm /tmp/chrome.zip

RUN npm install -g --no-fund --no-audit "agent-browser@${AGENT_BROWSER_VERSION}" \
  && npm cache clean --force

# The node user (uid 1000) owns ~/.agent-browser, created 0700 so the named
# volume mounted there starts with the same owner and mode.
RUN install -d -m 0700 -o node -g node /home/node/.agent-browser

USER node
WORKDIR /home/node
ENV AGENT_BROWSER_EXECUTABLE_PATH=/opt/chrome/chrome

# Stay up so `docker compose exec browser agent-browser …` has a container;
# agent-browser starts its own daemon on the first command.
CMD ["sleep", "infinity"]
