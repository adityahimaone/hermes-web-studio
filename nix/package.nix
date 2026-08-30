{ pkgs ? import <nixpkgs> {} }:

pkgs.stdenv.mkDerivation {
  pname = "hermes-web-studio";
  version = "0.1.0";
  src = ../.;
  nativeBuildInputs = [ pkgs.go pkgs.nodejs pkgs.pnpm ];
  buildPhase = ''
    export HOME=$(mktemp -d)
    pnpm --dir frontend install --frozen-lockfile
    pnpm --dir frontend build
    cp -R frontend/dist/. backend/internal/web/dist/
    (cd backend && go build -trimpath -ldflags='-s -w' -o hermes-web-studio ./cmd/hermes-web-studio)
  '';
  installPhase = ''
    install -Dm755 backend/hermes-web-studio $out/bin/hermes-web-studio
  '';
}
