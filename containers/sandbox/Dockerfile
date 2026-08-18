FROM nixos/nix:latest AS build
COPY flake.nix flake.lock /sandbox/
WORKDIR /sandbox
RUN rm -f /root/.nix-profile \
    && nix --extra-experimental-features 'nix-command flakes' build .#ttyProfile --no-write-lock-file --out-link /root/.nix-profile

FROM nixos/nix:latest
COPY --from=build /nix /nix
RUN rm -rf /root/.nix-profile
COPY --from=build /root/.nix-profile /root/.nix-profile

ENV PATH=/root/.nix-profile/bin:/nix/var/nix/profiles/default/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV SHELL=/root/.nix-profile/bin/zsh

RUN ln -sf /root/.nix-profile/bin/zsh /bin/zsh \
    && mkdir -p /root/.config

COPY rc/zshrc /root/.zshrc
COPY rc/tmux.conf /root/.tmux.conf
COPY rc/starship.toml /root/.config/starship.toml

CMD ["sleep", "infinity"]