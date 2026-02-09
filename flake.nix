{
  description = "Ebiten (Go) dev shell with X11/ALSA deps for Linux (Ubuntu 24.04 host)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            pkg-config

            # Audio (provides alsa.pc)
            alsa-lib

            # X11 headers Ebiten/GLFW expects
            xorg.libX11
            xorg.libXrandr
            xorg.libXinerama
            xorg.libXcursor
            xorg.libXi

            # OpenGL / GL headers
            mesa
            libGL
          ];

          # Helps some builds find pkg-config metadata
          shellHook = ''
            export CGO_ENABLED=1
            echo "Ebiten Linux deps loaded (X11 + ALSA + GL)."
            echo "Try: go build ./..."
          '';
        };
      });
}
