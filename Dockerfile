FROM public.ecr.aws/docker/library/alpine:3.23.3 AS vim-builder

ENV BUILD_BASE_VERSION=0.5-r3
ENV NCURSES_VERSION=6.5_p20251123-r0
ENV GTK2_VERSION=2.24.33-r11
ENV LIBX11_VERSION=1.8.12-r1
ENV LIBXT_VERSION=1.3.1-r0
ENV GIT_VERSION=2.52.0-r0

ENV VIM_VERSION=9.2.0088

RUN <<EOF
apk add --no-cache \
    build-base=$BUILD_BASE_VERSION \
    ncurses-static=$NCURSES_VERSION \
    ncurses-dev=$NCURSES_VERSION \
    gtk+2.0-dev=$GTK2_VERSION \
    libx11-dev=$LIBX11_VERSION \
    libxt-dev=$LIBXT_VERSION \
    git=$GIT_VERSION

git clone --depth 1 --branch v${VIM_VERSION} https://github.com/vim/vim.git /vim
EOF

WORKDIR /vim

RUN <<EOF
./configure \
  --disable-channel \
  --disable-gpm \
  --disable-gtktest \
  --disable-gui \
  --disable-netbeans \
  --disable-nls \
  --disable-selinux \
  --disable-smack \
  --disable-sysmouse \
  --disable-xsmp \
  --enable-multibyte \
  --with-features=huge \
  --without-x \
  LDFLAGS="-static" \
  LIBS="-lncursesw -ltinfo"

make "-j$(nproc)"

file src/vim && ldd src/vim || true
EOF

FROM scratch AS vim
COPY --from=vim-builder /vim/src/vim /vim
ENTRYPOINT ["/vim", "-u", "NONE", "-N"]

FROM public.ecr.aws/docker/library/golang:1.26.1-trixie AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o edmotion ./cmd/edmotion

FROM scratch AS edmotion

ENV VIM_PATH=/vim

COPY --from=builder /app/edmotion /edmotion
COPY --from=vim /vim /vim
WORKDIR /tmp

ENTRYPOINT ["/edmotion"]
