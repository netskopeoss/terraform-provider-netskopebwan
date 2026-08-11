{ pkgs, ... }:

let
  goPackage = if pkgs ? go_1_26 then pkgs.go_1_26 else pkgs.go;

  # The two HashiCorp code generators are not in nixpkgs, so they are built from
  # source here. That keeps every tool the build needs pinned by this file and
  # devenv.lock rather than resolved over the network at shell entry, so the
  # shell and CI run the same versions.
  tfplugingen-openapi = pkgs.buildGoModule rec {
    pname = "tfplugingen-openapi";
    version = "0.3.0";

    src = pkgs.fetchFromGitHub {
      owner = "hashicorp";
      repo = "terraform-plugin-codegen-openapi";
      rev = "v${version}";
      hash = "sha256-6xI6PVlvYHwOnWjE0pKYDF/FvdomE5KydS7gBokJ2EM=";
    };

    vendorHash = "sha256-PAygKSZjHiszNrU02xOqyGn+aj7sj34M8QXTpqZh/zQ=";
    subPackages = [ "cmd/tfplugingen-openapi" ];

    meta.description = "Generates a provider code specification from an OpenAPI document";
  };

  tfplugingen-framework = pkgs.buildGoModule rec {
    pname = "tfplugingen-framework";
    version = "0.4.1";

    src = pkgs.fetchFromGitHub {
      owner = "hashicorp";
      repo = "terraform-plugin-codegen-framework";
      rev = "v${version}";
      hash = "sha256-a5eWS2pcr7tbAd9xrGJKRZ3DzHoBwM0FMLV5RGQhGa4=";
    };

    vendorHash = "sha256-pwdbpxOYQMxBuh4iq2M58YEn3tqyE1dn65yzsumJGhM=";
    subPackages = [ "cmd/tfplugingen-framework" ];

    meta.description = "Generates framework provider code from a provider code specification";
  };
in
{
  packages = [
    goPackage
    pkgs.terraform
    pkgs.curl
    pkgs.git
    pkgs.gnumake
    pkgs.golangci-lint
    pkgs.python3

    # The generator toolchain. mockgen and terraform-plugin-docs are packaged,
    # the two above are not.
    pkgs.mockgen
    pkgs.terraform-plugin-docs
    tfplugingen-openapi
    tfplugingen-framework
  ];

  env.GOTOOLCHAIN = "auto";

  # OPENAPI_SPEC_URL is deliberately not set here. A devenv env entry overrides
  # whatever the caller exported, which would make the variable look settable
  # while being ignored; the Makefile defaults it with ?= instead, so CI can
  # point a run at a different spec.

  enterShell = ''
    export TOOLS_BIN="$PWD/.tools/bin"
    export GOBIN="$TOOLS_BIN"
    export PATH="$TOOLS_BIN:$PATH"

    mkdir -p "$TOOLS_BIN"

    # tfgen is this repository's own code, so it is built rather than fetched.
    make tfgen

    echo "devenv: ready; run 'make help' to see project targets"
  '';
}
