# Stage 1: Build the Home Manager activation package from cmdr modules.
#
# The flake imports cmdr's tty + tui graduated modules plus cherry-picked
# CLI modules (atuin, direnv, fonts, opencode).  The activation package
# contains every binary, config file, and shell integration that
# home-manager would install on a real workstation.
FROM nixos/nix:latest AS build

# Enable flakes globally for the build
RUN echo 'experimental-features = nix-command flakes' >> /etc/nix/nix.conf

# Copy the flake and cmdr modules into a single flake root.
# The flake references ./cmdr-home so everything stays in-tree.
COPY containers/sandbox/flake.nix /build/flake.nix
COPY cmdr/home /build/cmdr-home

WORKDIR /build

# Build the activation package.
RUN nix build .#activationPackage \
      --no-write-lock-file \
      --out-link /root/.hm-activation \
    && echo "activation package built"

# Stage 2: Activate into a clean image.
FROM nixos/nix:latest

# Pull the Nix store from the build stage (contains all closures)
COPY --from=build /nix /nix

# Copy the activation package symlink
COPY --from=build /root/.hm-activation /root/.hm-activation

# Run the activation script — this populates /root with dotfiles,
# shell configs, nvim configs, tmux configs, etc.
# HOME must be set for home-manager activation to target the right dir.
#
# The nixos/nix base image ships a default user profile with packages
# (man-db, gzip, etc.) that conflict with home-manager-path.  Nuke the
# user profile entirely but keep /nix/var/nix/profiles/default (which
# contains nix-build, nix-env, etc.) on PATH so activation can run.
ENV HOME=/root
ENV PATH=/nix/var/nix/profiles/default/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN rm -f /root/.nix-profile \
    && nix-env --uninstall '.*' 2>/dev/null || true
RUN /root/.hm-activation/activate

# Ensure zsh is findable as a login shell
RUN ln -sf "$(readlink -f /root/.nix-profile/bin/zsh 2>/dev/null || echo /nix/var/nix/profiles/per-user/root/profile/bin/zsh)" /bin/zsh 2>/dev/null; \
    [ -x /bin/zsh ] || ln -sf "$(find /nix/store -name zsh -type f -executable 2>/dev/null | head -1)" /bin/zsh

ENV SHELL=/bin/zsh
ENV PATH=/root/.nix-profile/bin:/root/.local/bin:/nix/var/nix/profiles/default/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

CMD ["sleep", "infinity"]
