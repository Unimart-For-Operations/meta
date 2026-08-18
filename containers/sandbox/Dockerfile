FROM nixos/nix:latest AS build
COPY flake.nix flake.lock /sandbox/
WORKDIR /sandbox
RUN rm -f /root/.nix-profile \
    && nix --extra-experimental-features 'nix-command flakes' build .#ttyProfile --no-write-lock-file --out-link /root/.nix-profile

FROM nixos/nix:latest
COPY --from=build /nix /nix
RUN rm -rf /root/.nix-profile
COPY --from=build /root/.nix-profile /root/.nix-profile

ENV PATH=/root/.nix-profile/bin:/root/.local/bin:/nix/var/nix/profiles/default/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV SHELL=/root/.nix-profile/bin/zsh

RUN ln -sf /root/.nix-profile/bin/zsh /bin/zsh \
    && mkdir -p /root/.config /root/.config/tmux /root/.config/opencode /root/.local/bin

COPY rc/zshrc /root/.zshrc
COPY rc/tmux.conf /root/.config/tmux/tmux.conf
COPY rc/starship.toml /root/.config/starship.toml
COPY rc/opencode/opencode.json /root/.config/opencode/opencode.json

RUN printf 'source-file ~/.config/tmux/tmux.conf\n' > /root/.tmux.conf

COPY hm/nvim-astro /root/.config/nvim-astro

RUN printf '#!/usr/bin/env bash\nexec env NVIM_APPNAME=nvim-astro /root/.nix-profile/bin/nvim "$@"\n' > /root/.local/bin/nvim-astro \
    && printf '#!/usr/bin/env bash\nexec "$(dirname "$0")/nvim-astro" "$@"\n' > /root/.local/bin/nvim \
    && chmod +x /root/.local/bin/nvim /root/.local/bin/nvim-astro

CMD ["sleep", "infinity"]